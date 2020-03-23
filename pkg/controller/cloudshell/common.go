package cloudshell

import (
	"fmt"
	"github.com/che-incubator/cloudshell-operator/pkg/apis/cloudshell/v1alpha1"
	"github.com/google/uuid"
	"strings"
)

func getID(instance *v1alpha1.CloudShell) (string, error) {
	uid, err := uuid.Parse(string(instance.UID))
	if err != nil {
		return "", err
	}
	return strings.Join(strings.Split(uid.String(), "-")[0:3], ""), nil
}

func getLabelsForID(id string) map[string]string {
	return map[string]string{
		"cloudshell.id": id,
		"che.workspace_id": id,
	}
}

func getServiceAccountName(instance *v1alpha1.CloudShell) string {
	return fmt.Sprintf("cloudshell-%s", instance.Status.Id)
}

func getProxySecretName(instance *v1alpha1.CloudShell) string {
	return fmt.Sprintf("cloudshell-%s", instance.Status.Id)
}

func getRouteName(instance *v1alpha1.CloudShell) string {
	return fmt.Sprintf("cloudshell-%s", instance.Status.Id)
}

func getServiceName(instance *v1alpha1.CloudShell) string {
	return fmt.Sprintf("cloudshell-%s", instance.Status.Id)
}
