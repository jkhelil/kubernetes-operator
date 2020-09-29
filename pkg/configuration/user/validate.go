package user

import (
	"github.com/jenkinsci/kubernetes-operator/pkg/apis/jenkins/v1alpha2"
	"github.com/jenkinsci/kubernetes-operator/pkg/configuration/backuprestore"
	"github.com/jenkinsci/kubernetes-operator/pkg/configuration/user/seedjobs"
)

// Validate validates Jenkins CR Spec section
func (r *reconcileUserConfiguration) Validate(jenkins *v1alpha2.Jenkins) ([]string, error) {
	backupAndRestore := backuprestore.New(r.Configuration, r.logger)
	msg := backupAndRestore.Validate()
	return msg, nil
}
