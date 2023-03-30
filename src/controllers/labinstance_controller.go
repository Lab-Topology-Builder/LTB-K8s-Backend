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
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	kubevirtv1 "kubevirt.io/api/core/v1"

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
			log.Info("LabInstance resource not found.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get LabInstance")
		return ctrl.Result{}, err
	}

	labTemplate := &ltbbackendv1alpha1.LabTemplate{}
	err = r.Get(ctx, types.NamespacedName{Name: labInstance.Spec.LabTemplateReference, Namespace: labInstance.Namespace}, labTemplate)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("LabTemplate resource not found.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get LabTemplate")
		return ctrl.Result{}, err
	}
	log.Info("LabTemplate resource found.", "LabTemplate.Namespace", labTemplate.Namespace, "LabTemplate.Name", labTemplate.Name)

	foundPod := &corev1.Pod{}
	err = r.Get(ctx, types.NamespacedName{Name: labInstance.Name, Namespace: labInstance.Namespace}, foundPod)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Pod
		pod := r.mapTemplateToPod(labInstance, labTemplate)
		log.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.Create(ctx, pod)
		if err != nil {
			log.Error(err, "Failed to create new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			return ctrl.Result{}, err
		}
		// Pod created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Pod")
		return ctrl.Result{}, err
	}

	// Check status of the pod
	if err := r.checkPodStatus(ctx, foundPod); err != nil {
		log.Error(err, "Failed to check Pod status")
		return ctrl.Result{}, err
	}

	foundVM := &kubevirtv1.VirtualMachine{}
	err = r.Get(ctx, types.NamespacedName{Name: labInstance.Name, Namespace: labInstance.Namespace}, foundVM)
	if err != nil && errors.IsNotFound(err) {
		// Define a new VM
		vm := r.mapTemplateToVM(labInstance, labTemplate)
		log.Info("Creating a new VM", "VM.Namespace", vm.Namespace, "VM.Name", vm.Name)
		err = r.Create(ctx, vm)
		if err != nil {
			log.Error(err, "Failed to create new VM", "VM.Namespace", vm.Namespace, "VM.Name", vm.Name)
			return ctrl.Result{}, err
		}
		// VM created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Pod")
		return ctrl.Result{}, err
	}

	// Check status of the VM
	if err := r.checkVMStatus(ctx, foundVM); err != nil {
		log.Error(err, "Failed to check VM status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) mapTemplateToPod(labInstance *ltbbackendv1alpha1.LabInstance, labTemplate *ltbbackendv1alpha1.LabTemplate) *corev1.Pod {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labInstance.Name,
			Namespace: labInstance.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    labTemplate.Spec.Nodes[0].Name,
					Image:   labTemplate.Spec.Nodes[0].Image.Type + ":" + labTemplate.Spec.Nodes[0].Image.Version,
					Command: []string{"/bin/sleep", "365d"},
				},
			},
		},
		Status: labInstance.Status.Phase,
	}

	ctrl.SetControllerReference(labInstance, pod, r.Scheme)
	return pod
}

func (r *LabInstanceReconciler) mapTemplateToVM(labInstance *ltbbackendv1alpha1.LabInstance, labTemplate *ltbbackendv1alpha1.LabTemplate) *kubevirtv1.VirtualMachine {
	bool := true
	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labInstance.Name,
			Namespace: labInstance.Namespace,
		},
		Spec: kubevirtv1.VirtualMachineSpec{
			Running: &bool,
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Resources: kubevirtv1.ResourceRequirements{
							Requests: corev1.ResourceList{"memory": resource.MustParse("2048M")},
						},
						CPU: &kubevirtv1.CPU{Cores: 1},
						Devices: kubevirtv1.Devices{
							Disks: []kubevirtv1.Disk{
								{Name: "containerdisk", DiskDevice: kubevirtv1.DiskDevice{Disk: &kubevirtv1.DiskTarget{Bus: "virtio"}}},
								{Name: "cloudinitdisk", DiskDevice: kubevirtv1.DiskDevice{Disk: &kubevirtv1.DiskTarget{Bus: "virtio"}}},
							},
						},
					},
					Volumes: []kubevirtv1.Volume{
						{Name: "containerdisk", VolumeSource: kubevirtv1.VolumeSource{ContainerDisk: &kubevirtv1.ContainerDiskSource{Image: "quay.io/containerdisks/" + labTemplate.Spec.Nodes[0].Image.Type + ":" + labTemplate.Spec.Nodes[0].Image.Version}}},
						{Name: "cloudinitdisk", VolumeSource: kubevirtv1.VolumeSource{CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{UserData: labTemplate.Spec.Nodes[0].Config}}}},
				},
			},
		},
	}
	return vm
}

func (r *LabInstanceReconciler) checkPodStatus(ctx context.Context, pod *corev1.Pod) error {
	for {
		phase := pod.Status.Phase
		fmt.Printf("Pod status: %v\n", phase)
		if phase == corev1.PodRunning {
			return nil
		} else if phase == corev1.PodFailed || phase == corev1.PodUnknown {
			return fmt.Errorf("pod %s in %s is in %v state", pod.Name, pod.Namespace, phase)
		} else {
			fmt.Printf("pod %s still starting, waiting 5 seconds...\n", pod.Name)
			time.Sleep(5 * time.Second)
			err := r.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pod)
			if err != nil {
				return err
			}

		}
	}
}

func (r *LabInstanceReconciler) checkVMStatus(ctx context.Context, vm *kubevirtv1.VirtualMachine) error {
	for {

		if vm.Status.Ready {
			fmt.Printf("VM Ready")
			return nil
		} else if vm.Status.StartFailure != nil {
			return fmt.Errorf("vm %s in %s failed and has %v state", vm.Name, vm.Namespace, vm.Status.StartFailure)
		} else {
			fmt.Printf("vm %s still being creating or pending, waiting 5 seconds...\n", vm.Name)
			time.Sleep(5 * time.Second)
			err := r.Get(ctx, types.NamespacedName{Name: vm.Name, Namespace: vm.Namespace}, vm)
			if err != nil {
				return err
			}

		}
	}
}

// func mapContainersToHosts(labTemplate *ltbbackendv1alpha1.LabTemplate) []corev1.Container {
// 	nodes := labTemplate.Spec.Nodes
// 	containers := []corev1.Container{}
// 	for _, node := range nodes {
// 		containers = append(containers, corev1.Container{
// 			Name:    node.Name,
// 			Image:   node.Image.Type + ":" + node.Image.Version,
// 			Command: []string{"/bin/sleep", "365d"},
// 		})
// 	}
// 	return containers
// }

// SetupWithManager sets up the controller with the Manager.
func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbbackendv1alpha1.LabInstance{}).
		Owns(&corev1.Pod{}).
		Owns(&kubevirtv1.VirtualMachine{}).
		Complete(r)
}
