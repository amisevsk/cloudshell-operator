// +build !ignore_autogenerated

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"./pkg/apis/cloudshell/v1alpha1.CloudShell":       schema_pkg_apis_cloudshell_v1alpha1_CloudShell(ref),
		"./pkg/apis/cloudshell/v1alpha1.CloudShellSpec":   schema_pkg_apis_cloudshell_v1alpha1_CloudShellSpec(ref),
		"./pkg/apis/cloudshell/v1alpha1.CloudShellStatus": schema_pkg_apis_cloudshell_v1alpha1_CloudShellStatus(ref),
	}
}

func schema_pkg_apis_cloudshell_v1alpha1_CloudShell(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CloudShell is the Schema for the cloudshells API",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/cloudshell/v1alpha1.CloudShellSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("./pkg/apis/cloudshell/v1alpha1.CloudShellStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"./pkg/apis/cloudshell/v1alpha1.CloudShellSpec", "./pkg/apis/cloudshell/v1alpha1.CloudShellStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_cloudshell_v1alpha1_CloudShellSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CloudShellSpec defines the desired state of CloudShell",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"image": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
				Required: []string{"image"},
			},
		},
	}
}

func schema_pkg_apis_cloudshell_v1alpha1_CloudShellStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CloudShellStatus defines the observed state of CloudShell",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"id": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"ready": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"url": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
				Required: []string{"id", "ready", "url"},
			},
		},
	}
}
