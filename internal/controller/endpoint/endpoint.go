package endpoint

import (
	"context"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	endpointsv1alpha1 "github.com/krateoplatformops/provider-argocd-endpoint/apis/endpoints/v1alpha1"
	"github.com/krateoplatformops/provider-argocd-endpoint/internal/clients"
	"github.com/krateoplatformops/provider-argocd-endpoint/internal/clients/accounts"

	corev1 "k8s.io/api/core/v1"
)

const (
	errNotEndpoint = "managed resource is not an argocd endpoint custom resource"
	//errGetPC          = "cannot get ProviderConfig"
	//errFmtKeyNotFound = "key %s is not found in referenced Kubernetes secret"
)

// Setup adds a controller that reconciles Token managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(endpointsv1alpha1.EndpointGroupKind)

	log := o.Logger.WithValues("controller", name)

	recorder := mgr.GetEventRecorderFor(name)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(endpointsv1alpha1.EndpointGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube: mgr.GetClient(),
			log:  log,
			rec:  recorder,
		}),
		managed.WithLogger(log),
		managed.WithRecorder(event.NewAPIRecorder(recorder)))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&endpointsv1alpha1.Endpoint{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connector struct {
	kube client.Client
	log  logging.Logger
	rec  record.EventRecorder
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*endpointsv1alpha1.Endpoint)
	if !ok {
		return nil, errors.New(errNotEndpoint)
	}

	cfg, err := clients.GetConfig(ctx, c.kube, cr)
	if err != nil {
		return nil, err
	}

	c.log.Debug("Created session", "token", cfg.AuthToken)

	return &external{
		kube: c.kube,
		log:  c.log,
		cfg:  cfg,
		rec:  c.rec,
	}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	kube client.Client
	log  logging.Logger
	cfg  *accounts.TokenProviderOptions
	rec  record.EventRecorder
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*endpointsv1alpha1.Endpoint)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotEndpoint)
	}

	spec := cr.Spec.ForProvider.DeepCopy()

	token, err := clients.GetEndpointSecret(ctx, e.kube, &spec.WriteSecretToRef)
	if err != nil {
		return managed.ExternalObservation{}, err
	}

	if len(token) > 0 {
		cr.SetConditions(xpv1.Available())

		// TODO handle token expiration?
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	return managed.ExternalObservation{
		ResourceExists:   false,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*endpointsv1alpha1.Endpoint)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotEndpoint)
	}

	cr.SetConditions(xpv1.Creating())

	spec := cr.Spec.ForProvider.DeepCopy()

	token, err := accounts.GenerateToken(e.cfg, spec.Account, 0)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	e.log.Debug("Generated argocd token", "account", spec.Account)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "TokenCreated", "Generated argocd token for account: %s", spec.Account)

	err = clients.CreateEndpointSecret(ctx, e.kube, clients.CreateSecretOpts{
		Token:     token,
		TargetURL: e.cfg.ServerUrl,
		SecretRef: &spec.WriteSecretToRef,
	})
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	e.log.Debug("Saved argocd token as secret", "account", spec.Account, "secret", spec.WriteSecretToRef.Name)
	e.rec.Eventf(cr, corev1.EventTypeNormal, "TokenSaved", "Saved argocd token for account '%s' into '%s' secret", spec.Account, spec.WriteSecretToRef.Name)

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil // noop
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*endpointsv1alpha1.Endpoint)
	if !ok {
		return errors.New(errNotEndpoint)
	}

	cr.SetConditions(xpv1.Deleting())

	spec := cr.Spec.ForProvider.DeepCopy()

	e.log.Debug("Deleting argocd token secret", "account", spec.Account, "secret", spec.WriteSecretToRef.Name)

	err := clients.DeleteEndpointSecret(ctx, e.kube, &spec.WriteSecretToRef)
	if err == nil {
		e.rec.Eventf(cr, corev1.EventTypeNormal, "TokenDeleted", "Deleted argocd token for account '%s' into '%s' secret", spec.Account, spec.WriteSecretToRef.Name)
	}

	return err
}
