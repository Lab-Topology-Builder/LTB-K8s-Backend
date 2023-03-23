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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ltbbackendv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/src/api/v1alpha1"
)

// LabTemplateReconciler reconciles a LabTemplate object
type LabTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labtemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labtemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labtemplates/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LabTemplate object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *LabTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	labTemplate := &ltbbackendv1alpha1.LabTemplate{}
	err := r.Get(ctx, req.NamespacedName, labTemplate)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("LabTemplate not found")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Unable to get LabTemplate")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *LabTemplateReconciler) deploymentForLabTemplate(labTemplate *ltbbackendv1alpha1.LabTemplate) *appsv1.Deployment {
	ls := labelsForLabTemplate(labTemplate.Name)
	hosts := labTemplate.Spec.BasicTemplate.Spec.Hosts
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labTemplate.Name,
			Namespace: labTemplate.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:    labTemplate.Labels,
					Name:      labTemplate.Name,
					Namespace: labTemplate.Namespace,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name: "MPLS",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 8080,
							Name:          "labtemplate",
						}},
					}},
				},
			},
		},
	}
	// Set LabTemplate instance as the owner and controller
	ctrl.SetControllerReference(labTemplate, dep, r.Scheme)
	return dep
}

func labelsForLabTemplate(name string) map[string]string {
	return map[string]string{"app": "labtemplate", "labtemplate_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *LabTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbbackendv1alpha1.LabTemplate{}).
		Complete(r)
}
