// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorVirtualMachine/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

// KubeArmorVirtualMachineReconciler reconciles a KubeArmorVirtualMachine object
type KubeArmorVirtualMachineReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=security.kubearmor.com,resources=kubearmorvirtualmachine,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.kubearmor.com,resources=kubearmorvirtualmachinestatus,verbs=get;update;patch

func (r *KubeArmorVirtualMachineReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kubearmorvirtualmachine", req.NamespacedName)

	policy := &securityv1.KubeArmorVirtualMachine{}

	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		log.Info("Invalid KubeArmorVirtualMachine")
		if errors.IsNotFound(err) {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return ctrl.Result{}, err
	}

	// Validate KubeArmorVirtualMachine
	// if there are some issues in the CRD then delete it and return failure code
	//TODO: Handle the yaml spec validation

	log.Info("Fetched KubeArmorVirtualMachine")
	return ctrl.Result{}, nil

}

func (r *KubeArmorVirtualMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1.KubeArmorVirtualMachine{}).
		Complete(r)
}
