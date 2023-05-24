/*
Copyright 2023 Jan Untersander, Tsigereda Nebai Kidane.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ltbbackendv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	"github.com/go-logr/logr"
)

type LabInstanceSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labinstancesets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labinstancesets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labinstancesets/finalizers,verbs=update

func (r *LabInstanceSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := r.Log.WithValues("labinstanceset", req.NamespacedName)
	log.Info("Managing resources.")
	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LabInstanceSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbbackendv1alpha1.LabInstanceSet{}).
		Complete(r)
}
