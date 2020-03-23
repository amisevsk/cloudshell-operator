package cloudshell

import (
	"context"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)
import rbacv1 "k8s.io/api/rbac/v1"

func (r *ReconcileCloudShell) reconcilePrereqs(ctx reconcileContext) deployStatus {
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cloudshell-exec",
			Namespace: ctx.instance.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				Resources: []string{"pods/exec"},
				APIGroups: []string{""},
				Verbs:     []string{"create"},
			},
		},
	}
	controllerutil.SetControllerReference(ctx.instance, role, r.scheme)
	err := r.client.Create(context.TODO(), role)
	if err != nil && !errors.IsAlreadyExists(err) {
		return deployStatus{Error: err}
	}

	bindings := []*rbacv1.RoleBinding{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cloudshell-view",
				Namespace: ctx.instance.Namespace,
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "ClusterRole",
				Name: "view",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind: "Group",
					Name: "system:serviceaccounts:" + ctx.instance.Namespace,
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cloudshell-exec",
				Namespace: ctx.instance.Namespace,
			},
			RoleRef: rbacv1.RoleRef{
				Kind: "Role",
				Name: "cloudshell-exec",
			},
			Subjects: []rbacv1.Subject{
				{
					Kind: "Group",
					Name: "system:serviceaccounts:" + ctx.instance.Namespace,
				},
			},
		},
	}
	for _, binding := range bindings {
		controllerutil.SetControllerReference(ctx.instance, binding, r.scheme)
		err := r.client.Create(context.TODO(), binding)
		if err != nil && !errors.IsAlreadyExists(err) {
			return deployStatus{Error: err}
		}
	}

	return deployStatus{Continue: true}
}
