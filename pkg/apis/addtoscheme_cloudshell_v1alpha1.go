package apis

import (
	"github.com/che-incubator/cloudshell-operator/pkg/apis/cloudshell/v1alpha1"
	routeV1 "github.com/openshift/api/route/v1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes, v1alpha1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes,
		routeV1.AddToScheme,
	)
}
