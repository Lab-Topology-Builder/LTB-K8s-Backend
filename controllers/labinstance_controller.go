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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	// "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/yaml"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	network "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
)

type LabInstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type ReturnToReconciler struct {
	shouldReturn bool
	result       ctrl.Result
	err          error
}

//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labinstances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labinstances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labinstances/finalizers,verbs=update
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubevirt.io,resources=virtualmachines/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=pods/status,verbs=get;update;patch

func (r *LabInstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var err error
	labInstance := &ltbv1alpha1.LabInstance{}
	err = r.Get(ctx, req.NamespacedName, labInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("LabInstance resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get LabInstance")
		return ctrl.Result{}, err
	}

	labTemplate := &ltbv1alpha1.LabTemplate{}
	retValue := r.GetLabTemplate(ctx, labInstance, labTemplate)
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	// Reconcile Network
	retValue = r.ReconcileNetwork(ctx, labInstance)
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	// Reconcile TTYD Service Account
	sa := corev1.ServiceAccount{}
	sa.Name = labInstance.Name + "-ttyd-svcacc"
	_, retValue = ReconcileResource(r, labInstance, &sa, nil, "")
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	// Reconcile TTYD Role
	role := rbacv1.Role{}
	role.Name = labInstance.Name + "-ttyd-role"
	_, retValue = ReconcileResource(r, labInstance, &role, nil, "")
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	// Reconcile TTYD Role Binding
	roleBinding := rbacv1.RoleBinding{}
	roleBinding.Name = labInstance.Name + "-ttyd-rolebind"
	_, retValue = ReconcileResource(r, labInstance, &roleBinding, nil, "")
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	// Reconcile TTYD Service
	ttydService := corev1.Service{}
	ttydService.Name = labInstance.Name + "-ttyd-service"
	_, retValue = ReconcileResource(r, labInstance, &ttydService, nil, "")
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	// Reconcile TTYD Pod
	ttydPod := corev1.Pod{}
	ttydPod.Name = labInstance.Name + "-ttyd-pod"
	_, retValue = ReconcileResource(r, labInstance, &ttydPod, nil, "")
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	nodes := labTemplate.Spec.Nodes
	pods := []*corev1.Pod{}
	vms := []*kubevirtv1.VirtualMachine{}
	for _, node := range nodes {
		nodeType := &ltbv1alpha1.NodeType{}
		retValue = r.GetNodeType(ctx, &node.NodeTypeRef, nodeType)
		if retValue.shouldReturn {
			return retValue.result, retValue.err
		}
		if nodeType.Spec.Kind == "vm" {
			virtualMachine := kubevirtv1.VirtualMachine{}
			virtualMachine.Name = labInstance.Name + "-" + node.Name
			vm, retValue := ReconcileResource(r, labInstance, &virtualMachine, &node, "")
			if retValue.shouldReturn {
				return retValue.result, retValue.err
			}
			vms = append(vms, vm.(*kubevirtv1.VirtualMachine))
		} else {
			p := corev1.Pod{}
			p.Name = labInstance.Name + "-" + node.Name
			pod, retValue := ReconcileResource(r, labInstance, &p, &node, "")
			if retValue.shouldReturn {
				return retValue.result, retValue.err
			}
			pods = append(pods, pod.(*corev1.Pod))
		}

		// Reconcile Remote Access Service
		if len(node.Ports) > 0 {
			s := corev1.Service{}
			s.Name = labInstance.Name + "-" + node.Name + "-remote-access"
			_, retValue = ReconcileResource(r, labInstance, &s, &node, "")
			if retValue.shouldReturn {
				return retValue.result, retValue.err
			}
		}

		// Reconcile Ingress
		ingress := networkingv1.Ingress{}
		ingress.Name = labInstance.Name + "-" + node.Name + "-ingress"
		_, retValue = ReconcileResource(r, labInstance, &ingress, &node, nodeType.Spec.Kind)
		if retValue.shouldReturn {
			return retValue.result, retValue.err
		}

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

func (r *LabInstanceReconciler) ReconcileNetwork(ctx context.Context, labInstance *ltbv1alpha1.LabInstance) ReturnToReconciler {
	log := log.FromContext(ctx)
	retValue := ReturnToReconciler{shouldReturn: true, result: ctrl.Result{}, err: nil}
	podNetworkDefinitionName := labInstance.Name + "-pod"
	vmNetworkDefinitionName := labInstance.Name + "-vm"
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
				retValue.err = err
				log.Error(err, "Failed to create NetworkAttachmentDefinition")
				return retValue
			}
			retValue.result = ctrl.Result{Requeue: true}
			return retValue
		} else if err != nil {
			retValue.err = err
			log.Error(err, "Failed to get NetworkAttachmentDefinition")
			return retValue
		}
	}
	retValue.shouldReturn = false
	return retValue
}

func ReconcileResource(r *LabInstanceReconciler, labInstance *ltbv1alpha1.LabInstance, resource client.Object, node *ltbv1alpha1.LabInstanceNodes, kind string) (client.Object, ReturnToReconciler) {
	ctx := context.Context(context.Background())
	log := log.FromContext(ctx)
	retValue := ReturnToReconciler{shouldReturn: true, result: ctrl.Result{}, err: nil}
	resource.SetNamespace(labInstance.Namespace)

	resourceExists, err := r.ResourceExists(resource)
	if err == nil && !resourceExists {
		createdResource, err := CreateResource(labInstance, node, resource, kind)
		if err != nil {
			retValue.err = err
			log.Error(err, "Failed to create new resource", "resource.Namespace", labInstance.Namespace, "resource.Name", reflect.ValueOf(*createdResource).Elem().FieldByName("Name"))
			return nil, retValue
		}
		log.Info("Creating a new resource", "resource.Namespace", labInstance.Namespace, "resource.Name", reflect.ValueOf(*createdResource).Elem().FieldByName("Name"))
		ctrl.SetControllerReference(labInstance, *createdResource, r.Scheme)

		err = r.Create(ctx, *createdResource)
		if err != nil {
			retValue.err = err
			log.Error(err, "Failed to create new resource", "resource.Namespace", labInstance.Namespace, "resource.Name", reflect.ValueOf(*createdResource).Elem().FieldByName("Name"))
			return nil, retValue
		}
		retValue.result = ctrl.Result{Requeue: true}
		return *createdResource, retValue
	} else if err != nil {
		retValue.err = err
		log.Error(err, "Failed to get resource")
		return resource, retValue
	}
	retValue.shouldReturn = false
	return resource, retValue
}

func CreateResource(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes, resource client.Object, kind string) (*client.Object, error) {
	ctx := context.Context(context.Background())
	log := log.FromContext(ctx)
	switch reflect.TypeOf(resource).Elem().Name() {
	case "Pod":
		resource = CreatePod(labInstance, node)
	case "VirtualMachine":
		resource = MapTemplateToVM(labInstance, node)
	case "Service":
		resource = CreateService(labInstance, node)
	case "Ingress":
		resource = CreateIngress(labInstance, node, kind)
	case "Role":
		_, role, _ := CreateSvcAccRoleRoleBind(labInstance)
		resource = role
	case "ServiceAccount":
		svcAcc, _, _ := CreateSvcAccRoleRoleBind(labInstance)
		resource = svcAcc
	case "RoleBinding":
		_, _, roleBind := CreateSvcAccRoleRoleBind(labInstance)
		resource = roleBind
	default:
		log.Info("Resource type not supported", "ResourceKind", resource.GetObjectKind().GroupVersionKind().Kind)
		return nil, errors.NewBadRequest(fmt.Sprintf("Resource type not supported: %s", reflect.TypeOf(resource).Elem().Name()))
	}
	return &resource, nil

}

func (r *LabInstanceReconciler) ResourceExists(resource client.Object) (bool, error) {
	ctx := context.Context(context.Background())
	resourceName := reflect.ValueOf(resource).Elem().FieldByName("Name").String()
	nameSpace := reflect.ValueOf(resource).Elem().FieldByName("Namespace").String()
	err := r.Get(ctx, types.NamespacedName{Name: resourceName, Namespace: nameSpace}, resource)
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (r *LabInstanceReconciler) GetLabTemplate(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, labTemplate *ltbv1alpha1.LabTemplate) ReturnToReconciler {
	log := log.FromContext(ctx)
	returnValue := ReturnToReconciler{shouldReturn: false, result: ctrl.Result{}, err: nil}
	err := r.Get(ctx, types.NamespacedName{Name: labInstance.Spec.LabTemplateReference, Namespace: labInstance.Namespace}, labTemplate)
	if err != nil && errors.IsNotFound(err) {
		log.Info("LabTemplate not found", "LabTemplate", labInstance.Spec.LabTemplateReference)
		returnValue.shouldReturn = true
		return returnValue
	} else if err != nil {
		returnValue.shouldReturn = true
		returnValue.err = err
		log.Error(err, "Failed to get LabTemplate")
		return returnValue
	}
	return returnValue
}

func (r *LabInstanceReconciler) GetNodeType(ctx context.Context, nodeTypeRef *ltbv1alpha1.NodeTypeRef, nodeType *ltbv1alpha1.NodeType) ReturnToReconciler {
	log := log.FromContext(ctx)
	returnValue := ReturnToReconciler{shouldReturn: false, result: ctrl.Result{}, err: nil}
	err := r.Get(ctx, types.NamespacedName{Name: nodeTypeRef.Type}, nodeType)
	if err != nil && errors.IsNotFound(err) {
		log.Info("NodeType not found", "NodeType", nodeTypeRef.Type)
		returnValue.shouldReturn = true
		return returnValue
	} else if err != nil {
		returnValue.shouldReturn = true
		returnValue.err = err
		log.Error(err, "Failed to get NodeType")
		return returnValue
	}
	return returnValue
}

func MapTemplateToPod(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) *corev1.Pod {
	log := log.FromContext(context.Background())
	metadata := metav1.ObjectMeta{
		Name:      labInstance.Name + "-" + node.Name,
		Namespace: labInstance.Namespace,
		Annotations: map[string]string{
			"k8s.v1.cni.cncf.io/networks": labInstance.Name + "-pod",
		},
		Labels: map[string]string{
			"app": labInstance.Name + "-" + node.Name + "-remote-access",
		},
	}
	podSpec := &corev1.PodSpec{}
	err := yaml.Unmarshal([]byte(node.RenderedNodeSpec), podSpec)
	if err != nil {
		log.Error(err, "Failed to unmarshal node spec")
	}
	pod := &corev1.Pod{
		ObjectMeta: metadata,
		Spec:       *podSpec,
	}
	return pod
}

func MapTemplateToVM(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) *kubevirtv1.VirtualMachine {
	log := log.FromContext(context.Background())
	metadata := metav1.ObjectMeta{
		Name:      labInstance.Name + "-" + node.Name,
		Namespace: labInstance.Namespace,
	}
	vmSpec := &kubevirtv1.VirtualMachineSpec{}
	err := yaml.Unmarshal([]byte(node.RenderedNodeSpec), vmSpec)
	if err != nil {
		log.Error(err, "Failed to unmarshal node spec")
	}
	networks := []kubevirtv1.Network{
		{Name: "default", NetworkSource: kubevirtv1.NetworkSource{Pod: &kubevirtv1.PodNetwork{}}},
		{Name: labInstance.Name, NetworkSource: kubevirtv1.NetworkSource{Multus: &kubevirtv1.MultusNetwork{NetworkName: labInstance.Name + "-vm"}}},
	}
	interfaces := []kubevirtv1.Interface{
		{Name: "default", InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{Bridge: &kubevirtv1.InterfaceBridge{}}},
		{Name: labInstance.Name, InterfaceBindingMethod: kubevirtv1.InterfaceBindingMethod{Bridge: &kubevirtv1.InterfaceBridge{}}},
	}
	// TODO: Hack for cloud init
	disk := kubevirtv1.Disk{
		Name: "cloudinitdisk",
		DiskDevice: kubevirtv1.DiskDevice{
			Disk: &kubevirtv1.DiskTarget{
				Bus: "virtio",
			},
		},
	}
	volume := kubevirtv1.Volume{
		Name: "cloudinitdisk",
		VolumeSource: kubevirtv1.VolumeSource{
			CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
				UserData: node.Config,
			},
		},
	}
	vmSpec.Template.Spec.Domain.Devices.Interfaces = interfaces
	vmSpec.Template.Spec.Networks = networks
	vmSpec.Template.ObjectMeta.Labels = map[string]string{"app": labInstance.Name + "-" + node.Name + "-remote-access"}
	vmSpec.Template.Spec.Volumes = append(vmSpec.Template.Spec.Volumes, volume)
	vmSpec.Template.Spec.Domain.Devices.Disks = append(vmSpec.Template.Spec.Domain.Devices.Disks, disk)
	log.Info("VM Spec", "Spec", vmSpec)
	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metadata,
		Spec:       *vmSpec,
	}
	return vm
}

func CreateIngress(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes, kind string) *networkingv1.Ingress {
	name := labInstance.Name + "-" + node.Name
	ingressName := name + "-ingress"
	metadata := metav1.ObjectMeta{
		Name:      ingressName,
		Namespace: labInstance.Namespace,
		Annotations: map[string]string{
			"nginx.ingress.kubernetes.io/rewrite-target": "/?arg=" + kind + "&arg=" + name + "&arg=bash",
		},
	}
	className := "nginx"
	ingress := &networkingv1.Ingress{
		ObjectMeta: metadata,
		Spec: networkingv1.IngressSpec{
			IngressClassName: &className,
			Rules: []networkingv1.IngressRule{
				{Host: ingressName + "." + labInstance.Spec.DNSAddress,
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

func CreatePod(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) *corev1.Pod {
	pod := &corev1.Pod{}

	if node == nil {
		pod.ObjectMeta = metav1.ObjectMeta{Namespace: labInstance.Namespace}
		pod.ObjectMeta.Name = labInstance.Name + "-ttyd-pod"
		pod.ObjectMeta.Labels = map[string]string{"app": labInstance.Name + "-ttyd-service"}
		pod.Spec = corev1.PodSpec{
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
		}
	} else {
		pod = MapTemplateToPod(labInstance, node)
	}
	return pod
}

func CreateService(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) *corev1.Service {
	var serviceName string
	ports := []corev1.ServicePort{}
	serviceType := corev1.ServiceTypeLoadBalancer

	if node == nil {
		serviceName = fmt.Sprintf("%s-%s", labInstance.Name, "ttyd-service")
		ports = append(ports, corev1.ServicePort{
			Name:       "ttyd",
			Port:       7681,
			TargetPort: intstr.FromInt(7681),
		})
		serviceType = corev1.ServiceTypeClusterIP
	} else {
		serviceName = fmt.Sprintf("%s-%s-%s", labInstance.Name, node.Name, "remote-access")
		for _, port := range node.Ports {
			ports = append(ports, corev1.ServicePort{
				Name:       port.Name,
				Port:       port.Port,
				TargetPort: intstr.IntOrString{IntVal: port.Port},
				Protocol:   port.Protocol,
			})
		}
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: labInstance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": serviceName},
			Ports:    ports,
			Type:     serviceType,
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

func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbv1alpha1.LabInstance{}).
		Owns(&corev1.Pod{}).
		Owns(&kubevirtv1.VirtualMachine{}).
		Complete(r)
}
