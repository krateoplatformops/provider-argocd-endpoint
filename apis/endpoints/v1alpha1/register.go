package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "argocd.krateo.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}
	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Task type metadata
var (
	EndpointKind             = reflect.TypeOf(Endpoint{}).Name()
	EndpointGroupKind        = schema.GroupKind{Group: Group, Kind: EndpointKind}.String()
	EndpointKindAPIVersion   = EndpointKind + "." + SchemeGroupVersion.String()
	EndpointGroupVersionKind = SchemeGroupVersion.WithKind(EndpointKind)
)

func init() {
	SchemeBuilder.Register(&Endpoint{}, &EndpointList{})
}
