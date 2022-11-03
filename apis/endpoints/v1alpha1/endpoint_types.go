package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// EndpointObservation are the observable fields of a Endpoint.
type EndpointObservation struct {
	ID        string `json:"id,omitempty"`
	ExpiresIn string `json:"expiresIn,omitempty"`
}

// EndpointParameters are the configurable fields of an Endpoint.
type EndpointParameters struct {
	// ID optional endpoint id. Fall back to uuid if not value specified
	// +optional
	ID string `json:"id,omitempty"`

	// Account name
	Account string `json:"account"`

	// ExpiresIn duration before the token will expire. (Default: No expiration)
	// +optional
	// ExpiresIn string `json:"expiresIn,omitempty"`

	WriteSecretToRef xpv1.SecretReference `json:"writeSecretToRef"`
}

// A EndpointSpec defines the desired state of an Endpoint.
type EndpointSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       EndpointParameters `json:"forProvider"`
}

// A EndpointStatus represents the observed state of an Endpoint.
type EndpointStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          EndpointObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,krateo,argocd}
// +kubebuilder:subresource:status
type Endpoint struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EndpointSpec   `json:"spec"`
	Status EndpointStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// EndpointList contains a list of Endpoint
type EndpointList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Endpoint `json:"items"`
}
