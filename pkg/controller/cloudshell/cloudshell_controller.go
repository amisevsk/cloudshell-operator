package cloudshell

import (
	"context"
	"github.com/go-logr/logr"
	routeV1 "github.com/openshift/api/route/v1"

	cloudshellv1alpha1 "github.com/che-incubator/cloudshell-operator/pkg/apis/cloudshell/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

type deployStatus struct {
	Continue bool
	Requeue  bool
	Error    error
}

type reconcileContext struct {
	instance *cloudshellv1alpha1.CloudShell
	log      logr.Logger
}

var log = logf.Log.WithName("controller_cloudshell")

// Add creates a new CloudShell Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCloudShell{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("cloudshell-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CloudShell
	err = c.Watch(&source.Kind{Type: &cloudshellv1alpha1.CloudShell{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resources
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cloudshellv1alpha1.CloudShell{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &routeV1.Route{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cloudshellv1alpha1.CloudShell{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cloudshellv1alpha1.CloudShell{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.ServiceAccount{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &cloudshellv1alpha1.CloudShell{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCloudShell implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCloudShell{}

// ReconcileCloudShell reconciles a CloudShell object
type ReconcileCloudShell struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileCloudShell) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling CloudShell")

	// Fetch the CloudShell instance
	instance := &cloudshellv1alpha1.CloudShell{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	ctx := reconcileContext{
		instance: instance,
		log:      reqLogger,
	}

	if instance.Status.Id == "" {
		id, err := getID(instance)
		if err != nil {
			// TODO: Fail
			return reconcile.Result{}, err
		}
		instance.Status.Id = id
		err = r.client.Status().Update(context.TODO(), instance)
		return reconcile.Result{Requeue: true}, err
	}

	networkStatus := r.reconcileRouting(ctx)
	if !networkStatus.Continue {
		return reconcile.Result{Requeue: networkStatus.Requeue}, networkStatus.Error
	}

	deploymentStatus := r.reconcileDeployment(ctx)
	if !deploymentStatus.Continue {
		return reconcile.Result{Requeue: deploymentStatus.Requeue}, deploymentStatus.Error
	}

	return reconcile.Result{}, nil
}
