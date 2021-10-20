// Copyright 2021 Authors of KubeArmor
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	securityv1 "github.com/kubearmor/KubeArmor/pkg/KubeArmorExternalWorkload/api/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
)

// KubeArmorExternalWorkloadReconciler reconciles a KubeArmorExternalWorkload object
type KubeArmorExternalWorkloadReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=security.kubearmor.com,resources=kubearmorexternalworkload,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=security.kubearmor.com,resources=kubearmorexternalworkloadstatus,verbs=get;update;patch

func (r *KubeArmorExternalWorkloadReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("kubearmorexternalworkload", req.NamespacedName)

	policy := &securityv1.KubeArmorExternalWorkload{}

	if err := r.Get(ctx, req.NamespacedName, policy); err != nil {
		log.Info("Invalid KubeArmorExternalWorkload")
		if errors.IsNotFound(err) {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		return ctrl.Result{}, err
	}

	// Validate KubeArmorExternalWorkload
	// if there are some issues in the CRD then delete it and return failure code
	//TODO: Handle the yaml spec validation

	log.Info("Fetched KubeArmorExternalWorkload")
	return ctrl.Result{}, nil

}

func (r *KubeArmorExternalWorkloadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&securityv1.KubeArmorExternalWorkload{}).
		Complete(r)
}
