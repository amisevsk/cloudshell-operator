package controller

import (
	"github.com/che-incubator/cloudshell-operator/pkg/controller/cloudshell"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, cloudshell.Add)
}
