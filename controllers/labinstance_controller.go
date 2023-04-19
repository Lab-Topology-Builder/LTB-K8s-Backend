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

	network "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
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
	var err error
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
	if shouldReturn, result, err := r.getLabTemplate(ctx, labInstance, labTemplate); shouldReturn {
		return result, err
	}

	r.reconcileNetwork(ctx, labInstance)

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

	// Update LabInstance status according to the status of the pods and vms
	updateLabInstanceStatus(ctx, pods, vms, labInstance)

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
	return false, ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) reconcileNetwork(ctx context.Context, labInstance *ltbv1alpha1.LabInstance) (bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	podNetworkDefinitionName := labInstance.Name + "pod"
	vmNetworkDefinitionName := labInstance.Name + "vm"
	networkdefinitionNames := []string{podNetworkDefinitionName, vmNetworkDefinitionName}
	for _, networkDefinitionName := range networkdefinitionNames {
		foundNetworkAttachmentDefinition := &network.NetworkAttachmentDefinition{}
		err := r.Get(ctx, types.NamespacedName{Name: networkDefinitionName, Namespace: labInstance.Namespace}, foundNetworkAttachmentDefinition)
		if err != nil && errors.IsNotFound(err) {
			networkAttachmentDefinition := &network.NetworkAttachmentDefinition{}
			networkAttachmentDefinition.Name = networkDefinitionName
			networkAttachmentDefinition.Namespace = labInstance.Namespace
			if networkDefinitionName == podNetworkDefinitionName {
				// Don't change mode to "passthru" as it will takeover the kubernetes node interface and cause a network outage
				networkAttachmentDefinition.Spec.Config = `{
				"cniVersion": "0.3.1",
				"name": "mynet",
				"type": "bridge",
				"bridge": "mynet0",
				"ipam": {
					"type": "host-local",
					"ranges": [
						[ {
							"subnet": "10.10.0.0/24",
							"rangeStart": "10.10.0.10",
							"rangeEnd": "10.10.0.250"
						} ]
					]
				}
			}`
			} else {
				networkAttachmentDefinition.Spec.Config = `{
					"cniVersion": "0.3.1",
					"name": "mynet",
					"type": "bridge",
					"bridge": "mynet0",
					"ipam": {}
				}`
			}
			ctrl.SetControllerReference(labInstance, networkAttachmentDefinition, r.Scheme)
			log.Info("Creating a new NetworkAttachmentDefinition", "NetworkAttachmentDefinition.Namespace", networkAttachmentDefinition.Namespace, "NetworkAttachmentDefinition.Name", networkAttachmentDefinition.Name)
			err = r.Create(ctx, networkAttachmentDefinition)
			if err != nil {
				log.Error(err, "Failed to create NetworkAttachmentDefinition")
				return true, ctrl.Result{}, err
			}
			return true, ctrl.Result{Requeue: true}, nil
		} else if err != nil {
			log.Error(err, "Failed to get NetworkAttachmentDefinition")
			return true, ctrl.Result{}, err
		}
	}
	return false, ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) reconcilePod(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (*corev1.Pod, bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	foundPod := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{Name: labInstance.Name + node.Name, Namespace: labInstance.Namespace}, foundPod)
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
	err := r.Get(ctx, types.NamespacedName{Name: labInstance.Name + node.Name, Namespace: labInstance.Namespace}, foundVM)
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
		Name:      labInstance.Name + node.Name,
		Namespace: labInstance.Namespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks": labInstance.Name + "pod",
		},
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
		Name:      labInstance.Name + node.Name,
		Namespace: labInstance.Namespace,
	}
	disks := []kubevirtv1.Disk{
		{Name: "containerdisk", DiskDevice: kubevirtv1.DiskDevice{Disk: &kubevirtv1.DiskTarget{Bus: "virtio"}}},
		{Name: "cloudinitdisk", DiskDevice: kubevirtv1.DiskDevice{Disk: &kubevirtv1.DiskTarget{Bus: "virtio"}}},
	}
	volumes := []kubevirtv1.Volume{
		{Name: "containerdisk", VolumeSource: kubevirtv1.VolumeSource{ContainerDisk: &kubevirtv1.ContainerDiskSource{Image: "quay.io/containerdisks/" + node.Image.Type + ":" + node.Image.Version}}},
		{Name: "cloudinitdisk", VolumeSource: kubevirtv1.VolumeSource{CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{UserData: node.Config}}},
	}
	networks := []kubevirtv1.Network{
		{Name: "default", NetworkSource: kubevirtv1.NetworkSource{Pod: &kubevirtv1.PodNetwork{}}},
		{Name: labInstance.Name, NetworkSource: kubevirtv1.NetworkSource{Multus: &kubevirtv1.MultusNetwork{NetworkName: labInstance.Name + "vm"}}},
	}
	interfaces := []kubevirtv1.Interface{
		{Name: "default", InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{Bridge: &kubevirtv1.InterfaceBridge{}}},
		{Name: labInstance.Name, InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{Bridge: &kubevirtv1.InterfaceBridge{}}},
	}
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
							Disks:      disks,
							Interfaces: interfaces,
						},
					},
					Volumes:  volumes,
					Networks: networks,
				},
			},
		},
	}
	return vm
}

func updateLabInstanceStatus(ctx context.Context, pods []*corev1.Pod, vms []*kubevirtv1.VirtualMachine, labInstance *ltbv1alpha1.LabInstance) {
	var podStatus corev1.PodPhase
	var vmStatus kubevirtv1.VirtualMachinePrintableStatus
	for _, pod := range pods {
		podStatus = pod.Status.Phase
		if pod.Status.Phase != corev1.PodRunning {
			labInstance.Status.PodStatus = string(pod.Status.Phase)
			break
		}
	}
	labInstance.Status.PodStatus = string(podStatus)

	for _, vm := range vms {
		vmStatus = vm.Status.PrintableStatus
		if !vm.Status.Ready {
			labInstance.Status.VMStatus = string(vmStatus)
			break
		}
	}
	labInstance.Status.VMStatus = string(vmStatus)

	if labInstance.Status.PodStatus == "Running" && labInstance.Status.VMStatus == "VM Ready" {
		labInstance.Status.Status = "Running"
	} else {
		if labInstance.Status.PodStatus != "Running" {
			labInstance.Status.Status = string(labInstance.Status.PodStatus)
		} else {
			labInstance.Status.Status = labInstance.Status.VMStatus
		}
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
