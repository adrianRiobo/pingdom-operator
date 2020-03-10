package controller

import (
	"github.com/adrianRiobo/pingdom-operator/pkg/controller/pingdomcheck"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, pingdomcheck.Add)
}
