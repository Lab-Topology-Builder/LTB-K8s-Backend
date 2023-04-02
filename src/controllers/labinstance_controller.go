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
// (user): Modify the Reconcile function to compare the state specified by
// the LabInstance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
// TODO: Check if deployment already exists, if not create a new one, or update an existing one
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
	// TODO: Maybe add fatal error handling here
	if shouldReturn, result, err := r.getLabTemplate(ctx, labInstance, labTemplate); shouldReturn {
		return result, err
	}

	pod, shouldReturn, result, err := r.reconcilePod(ctx, labInstance, labTemplate)
	if shouldReturn {
		return result, err
	}

	// Check status of the pod
	if shouldReturn, result, err := r.checkPodStatus(ctx, pod); shouldReturn {
		return result, err
	}

	foundVM, shouldReturn, result, err := r.reconcileVM(ctx, labInstance, labTemplate)
	if shouldReturn {
		return result, err
	}

	// Check status of the VM
	if shouldReturn, result, err := r.checkVMStatus(ctx, foundVM); shouldReturn {
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) getLabTemplate(ctx context.Context, labInstance *ltbbackendv1alpha1.LabInstance, labTemplate *ltbbackendv1alpha1.LabTemplate) (bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	err := r.Get(ctx, types.NamespacedName{Name: labInstance.Spec.LabTemplateReference, Namespace: labInstance.Namespace}, labTemplate)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("LabTemplate resource not found.")
			return true, ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get LabTemplate")
		return true, ctrl.Result{}, err
	}
	log.Info("LabTemplate resource found.", "LabTemplate.Namespace", labTemplate.Namespace, "LabTemplate.Name", labTemplate.Name)
	return false, ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) reconcileVM(ctx context.Context, labInstance *ltbbackendv1alpha1.LabInstance, labTemplate *ltbbackendv1alpha1.LabTemplate) (*kubevirtv1.VirtualMachine, bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	node := labTemplate.Spec.Nodes[0]
	foundVM := &kubevirtv1.VirtualMachine{}
	err := r.Get(ctx, types.NamespacedName{Name: node.Name, Namespace: labInstance.Namespace}, foundVM)
	if err != nil && errors.IsNotFound(err) {

		vm := mapTemplateToVM(labInstance, labTemplate)
		ctrl.SetControllerReference(labInstance, vm, r.Scheme)
		log.Info("Creating a new VM", "VM.Namespace", vm.Namespace, "VM.Name", vm.Name)
		err = r.Create(ctx, vm)
		if err != nil {
			log.Error(err, "Failed to create new VM", "VM.Namespace", vm.Namespace, "VM.Name", vm.Name)
			return nil, true, ctrl.Result{}, err
		}

		return nil, true, ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get VM")
		return nil, true, ctrl.Result{}, err
	}
	return foundVM, false, ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) reconcilePod(ctx context.Context, labInstance *ltbbackendv1alpha1.LabInstance, labTemplate *ltbbackendv1alpha1.LabTemplate) (*corev1.Pod, bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	node := labTemplate.Spec.Nodes[1]
	foundPod := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{Name: node.Name, Namespace: labInstance.Namespace}, foundPod)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Pod
		pod := mapTemplateToPod(labInstance, labTemplate)
		ctrl.SetControllerReference(labInstance, pod, r.Scheme)
		log.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.Create(ctx, pod)
		if err != nil {
			log.Error(err, "Failed to create new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			return pod, true, ctrl.Result{}, err
		}
		// Pod created successfully - return and requeue
		return pod, true, ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Pod")
		return foundPod, true, ctrl.Result{}, err
	}
	return foundPod, false, ctrl.Result{}, nil
}

func mapTemplateToPod(labInstance *ltbbackendv1alpha1.LabInstance, labTemplate *ltbbackendv1alpha1.LabTemplate) *corev1.Pod {
	node := labTemplate.Spec.Nodes[1]
	metadata := metav1.ObjectMeta{
		Name:      node.Name,
		Namespace: labInstance.Namespace,
	}
	pod := &corev1.Pod{
		ObjectMeta: metadata,
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    node.Name,
					Image:   node.Image.Type + ":" + node.Image.Version,
					Command: []string{"/bin/sleep", "365d"},
				},
			},
		},
		Status: labInstance.Status.Phase, // TODO check if this is needed
	}
	return pod
}

func mapTemplateToVM(labInstance *ltbbackendv1alpha1.LabInstance, labTemplate *ltbbackendv1alpha1.LabTemplate) *kubevirtv1.VirtualMachine {
	running := true
	node := labTemplate.Spec.Nodes[0]
	resources := kubevirtv1.ResourceRequirements{
		Requests: corev1.ResourceList{"memory": resource.MustParse("2048M")},
	}
	cpu := &kubevirtv1.CPU{Cores: 1}
	metadata := metav1.ObjectMeta{
		Name:      node.Name,
		Namespace: labInstance.Namespace,
	}
	disks := []kubevirtv1.Disk{
		{Name: "containerdisk", DiskDevice: kubevirtv1.DiskDevice{Disk: &kubevirtv1.DiskTarget{Bus: "virtio"}}},
		{Name: "cloudinitdisk", DiskDevice: kubevirtv1.DiskDevice{Disk: &kubevirtv1.DiskTarget{Bus: "virtio"}}},
	}
	volumes := []kubevirtv1.Volume{
		{Name: "containerdisk", VolumeSource: kubevirtv1.VolumeSource{ContainerDisk: &kubevirtv1.ContainerDiskSource{Image: "quay.io/containerdisks/" + node.Image.Type + ":" + node.Image.Version}}},
		{Name: "cloudinitdisk", VolumeSource: kubevirtv1.VolumeSource{CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{UserData: node.Config}}}}
	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metadata,
		Spec: kubevirtv1.VirtualMachineSpec{
			Running: &running,
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Domain: kubevirtv1.DomainSpec{
						Resources: resources,
						CPU:       cpu,
						Devices: kubevirtv1.Devices{
							Disks: disks,
						},
					},
					Volumes: volumes,
				},
			},
		},
	}
	return vm
}

// TODO: continue working on this
//func (r *LabInstanceReconciler) defineFieldsForKubectl(podStatus corev1.PodStatus, pod *corev1.Pod) {
// Define fields for kubectl
//	header := []string{"NAME", "AGE", "STATUS", "POD-NAME"}
//	formatString := "{{.metadata.name}}\t{{.metadata.creationTimestamp}}\t{{.status.phase}}\t{{.metadata.name}}"
//	if podStatus.Phase == corev1.PodRunning {
//		podName := "All"
//	}
//
//}

// Just ignore this function for now: this is the updated version of the function checkPodStatus and is work in progress
func (r *LabInstanceReconciler) getPodStatus(ctx context.Context, labInstance *ltbbackendv1alpha1.LabInstance) (corev1.PodStatus, string, error) {
	podList := &corev1.PodList{}
	var podName string
	var podStatus corev1.PodStatus
	var err error

	listOpts := []client.ListOption{client.InNamespace(labInstance.Namespace)}
	if err := r.List(ctx, podList, listOpts...); err != nil {
		return corev1.PodStatus{}, "None", err
	}
	for _, pod := range podList.Items {
		if pod.Status.Phase == corev1.PodFailed {
			podStatus = pod.Status
			podName = pod.Name
			err = fmt.Errorf("pod %s in %s is in %v state", pod.Name, pod.Namespace, pod.Status.Phase)
			break
		} else if pod.Status.Phase == corev1.PodUnknown {
			podStatus = pod.Status
			podName = pod.Name
			err = fmt.Errorf("pod %s in %s is in %v state", pod.Name, pod.Namespace, pod.Status.Phase)
			break
		} else if pod.Status.Phase == corev1.PodRunning {
			podStatus = pod.Status
			podName = pod.Name
			err = nil
		} else if pod.Status.Phase == corev1.PodPending {
			podStatus = pod.Status
			podName = pod.Name
			err = fmt.Errorf("pod %s in %s is in %v state", pod.Name, pod.Namespace, pod.Status.Phase)
			break
		}
	}
	return podStatus, podName, err
}

func (r *LabInstanceReconciler) checkPodStatus(ctx context.Context, pod *corev1.Pod) (bool, ctrl.Result, error) {
	phase := pod.Status.Phase
	fmt.Printf("Pod status: %v\n", phase)
	if phase == corev1.PodRunning {
		return false, ctrl.Result{}, nil
	} else if phase == corev1.PodFailed || phase == corev1.PodUnknown {
		return true, ctrl.Result{RequeueAfter: 2 * time.Second}, fmt.Errorf("pod %s in %s is in %v state", pod.Name, pod.Namespace, phase)
	} else {
		err := r.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pod)
		return true, ctrl.Result{RequeueAfter: 2 * time.Second}, err
	}
}

func (r *LabInstanceReconciler) checkVMStatus(ctx context.Context, vm *kubevirtv1.VirtualMachine) (bool, ctrl.Result, error) {
	if vm.Status.Ready {
		fmt.Printf("VM Ready")
		return false, ctrl.Result{}, nil
	} else if vm.Status.StartFailure != nil {
		return true, ctrl.Result{RequeueAfter: 2 * time.Second}, fmt.Errorf("vm %s in %s failed and has %v state", vm.Name, vm.Namespace, vm.Status.StartFailure)
	} else {
		err := r.Get(ctx, types.NamespacedName{Name: vm.Name, Namespace: vm.Namespace}, vm)
		return true, ctrl.Result{RequeueAfter: 2 * time.Second}, err
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbbackendv1alpha1.LabInstance{}).
		Owns(&corev1.Pod{}).
		Owns(&kubevirtv1.VirtualMachine{}).
		Complete(r)
}
