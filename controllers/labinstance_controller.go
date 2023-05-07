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
//+kubebuilder:scaffold:scheme

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	// TODO: refactor not found error handling
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
	if shouldReturn, result, err := r.GetLabTemplate(ctx, labInstance, labTemplate); shouldReturn {
		return result, err
	}
	r.ReconcileNetwork(ctx, labInstance)

	node := &ltbv1alpha1.LabInstanceNodes{}
	r.ReconcileSvcAccRoleRoleBind(ctx, labInstance)

	r.ReconcileService(ctx, labInstance, labInstance.Name+"-ttyd-service", "pod")
	// Reconile ttyd pod
	r.ReconcilePod(ctx, labInstance, node)

	nodes := labTemplate.Spec.Nodes
	pods := []*corev1.Pod{}
	vms := []*kubevirtv1.VirtualMachine{}
	for _, node := range nodes {
		if node.Image.Kind == "vm" {
			vm, shouldReturn, result, err := r.ReconcileVM(ctx, labInstance, &node)
			if shouldReturn {
				return result, err
			}
			vms = append(vms, vm)
		} else {
			// If not vm, assume it is a pod
			pod, shouldReturn, result, err := r.ReconcilePod(ctx, labInstance, &node)
			if shouldReturn {
				return result, err
			}
			pods = append(pods, pod)
		}
		kind := node.Image.Kind
		if kind == "" {
			kind = "pod"
		}
		r.ReconcileService(ctx, labInstance, labInstance.Name+"-"+node.Name+"-remote-access", kind)
		r.ReconcileIngress(ctx, labInstance, &node)

	}

	// Update LabInstance status according to the status of the pods and vms
	UpdateLabInstanceStatus(ctx, pods, vms, labInstance)

	err = r.Status().Update(ctx, labInstance)
	if err != nil {
		log.Error(err, "Failed to update LabInstance status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) GetLabTemplate(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, labTemplate *ltbv1alpha1.LabTemplate) (bool, ctrl.Result, error) {
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

func (r *LabInstanceReconciler) ReconcileNetwork(ctx context.Context, labInstance *ltbv1alpha1.LabInstance) (bool, ctrl.Result, error) {
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

func (r *LabInstanceReconciler) ReconcilePod(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (*corev1.Pod, bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	foundPod := &corev1.Pod{}
	var pod *corev1.Pod
	var name string
	var podType string
	if node.Name != "" {
		name = labInstance.Name + "-" + node.Name
		podType = "pod"
	} else {
		name = labInstance.Name + "-ttyd"
		podType = "ttyd"
	}
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: labInstance.Namespace}, foundPod)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Pod
		if podType == "ttyd" {
			pod, _ = CreateTtydPodAndService(labInstance)
		} else {
			pod = MapTemplateToPod(labInstance, node)
		}
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

func (r *LabInstanceReconciler) ReconcileVM(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (*kubevirtv1.VirtualMachine, bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	foundVM := &kubevirtv1.VirtualMachine{}
	err := r.Get(ctx, types.NamespacedName{Name: labInstance.Name + "-" + node.Name, Namespace: labInstance.Namespace}, foundVM)
	if err != nil && errors.IsNotFound(err) {

		vm := MapTemplateToVM(labInstance, node)
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

func (r *LabInstanceReconciler) ReconcileService(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, serviceName string, kind string) (bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	var service *corev1.Service
	foundService := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: serviceName, Namespace: labInstance.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Service
		if serviceName == labInstance.Name+"-ttyd-service" {
			_, service = CreateTtydPodAndService(labInstance)
		} else {
			service = CreateService(labInstance, serviceName, kind)
		}
		ctrl.SetControllerReference(labInstance, service, r.Scheme)
		log.Info("Creating a new Service", "Service.Namespace", labInstance.Namespace, "Service.Name", service.Name)
		err = r.Create(ctx, service)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", labInstance.Namespace, "Service.Name", service.Name)
			return true, ctrl.Result{}, err
		}
		// Service created successfully - return and requeue
		return true, ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return true, ctrl.Result{}, err
	}
	return false, ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) ReconcileIngress(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	foundIngress := &networkingv1.Ingress{}
	name := labInstance.Name + "-" + node.Name
	var resourceType string
	if node.Image.Kind != "" {
		resourceType = node.Image.Kind
	} else {
		resourceType = "pod"
	}

	err := r.Get(ctx, types.NamespacedName{Name: name + "-ingress", Namespace: labInstance.Namespace}, foundIngress)
	if err != nil && errors.IsNotFound(err) {
		// Define a new Ingress
		ingress := CreateIngress(labInstance, resourceType, name)
		ctrl.SetControllerReference(labInstance, ingress, r.Scheme)
		log.Info("Creating a new Ingress", "Ingress.Namespace", labInstance.Namespace, "Ingress.Name", ingress.Name)
		err = r.Create(ctx, ingress)
		if err != nil {
			log.Error(err, "Failed to create new Ingress", "Ingress.Namespace", labInstance.Namespace, "Ingress.Name", ingress.Name)
			return true, ctrl.Result{}, err
		}
		// Ingress created successfully - return and requeue
		return true, ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Ingress")
		return true, ctrl.Result{}, err
	}
	return false, ctrl.Result{}, nil
}

func (r *LabInstanceReconciler) ReconcileSvcAccRoleRoleBind(ctx context.Context, labInstance *ltbv1alpha1.LabInstance) (bool, ctrl.Result, error) {
	log := log.FromContext(ctx)
	foundRole := &rbacv1.Role{}
	foundSvcAcc := &corev1.ServiceAccount{}
	foundRoleBind := &rbacv1.RoleBinding{}
	err := r.Get(ctx, types.NamespacedName{Name: labInstance.Name + "-ttyd-role", Namespace: labInstance.Namespace}, foundRole)
	err = r.Get(ctx, types.NamespacedName{Name: labInstance.Name + "-ttyd-svcacc", Namespace: labInstance.Namespace}, foundSvcAcc)
	err = r.Get(ctx, types.NamespacedName{Name: labInstance.Name + "-ttyd-rolebind", Namespace: labInstance.Namespace}, foundRoleBind)
	if err != nil && errors.IsNotFound(err) {
		svcAcc, role, roleBind := CreateSvcAccRoleRoleBind(labInstance)
		ctrl.SetControllerReference(labInstance, svcAcc, r.Scheme)
		ctrl.SetControllerReference(labInstance, role, r.Scheme)
		ctrl.SetControllerReference(labInstance, roleBind, r.Scheme)
		err = r.Create(ctx, svcAcc)
		if err != nil {
			log.Error(err, "Failed to create new ServiceAccount", "ServiceAccount.Namespace", labInstance.Namespace, "ServiceAccount.Name", svcAcc.Name)
			return true, ctrl.Result{}, err
		}
		err = r.Create(ctx, role)
		if err != nil {
			log.Error(err, "Failed to create new Role", "Role.Namespace", labInstance.Namespace, "Role.Name", role.Name)
			return true, ctrl.Result{}, err
		}
		err = r.Create(ctx, roleBind)
		if err != nil {
			log.Error(err, "Failed to create new RoleBinding", "RoleBinding.Namespace", labInstance.Namespace, "RoleBinding.Name", roleBind.Name)
			return true, ctrl.Result{}, err
		}
		return true, ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get Role or ServiceAccount or RoleBinding")
		return true, ctrl.Result{}, err
	}
	return false, ctrl.Result{}, nil
}

func MapTemplateToPod(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) *corev1.Pod {
	metadata := metav1.ObjectMeta{
		Name:      labInstance.Name + "-" + node.Name,
		Namespace: labInstance.Namespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks": labInstance.Name + "pod",
		},
		Labels: map[string]string{
			"app": labInstance.Name + "-" + node.Name + "-remote-access",
		},
	}
	pod := &corev1.Pod{
		ObjectMeta: metadata,

		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    node.Name,
					Image:   node.Image.Type + ":" + node.Image.Version,
					Command: []string{"/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"},
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: labInstance.Spec.Port,
						},
					},
				},
			},
		},
	}
	return pod
}

func MapTemplateToVM(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) *kubevirtv1.VirtualMachine {
	running := true
	resources := kubevirtv1.ResourceRequirements{
		Requests: corev1.ResourceList{"memory": resource.MustParse("2048M")},
	}
	cpu := &kubevirtv1.CPU{Cores: 1}
	metadata := metav1.ObjectMeta{
		Name:      labInstance.Name + "-" + node.Name,
		Namespace: labInstance.Namespace,
		Labels: map[string]string{
			"special": labInstance.Name + "-" + node.Name + "-remote-access",
		},
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

func UpdateLabInstanceStatus(ctx context.Context, pods []*corev1.Pod, vms []*kubevirtv1.VirtualMachine, labInstance *ltbv1alpha1.LabInstance) {
	var podStatus corev1.PodPhase
	var vmStatus kubevirtv1.VirtualMachinePrintableStatus
	var numVMsRunning, numPodsRunning int
	for _, pod := range pods {
		podStatus = pod.Status.Phase
		if podStatus != corev1.PodRunning {
			break
		}
		numPodsRunning++
	}
	labInstance.Status.NumPodsRunning = fmt.Sprint(numPodsRunning) + "/" + fmt.Sprint(len(pods))

	for _, vm := range vms {
		vmStatus = vm.Status.PrintableStatus
		if !vm.Status.Ready {
			break
		}
		numVMsRunning++
	}
	labInstance.Status.NumVMsRunning = fmt.Sprint(numVMsRunning) + "/" + fmt.Sprint(len(vms))

	if podStatus == "Running" && vmStatus == "VM Ready" {
		labInstance.Status.Status = "Running"
	} else {
		if podStatus != "Running" {
			labInstance.Status.Status = string(podStatus)
		} else {
			labInstance.Status.Status = string(vmStatus)
		}
	}
}

func CreateIngress(labInstance *ltbv1alpha1.LabInstance, resourceType string, name string) *networkingv1.Ingress {
	ingressName := name + "-ingress"
	metadata := metav1.ObjectMeta{
		Name:      ingressName,
		Namespace: labInstance.Namespace,
		Annotations: map[string]string{
			"nginx.ingress.kubernetes.io/rewrite-target": "/?arg=" + resourceType + "&arg=" + name + "&arg=bash",
		},
	}
	className := "nginx"
	ingress := &networkingv1.Ingress{
		ObjectMeta: metadata,
		Spec: networkingv1.IngressSpec{
			IngressClassName: &className,
			Rules: []networkingv1.IngressRule{
				{Host: ingressName + ".sr-118142.network.garden",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path: "/",
									PathType: func() *networkingv1.PathType {
										pathType := networkingv1.PathTypePrefix
										return &pathType
									}(),
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: labInstance.Name + "-ttyd-service",
											Port: networkingv1.ServiceBackendPort{
												Name: "ttyd",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return ingress
}

func CreateTtydPodAndService(labInstance *ltbv1alpha1.LabInstance) (*corev1.Pod, *corev1.Service) {
	metadata := metav1.ObjectMeta{
		Name:      labInstance.Name + "-ttyd",
		Namespace: labInstance.Namespace,
		Labels: map[string]string{
			"app": "ttyd-app",
		},
	}
	pod := &corev1.Pod{
		ObjectMeta: metadata,

		Spec: corev1.PodSpec{
			ServiceAccountName: labInstance.Name + "-ttyd-svcacc",
			Containers: []corev1.Container{
				{
					Name:  labInstance.Name + "-ttyd-container",
					Image: "ghcr.io/insrapperswil/kube-ttyd:latest",
					Args:  []string{"ttyd", "-a", "konnect"},
					Ports: []corev1.ContainerPort{
						{
							ContainerPort: 7681,
						},
					},
				},
			},
		},
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labInstance.Name + "-ttyd-service",
			Namespace: labInstance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       7681,
					TargetPort: intstr.IntOrString{IntVal: 7681},
					Name:       "ttyd",
				},
			},
			Selector: map[string]string{
				"app": "ttyd-app",
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}
	return pod, service
}

func CreateService(labInstance *ltbv1alpha1.LabInstance, serviceName string, kind string) *corev1.Service {
	selectors := map[string]string{}
	if kind == "vm" {
		selectors = map[string]string{
			"special": serviceName,
		}
	} else {
		selectors = map[string]string{
			"app": serviceName,
		}
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: labInstance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       labInstance.Spec.Port,
					TargetPort: intstr.IntOrString{IntVal: labInstance.Spec.Port},
					Name:       "ssh",
				},
			},
			Selector: selectors,
			Type:     corev1.ServiceTypeNodePort,
		},
	}
	return service
}

func CreateSvcAccRoleRoleBind(labInstance *ltbv1alpha1.LabInstance) (*corev1.ServiceAccount, *rbacv1.Role, *rbacv1.RoleBinding) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labInstance.Name + "-ttyd-svcacc",
			Namespace: labInstance.Namespace,
		},
	}
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labInstance.Name + "-ttyd-role",
			Namespace: labInstance.Namespace,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/log"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"pods/exec"},
				Verbs:     []string{"create"},
			},
			{
				APIGroups: []string{"subresources.kubevirt.io"},
				Resources: []string{"virtualmachineinstances/console"},
				Verbs:     []string{"get", "list", "create", "update", "delete"},
			},
		},
	}

	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      labInstance.Name + "-ttyd-rolebind",
			Namespace: labInstance.Namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      labInstance.Name + "-ttyd-svcacc",
				Namespace: labInstance.Namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     labInstance.Name + "-ttyd-role",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	return serviceAccount, role, roleBinding

}

// SetupWithManager sets up the controller with the Manager.
func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbv1alpha1.LabInstance{}).
		Owns(&corev1.Pod{}).
		Owns(&kubevirtv1.VirtualMachine{}).
		Complete(r)
}
