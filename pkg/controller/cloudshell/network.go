package cloudshell

import (
	"context"
	"fmt"
	"github.com/che-incubator/cloudshell-operator/pkg/apis/cloudshell/v1alpha1"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	routeV1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

var serviceDiffOpts = cmp.Options{
	cmpopts.IgnoreFields(corev1.Service{}, "TypeMeta", "ObjectMeta", "Status"),
	cmpopts.IgnoreFields(corev1.ServiceSpec{}, "ClusterIP", "SessionAffinity"),
	cmpopts.IgnoreFields(corev1.ServicePort{}, "TargetPort"),
	cmpopts.SortSlices(func(a, b corev1.ServicePort) bool {
		return strings.Compare(a.Name, b.Name) > 0
	}),
}

var routeDiffOpts = cmp.Options{
	cmpopts.IgnoreFields(routeV1.Route{}, "TypeMeta", "ObjectMeta", "Status"),
	cmpopts.IgnoreFields(routeV1.RouteSpec{}, "WildcardPolicy"),
	cmpopts.IgnoreFields(routeV1.RouteTargetReference{}, "Weight"),
}

func (r *ReconcileCloudShell) reconcileRouting(ctx reconcileContext) deployStatus {
	specService, specRoute := r.getSpecRouting(ctx.instance)

	serviceOk, err := r.reconcileService(specService, ctx.log)
	if err != nil || !serviceOk {
		return deployStatus{
			Requeue: true,
			Error:   err,
		}
	}

	routeOk, err := r.reconcileRoute(specRoute, ctx.log)
	if err != nil || !routeOk {
		return deployStatus{
			Requeue: true,
			Error:   err,
		}
	}

	return deployStatus{
		Continue: true,
	}
}

func (r *ReconcileCloudShell) getSpecRouting(instance *v1alpha1.CloudShell) (*corev1.Service, *routeV1.Route) {
	id := instance.Status.Id
	labels := getLabelsForID(id)
	service := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Name:      getServiceName(instance),
			Namespace: instance.Namespace, // TODO: Make configurable
			Labels:    labels,
			Annotations: map[string]string{
				"service.alpha.openshift.io/serving-cert-secret-name": getServiceAccountName(instance),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "cloud-shell-proxy",
					Protocol:   corev1.ProtocolTCP,
					Port:       8443,
					TargetPort: intstr.FromInt(8443),
				},
			},
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	route := &routeV1.Route{
		ObjectMeta: v1.ObjectMeta{
			Name:      getRouteName(instance),
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: routeV1.RouteSpec{
			Host: fmt.Sprintf("%s-%s.%s", "cloudshell", instance.Status.Id, "192.168.42.191.nip.io"),
			To: routeV1.RouteTargetReference{
				Kind: "Service",
				Name: service.Name,
			},
			TLS: &routeV1.TLSConfig{
				Termination:                   routeV1.TLSTerminationReencrypt,
				InsecureEdgeTerminationPolicy: routeV1.InsecureEdgeTerminationPolicyRedirect,
			},
		},
	}

	controllerutil.SetControllerReference(instance, service, r.scheme)
	controllerutil.SetControllerReference(instance, route, r.scheme)

	return service, route
}

// The remaining functions are a good argument for type parameters/generics in Go.
func (r *ReconcileCloudShell) reconcileService(spec *corev1.Service, log logr.Logger) (ok bool, err error) {
	cluster, err := r.getClusterService(spec)
	if err != nil {
		return false, err
	}
	if cluster == nil {
		log.Info("Creating service")
		err = r.client.Create(context.TODO(), spec)
		return false, err
	}
	if !cmp.Equal(spec, cluster, serviceDiffOpts) {
		log.Info("Patching service")
		patch := client.MergeFrom(spec)
		err = r.client.Patch(context.TODO(), cluster, patch)
		return false, err
	}
	return true, nil
}

func (r *ReconcileCloudShell) reconcileRoute(spec *routeV1.Route, log logr.Logger) (ok bool, err error) {
	cluster, err := r.getClusterRoute(spec)
	if err != nil {
		return false, err
	}
	if cluster == nil {
		log.Info("Creating route")
		err = r.client.Create(context.TODO(), spec)
		if errors.IsAlreadyExists(err) {
			// TODO: Investigate this more, always happens once, maybe due to time it takes for route to be created.
			log.Info("Route already exists")
			return false, nil
		}
		return false, err
	}
	if !cmp.Equal(spec, cluster, routeDiffOpts) {
		log.Info("Patching route")
		patch := client.MergeFrom(spec)
		err = r.client.Patch(context.TODO(), cluster, patch)
		return false, err
	}
	return true, nil
}

func (r *ReconcileCloudShell) getClusterService(spec *corev1.Service) (*corev1.Service, error) {
	found := &corev1.Service{}
	namespaceName := types.NamespacedName{
		Name:      spec.Name,
		Namespace: spec.Namespace,
	}
	err := r.client.Get(context.TODO(), namespaceName, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return found, err
}

func (r *ReconcileCloudShell) getClusterRoute(spec *routeV1.Route) (*routeV1.Route, error) {
	found := &routeV1.Route{}
	namespaceName := types.NamespacedName{
		Name:      spec.Name,
		Namespace: spec.Namespace,
	}
	err := r.client.Get(context.TODO(), namespaceName, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return found, err
}
