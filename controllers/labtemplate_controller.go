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
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	util "github.com/Lab-Topology-Builder/LTB-K8s-Backend/util"
)

type LabTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labtemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labtemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labtemplates/finalizers,verbs=update
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=nodetypes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=nodetypes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=nodetypes/finalizers,verbs=update

func (r *LabTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	labTemplate := &ltbv1alpha1.LabTemplate{}
	l.Info("Reconciling LabTemplate")
	err := r.Get(ctx, req.NamespacedName, labTemplate)
	if err != nil {
		l.Error(err, "Failed to get labtemplate, ignoring must have been deleted")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	nodes := &labTemplate.Spec.Nodes
	for i := 0; i < len(*nodes); i++ {
		nodetype := &ltbv1alpha1.NodeType{}
		err := r.Get(ctx, client.ObjectKey{Namespace: labTemplate.Namespace, Name: (*nodes)[i].NodeTypeRef.Type}, nodetype)
		if err != nil {
			l.Error(err, "Failed to get nodetype")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		var renderedNodeSpec strings.Builder
		if err = util.ParseAndRenderTemplate(nodetype, &renderedNodeSpec, (*nodes)[i]); err != nil {
			l.Error(err, "Failed to render template")
			return ctrl.Result{}, err
		}
		(*nodes)[i].RenderedNodeSpec = renderedNodeSpec.String()
	}

	err = r.Update(ctx, labTemplate)
	if err != nil {
		l.Error(err, "Failed to update labtemplate")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LabTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbv1alpha1.LabTemplate{}).
		Complete(r)
}
