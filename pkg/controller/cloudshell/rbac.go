package cloudshell

import (
	"context"
	"fmt"
	"github.com/che-incubator/cloudshell-operator/pkg/apis/cloudshell/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const proxyServiceAcctAnnotationKeyFmt string = "serviceaccounts.openshift.io/oauth-redirectreference.%s"
const proxyServiceAcctAnnotationValueFmt string = `{"kind":"OAuthRedirectReference","apiVersion":"v1","reference":{"kind":"Route","name":"%s"}}`

func (r *ReconcileCloudShell) reconcileServiceAcct(ctx reconcileContext) deployStatus {
	spec, err := r.getSpecSA(ctx.instance)
	if err != nil {
		return deployStatus{Error: err}
	}
	cluster, err := r.getClusterSA(spec)
	if err != nil {
		return deployStatus{Error: err}
	}
	if cluster == nil {
		ctx.log.Info("Creating service account")
		err = r.client.Create(context.TODO(), spec)
		return deployStatus{Requeue: true, Error: err}
	}
	redirectAnnotation := fmt.Sprintf(proxyServiceAcctAnnotationKeyFmt, ctx.instance.Status.Id)
	val, ok := cluster.Annotations[redirectAnnotation]
	if !ok || val != spec.Annotations[redirectAnnotation] {
		ctx.log.Info("Updating service account")
		patch := client.MergeFrom(spec)
		err = r.client.Patch(context.TODO(), cluster, patch)
		return deployStatus{Requeue: true, Error: err}
	}

	return deployStatus{
		Continue: true,
	}
}

func (r *ReconcileCloudShell) getSpecSA(instance *v1alpha1.CloudShell) (*corev1.ServiceAccount, error) {
	autoMountServiceAccount := true
	annotations := map[string]string{
		fmt.Sprintf(proxyServiceAcctAnnotationKeyFmt, instance.Status.Id): fmt.Sprintf(proxyServiceAcctAnnotationValueFmt, getRouteName(instance)),
	}

	sa := &corev1.ServiceAccount{
		ObjectMeta: v1.ObjectMeta{
			Name:        getServiceAccountName(instance),
			Namespace:   instance.Namespace,
			Annotations: annotations,
		},
		AutomountServiceAccountToken: &autoMountServiceAccount,
	}

	err := controllerutil.SetControllerReference(instance, sa, r.scheme)
	if err != nil {
		return nil, err
	}
	return sa, nil
}

func (r *ReconcileCloudShell) getClusterSA(spec *corev1.ServiceAccount) (*corev1.ServiceAccount, error) {
	sa := &corev1.ServiceAccount{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      spec.Name,
		Namespace: spec.Namespace,
	}, sa)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return sa, err
}
