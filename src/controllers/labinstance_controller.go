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

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/src/api/v1alpha1"
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
func (r *LabInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var labInstanceStatus ltbv1alpha1.LabInstanceStatus
	var err error
	var result ctrl.Result
	var shouldReturn bool
	labInstance := &ltbv1alpha1.LabInstance{}
	err = r.Get(ctx, req.NamespacedName, labInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("LabInstance resource not found.")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get LabInstance")
		return ctrl.Result{}, err
	}

	labTemplate := &ltbv1alpha1.LabTemplate{}
	// TODO: Maybe add fatal error handling here
	if shouldReturn, result, err := r.getLabTemplate(ctx, labInstance, labTemplate); shouldReturn {
		return result, err
	}

	nodes := labTemplate.Spec.Nodes
	pods := []*corev1.Pod{}
	vms := []*kubevirtv1.VirtualMachine{}
	for _, node := range nodes {
		if node.Image.Kind == "vm" {
			vm, shouldReturn, result, err := r.reconcileVM(ctx, labInstance, &node)
			if shouldReturn {
				return result, err
			}
			vms = append(vms, vm)
		} else {
			// If not vm, assume it is a pod
			pod, shouldReturn, result, err := r.reconcilePod(ctx, labInstance, &node)
			if shouldReturn {
				return result, err
			}
			pods = append(pods, pod)
		}
	}
	// check labInstance status
	shouldReturn, result, labInstanceStatus, err = r.getLabInstanceStatus(ctx, pods, vms, labInstance)
	if shouldReturn {
		return result, err
	}
	if labInstanceStatus.PodStatus == "Running" && labInstanceStatus.VMStatus == "VM Ready" {
		labInstance.Status.PodStatus = labInstanceStatus.PodStatus
		labInstance.Status.VMStatus = labInstanceStatus.VMStatus
		labInstance.Status.Status = "Running"
	} else {
		if labInstanceStatus.PodStatus != "Running" {
			labInstance.Status.Status = string(labInstanceStatus.PodStatus)
		} else {
			labInstance.Status.Status = labInstanceStatus.VMStatus
		}
	}

	fmt.Printf("\nLabInstance status => PodsStatus: %s VMsStatus: %s\n", labInstance.Status.PodStatus, labInstance.Status.VMStatus)
	fmt.Printf("LabInstanceStatus: %s\n", labInstance.Status.Status)

	err = r.Status().Update(ctx, labInstance)
	if err != nil {
		log.Error(err, "Failed to update LabInstance status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) getLabTemplate(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, labTemplate *ltbv1alpha1.LabTemplate) (bool, ctrl.Result, error) {
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

func (r *LabInstanceReconciler) reconcilePod(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (*corev1.Pod, bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	foundPod := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{Name: node.Name, Namespace: labInstance.Namespace}, foundPod)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Pod
		pod := mapTemplateToPod(labInstance, node)
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

func (r *LabInstanceReconciler) reconcileVM(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (*kubevirtv1.VirtualMachine, bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	foundVM := &kubevirtv1.VirtualMachine{}
	err := r.Get(ctx, types.NamespacedName{Name: node.Name, Namespace: labInstance.Namespace}, foundVM)
	if err != nil && errors.IsNotFound(err) {

		vm := mapTemplateToVM(labInstance, node)
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

func mapTemplateToPod(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) *corev1.Pod {
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
	}
	return pod
}

func mapTemplateToVM(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) *kubevirtv1.VirtualMachine {
	running := true
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

func (r *LabInstanceReconciler) getLabInstanceStatus(ctx context.Context, pods []*corev1.Pod, vms []*kubevirtv1.VirtualMachine, labInstance *ltbv1alpha1.LabInstance) (bool, ctrl.Result, ltbv1alpha1.LabInstanceStatus, error) {
	var podStatus string
	var vmStatus string
	var result ctrl.Result
	var shouldReturn bool
	for _, pod := range pods {
		shouldReturn, result, status, err := r.checkPodStatus(ctx, pod)
		fmt.Printf("Pod Status: %v\n", status)
		labInstance.Status.PodStatus = string(status.Phase)
		if err != nil {
			return shouldReturn, result, labInstance.Status, err
		} else {
			if status.Phase != corev1.PodRunning {
				podStatus = string(status.Phase)
				break
			}
			podStatus = string(corev1.PodRunning)

		}
	}
	labInstance.Status.PodStatus = podStatus

	for _, vm := range vms {
		shouldReturn, result, status, err := r.checkVMStatus(ctx, vm)

		if err != nil {
			return shouldReturn, result, labInstance.Status, err
		} else {
			if status.Ready {
				vmStatus = "VM Ready"
			} else {
				vmStatus = "Not Ready"
				break
			}
		}
	}
	labInstance.Status.VMStatus = vmStatus
	fmt.Println("LabInstance Status: ", labInstance.Status)
	return shouldReturn, result, labInstance.Status, nil
}

func (r *LabInstanceReconciler) checkPodStatus(ctx context.Context, pod *corev1.Pod) (bool, ctrl.Result, corev1.PodStatus, error) {
	phase := pod.Status.Phase
	if phase == corev1.PodRunning {
		return false, ctrl.Result{}, pod.Status, nil
	} else if phase == corev1.PodFailed || phase == corev1.PodUnknown {
		return true, ctrl.Result{RequeueAfter: 2 * time.Second}, pod.Status, fmt.Errorf("pod %s in %s is in %v state", pod.Name, pod.Namespace, phase)
	} else {
		err := r.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, pod)
		return true, ctrl.Result{RequeueAfter: 2 * time.Second}, pod.Status, err
	}
}

func (r *LabInstanceReconciler) checkVMStatus(ctx context.Context, vm *kubevirtv1.VirtualMachine) (bool, ctrl.Result, kubevirtv1.VirtualMachineStatus, error) {
	if vm.Status.Ready {
		return false, ctrl.Result{}, vm.Status, nil
	} else if vm.Status.StartFailure != nil {
		return true, ctrl.Result{RequeueAfter: 2 * time.Second}, vm.Status, fmt.Errorf("vm %s in %s failed and has %v state", vm.Name, vm.Namespace, vm.Status.StartFailure)
	} else {
		err := r.Get(ctx, types.NamespacedName{Name: vm.Name, Namespace: vm.Namespace}, vm)
		return true, ctrl.Result{RequeueAfter: 2 * time.Second}, vm.Status, err
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbv1alpha1.LabInstance{}).
		Owns(&corev1.Pod{}).
		Owns(&kubevirtv1.VirtualMachine{}).
		Complete(r)
}
