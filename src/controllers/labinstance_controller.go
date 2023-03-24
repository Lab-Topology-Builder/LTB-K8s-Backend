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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ltbbackendv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/src/api/v1alpha1"
)

// LabInstanceReconciler reconciles a LabInstance object
type LabInstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labinstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labinstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labinstances/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the LabInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
// TODO: check if lab template exists, if not, throw an error
// TODO: Check if deployment already exists, if not create a new one, or update an existing one
// TODO: Check how to make the reference to the lab template, maybe try to make use of the context to get the lab template
func (r *LabInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	labInstance := &ltbbackendv1alpha1.LabInstance{}
	err := r.Get(ctx, req.NamespacedName, labInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("LabInstance resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get LabInstance")
		return ctrl.Result{}, err
	}

	found := &appsv1.Deployment{}

	err = r.Get(ctx, types.NamespacedName{Name: labInstance.Name, Namespace: labInstance.Namespace}, found)

	labTemplate := &ltbbackendv1alpha1.LabTemplate{}
	err = r.Get(ctx, types.NamespacedName{Name: labInstance.Spec.LabTemplateReference, Namespace: labInstance.Namespace}, labTemplate)

	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		dep := r.deploymentForLabInstance(labInstance, labTemplate)
		log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		err = r.Create(ctx, dep)
		if err != nil {
			log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) deploymentForLabInstance(labInstance *ltbbackendv1alpha1.LabInstance, labTemplate *ltbbackendv1alpha1.LabTemplate) *appsv1.Deployment {
	ls := labelsForLabInstance(labInstance.Name)

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labInstance.Name,
			Namespace: labInstance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
					Name:   labInstance.Spec.LabTemplateReference,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image: "ubuntu:latest",
						Name:  labInstance.Spec.LabInstanceName,
						Ports: []corev1.ContainerPort{{ContainerPort: 8080}},
					}},
				},
			},
		},
	}
	// Set LabInstance instance as the owner and controller
	ctrl.SetControllerReference(labInstance, dep, r.Scheme)
	return dep
}

func labelsForLabInstance(name string) map[string]string {
	return map[string]string{"app": "labinstance", "labinstance_cr": name}
}

// SetupWithManager sets up the controller with the Manager.
func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbbackendv1alpha1.LabInstance{}).
		Complete(r)
}
