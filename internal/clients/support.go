package clients

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

const (
	tokenKey = "bearer"
)

type CreateSecretOpts struct {
	Token     string
	TargetURL string
	SecretRef *xpv1.SecretReference
}

func CreateEndpointSecret(ctx context.Context, k client.Client, opts CreateSecretOpts) error {
	if opts.SecretRef == nil {
		return errors.New("no endpoint secret referenced")
	}

	s := &corev1.Secret{}
	s.Name = opts.SecretRef.Name
	s.Namespace = opts.SecretRef.Namespace
	s.Labels = map[string]string{
		"app.kubernetes.io/created-by": "krateo",
		"category":                     "delivery",
		"group":                        "endpoint",
		"icon":                         "fa-solid_fa-truck",
		"type":                         "argocd",
	}
	s.StringData = map[string]string{
		tokenKey: opts.Token,
		"target": opts.TargetURL,
	}

	return k.Create(ctx, s)
}

func GetEndpointSecret(ctx context.Context, k client.Client, ref *xpv1.SecretReference) (string, error) {
	if ref == nil {
		return "", errors.New("no credentials secret referenced")
	}

	s := &corev1.Secret{}
	err := k.Get(ctx, types.NamespacedName{Namespace: ref.Namespace, Name: ref.Name}, s)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return "", nil
		}
		return "", errors.Wrapf(err, "cannot get %s secret in namespace %s", ref.Name, ref.Namespace)
	}

	return string(s.Data[tokenKey]), nil
}

func DeleteEndpointSecret(ctx context.Context, k client.Client, ref *xpv1.SecretReference) error {
	if ref == nil {
		return errors.New("no endpoint secret referenced")
	}

	s := &corev1.Secret{}
	s.Name = ref.Name
	s.Namespace = ref.Namespace

	return k.Delete(ctx, s)
}

/*
func ErrorIsNotFound(err error) bool {
	ex, ok := err.(*apierrors.StatusError)
	return ok || (ex.Status().Code == http.StatusNotFound)
}
*/

// IsBoolPtrEqualToBool compares a *bool with bool
func IsBoolPtrEqualToBool(bp *bool, b bool) bool {
	if bp == nil {
		return false
	}

	return (*bp == b)
}
