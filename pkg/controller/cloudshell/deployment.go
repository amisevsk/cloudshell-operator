package cloudshell

import (
	"context"
	"fmt"
	"github.com/che-incubator/cloudshell-operator/pkg/apis/cloudshell/v1alpha1"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

const openShiftProxySARFmt = `{"namespace": "%s", "resource": "pods", "name": "%s", "verb": "exec"}`

var deploymentDiffOpts = cmp.Options{
	cmpopts.IgnoreFields(appsv1.Deployment{}, "TypeMeta", "ObjectMeta", "Status"),
	cmpopts.IgnoreFields(appsv1.DeploymentSpec{}, "RevisionHistoryLimit", "ProgressDeadlineSeconds"),
	cmpopts.IgnoreFields(corev1.PodSpec{}, "DNSPolicy", "SchedulerName", "DeprecatedServiceAccount", "RestartPolicy", "SecurityContext"),
	cmpopts.IgnoreFields(corev1.Container{}, "TerminationMessagePath", "TerminationMessagePolicy", "ImagePullPolicy"),
	cmpopts.SortSlices(func(a, b corev1.Container) bool {
		return strings.Compare(a.Name, b.Name) > 0
	}),
	cmpopts.SortSlices(func(a, b corev1.Volume) bool {
		return strings.Compare(a.Name, b.Name) > 0
	}),
}

func (r *ReconcileCloudShell) reconcileDeployment(ctx reconcileContext) deployStatus {
	spec, err := r.getSpecDeployment(ctx.instance)
	if err != nil {
		return deployStatus{Error: err}
	}

	cluster, err := r.getClusterDeployment(spec)
	if err != nil {
		return deployStatus{Error: err}
	}
	if cluster == nil {
		ctx.log.Info("Creating deployment")
		err = r.client.Create(context.TODO(), spec)
		if errors.IsAlreadyExists(err) {
			return deployStatus{Requeue: true}
		}
		return deployStatus{Requeue: true, Error: err}
	}
	if !cmp.Equal(spec, cluster, deploymentDiffOpts) {
		ctx.log.Info("Patching deployment")
		patch := client.MergeFrom(spec)
		err = r.client.Patch(context.TODO(), cluster, patch)
		if errors.IsConflict(err) {
			// Modified since we started, requeue
			return deployStatus{Requeue: true}
		}
		return deployStatus{Requeue: true, Error: err}
	}

	if !deploymentReady(cluster) {
		ctx.log.Info("Deployment not ready")
		return deployStatus{}
	}

	return deployStatus{
		Continue: true,
	}
}

func deploymentReady(deployment *appsv1.Deployment) (ready bool) {
	// TODO: available doesn't mean what you might think
	for _, condition := range deployment.Status.Conditions {
		if condition.Type != appsv1.DeploymentAvailable || condition.Status != corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func (r *ReconcileCloudShell) getClusterDeployment(spec *appsv1.Deployment) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	namespacedName := types.NamespacedName{
		Namespace: spec.Namespace,
		Name:      spec.Name,
	}
	err := r.client.Get(context.TODO(), namespacedName, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return deployment, nil
}

func (r *ReconcileCloudShell) getSpecDeployment(instance *v1alpha1.CloudShell) (*appsv1.Deployment, error) {
	id := instance.Status.Id
	labels := getLabelsForID(id)
	replicas := int32(1)
	rollingUpdateParam := intstr.FromInt(1)

	deployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      "cloudshell-" + id,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Name:      instance.Status.Id,
					Namespace: instance.Namespace,
					Labels:    labels,
				},
				Spec: getSpecPod(instance),
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: "RollingUpdate",
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxSurge:       &rollingUpdateParam,
					MaxUnavailable: &rollingUpdateParam,
				},
			},
		},
	}
	err := controllerutil.SetControllerReference(instance, deployment, r.scheme)
	return deployment, err
}

func getSpecPod(instance *v1alpha1.CloudShell) corev1.PodSpec {
	terminationGracePeriod := int64(1)
	var volumeDefaultMode int32 = 420
	proxySecretName := getProxySecretName(instance)
	resources := corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
	}

	return corev1.PodSpec{
		Volumes: []corev1.Volume{
			{
				Name: proxySecretName,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName:  proxySecretName,
						DefaultMode: &volumeDefaultMode,
					},
				},
			},
		},
		Containers: []corev1.Container{
			{
				Name:                     "shell-host",
				Image:                    instance.Spec.Image,
				ImagePullPolicy:          corev1.PullAlways,
				TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
				Resources:                resources,
			},
			{
				Name:  "machine-exec",
				Image: "quay.io/eclipse/che-machine-exec:nightly",
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 4444,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				Resources:                resources,
				ImagePullPolicy:          corev1.PullAlways,
				TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
			},
			{
				Name:  "oauth-proxy",
				Image: "openshift/oauth-proxy:latest",
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 8443,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      proxySecretName,
						MountPath: "/etc/tls/private",
					},
				},
				ImagePullPolicy:          corev1.PullAlways,
				TerminationMessagePolicy: corev1.TerminationMessageFallbackToLogsOnError,
				Resources:                resources,
				Args: []string{
					"--https-address=:8443",
					"--http-address=127.0.0.1:8080",
					"--provider=openshift",
					// TODO:
					"--openshift-service-account=" + getServiceAccountName(instance),
					"--upstream=http://localhost:4444",
					"--tls-cert=/etc/tls/private/tls.crt",
					"--tls-key=/etc/tls/private/tls.key",
					"--cookie-secret=SECRET_TODO", // TODO
					// Currently: block anyone who can't exec in the current namespace
					"--openshift-sar=" + fmt.Sprintf(openShiftProxySARFmt, "", ""),
				},
			},
		},
		TerminationGracePeriodSeconds: &terminationGracePeriod,
		ServiceAccountName:            getServiceAccountName(instance),
	}
}
