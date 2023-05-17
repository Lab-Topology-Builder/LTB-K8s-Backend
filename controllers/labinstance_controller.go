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
	"strings"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	// "k8s.io/apimachinery/pkg/api/resource"
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

	node := &ltbv1alpha1.LabInstanceNodes{}

	// Reconcile TTYD Service Account
	_, retValue = ReconcileResource(r, labInstance, &corev1.ServiceAccount{}, node, labInstance.Name+"-ttyd-svcacc")
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	// Reconcile TTYD Role
	_, retValue = ReconcileResource(r, labInstance, &rbacv1.Role{}, node, labInstance.Name+"-ttyd-role")
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	// Reconcile TTYD Role Binding
	_, retValue = ReconcileResource(r, labInstance, &rbacv1.RoleBinding{}, node, labInstance.Name+"-ttyd-rolebind")
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	// Reconcile TTYD Service
	_, retValue = ReconcileResource(r, labInstance, &corev1.Service{}, node, labInstance.Name+"-ttyd-service")
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	// Reconcile TTYD Pod
	_, retValue = ReconcileResource(r, labInstance, &corev1.Pod{}, node, labInstance.Name+"-ttyd-pod")
	if retValue.shouldReturn {
		return retValue.result, retValue.err
	}

	nodes := labTemplate.Spec.Nodes
	pods := []*corev1.Pod{}
	vms := []*kubevirtv1.VirtualMachine{}
	for _, node := range nodes {
		if node.Image.Kind == "vm" {
			vm, retValue := ReconcileResource(r, labInstance, &kubevirtv1.VirtualMachine{}, &node, labInstance.Name+"-"+node.Name)
			if retValue.shouldReturn {
				return retValue.result, retValue.err
			}
			vms = append(vms, vm.(*kubevirtv1.VirtualMachine))
		} else {
			pod, retValue := ReconcileResource(r, labInstance, &corev1.Pod{}, &node, labInstance.Name+"-"+node.Name)
			if retValue.shouldReturn {
				return retValue.result, retValue.err
			}
			pods = append(pods, pod.(*corev1.Pod))
		}
		kind := node.Image.Kind
		if kind == "" {
			kind = "pod"
		}

		// Reconcile Remote Access Service
		if len(node.Ports) > 0 {
			_, retValue = ReconcileResource(r, labInstance, &corev1.Service{}, &node, labInstance.Name+"-"+node.Name+"-remote-access")
			if retValue.shouldReturn {
				return retValue.result, retValue.err
			}
		}

		// Reconcile Ingress
		_, retValue = ReconcileResource(r, labInstance, &networkingv1.Ingress{}, &node, labInstance.Name+"-"+node.Name+"-ingress")
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

func ReconcileResource(r *LabInstanceReconciler, labInstance *ltbv1alpha1.LabInstance, resource client.Object, node *ltbv1alpha1.LabInstanceNodes, resourceName string) (client.Object, ReturnToReconciler) {
	ctx := context.Context(context.Background())
	log := log.FromContext(ctx)
	retValue := ReturnToReconciler{shouldReturn: true, result: ctrl.Result{}, err: nil}
	resourceExists, err := ResourceExists(r, resource, resourceName, labInstance.Namespace)
	if err == nil && !resourceExists {
		createdResource := CreateResource(labInstance, node, resourceName, resource)
		log.Info("Creating a new resource", "resource.Namespace", labInstance.Namespace, "resource.Name", reflect.ValueOf(createdResource).Elem().FieldByName("Name"))
		ctrl.SetControllerReference(labInstance, createdResource, r.Scheme)

		err = r.Create(ctx, createdResource)
		if err != nil {
			retValue.err = err
			log.Error(err, "Failed to create new resource", "resource.Namespace", labInstance.Namespace, "resource.Name", reflect.ValueOf(createdResource).Elem().FieldByName("Name"))
			return nil, retValue
		}
		retValue.result = ctrl.Result{Requeue: true}
		return createdResource, retValue
	} else if err != nil {
		retValue.err = err
		log.Error(err, "Failed to get resource")
		return resource, retValue
	}
	retValue.shouldReturn = false
	return resource, retValue
}

// TODO: Remove return value use pointers
func CreateResource(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes, resourceName string, resource client.Object) client.Object {
	var kind string
	ctx := context.Context(context.Background())
	log := log.FromContext(ctx)
	if node != nil && node.Image.Kind != "" {
		kind = node.Image.Kind
	} else {
		kind = "pod"
	}
	switch reflect.TypeOf(resource).Elem().Name() {
	case "Pod":
		if node.Name != "" {
			return MapTemplateToPod(labInstance, node)
		} else {
			pod, _ := CreateTtydPodAndService(labInstance)
			resource = pod
			return pod
		}
	case "VirtualMachine":
		return MapTemplateToVM(labInstance, node)
	case "Service":
		if strings.Contains(resourceName, "ttyd") {
			_, service := CreateTtydPodAndService(labInstance)
			return service
		} else {
			return CreateService(node, resourceName, labInstance.Namespace)
		}
	case "Ingress":
		return CreateIngress(labInstance, kind, labInstance.Name+"-"+node.Name)
	case "Role":
		_, role, _ := CreateSvcAccRoleRoleBind(labInstance)
		return role
	case "ServiceAccount":
		svcAcc, _, _ := CreateSvcAccRoleRoleBind(labInstance)
		return svcAcc
	case "RoleBinding":
		_, _, roleBind := CreateSvcAccRoleRoleBind(labInstance)
		return roleBind
	default:
		log.Info("Resource type not supported", "ResourceKind", resource.GetObjectKind().GroupVersionKind().Kind)
		return nil
	}

}

func ResourceExists(r *LabInstanceReconciler, resource client.Object, resourceName string, nameSpace string) (bool, error) {
	ctx := context.Context(context.Background())
	err := r.Get(ctx, types.NamespacedName{Name: resourceName, Namespace: nameSpace}, resource)
	if errors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (r *LabInstanceReconciler) GetLabTemplate(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, labTemplate *ltbv1alpha1.LabTemplate) ReturnToReconciler {
	err := r.Get(ctx, types.NamespacedName{Name: labInstance.Spec.LabTemplateReference, Namespace: labInstance.Namespace}, labTemplate)
	returnValue := ErrorMsg(ctx, err, "LabTemplate")
	return returnValue
}

func MapTemplateToPod(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) *corev1.Pod {
	ports := []corev1.ContainerPort{}
	for _, port := range node.Ports {
		ports = append(ports, corev1.ContainerPort{
			ContainerPort: port.Port,
		})
	}
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
	pod := &corev1.Pod{
		ObjectMeta: metadata,

		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    node.Name,
					Image:   node.Image.Type + ":" + node.Image.Version,
					Command: []string{"/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"},
					Ports:   ports,
				},
			},
		},
	}
	return pod
}

func MapTemplateToVM(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) *kubevirtv1.VirtualMachine {
	metadata := metav1.ObjectMeta{
		Name:      labInstance.Name + "-" + node.Name,
		Namespace: labInstance.Namespace,
		Labels: map[string]string{
			"app": labInstance.Name + "-" + node.Name + "-remote-access",
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
		{Name: labInstance.Name, NetworkSource: kubevirtv1.NetworkSource{Multus: &kubevirtv1.MultusNetwork{NetworkName: labInstance.Name + "-vm"}}},
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
		Name:      labInstance.Name + "-ttyd-pod",
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

func CreateService(node *ltbv1alpha1.LabInstanceNodes, serviceName string, nameSpace string) *corev1.Service {
	ports := []corev1.ServicePort{}
	for _, port := range node.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: intstr.IntOrString{IntVal: port.Port},
		})
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: nameSpace,
		},
		Spec: corev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				"app": serviceName,
			},
			Type: corev1.ServiceTypeLoadBalancer,
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
		Name:      labInstance.Name + "-ttyd-pod",
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

func CreateService(node *ltbv1alpha1.LabInstanceNodes, serviceName string, nameSpace string) *corev1.Service {
	ports := []corev1.ServicePort{}
	for _, port := range node.Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       port.Name,
			Port:       port.Port,
			TargetPort: intstr.IntOrString{IntVal: port.Port},
		})
	}
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: nameSpace,
		},
		Spec: corev1.ServiceSpec{
			Ports: ports,
			Selector: map[string]string{
				"app": serviceName,
			},
			Type: corev1.ServiceTypeLoadBalancer,
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

// This function could be moved to utils
func ErrorMsg(ctx context.Context, err error, resource string) ReturnToReconciler {
	log := log.FromContext(ctx)
	returnValue := ReturnToReconciler{shouldReturn: false, result: ctrl.Result{}, err: nil}
	if err != nil && errors.IsNotFound(err) {
		log.Info(resource + " resource not found.")
		returnValue.shouldReturn = true
		return returnValue
	} else if err != nil {
		returnValue.shouldReturn = true
		returnValue.err = err
		log.Error(err, "Failed to get "+resource)
		return returnValue
	}
	return returnValue
}

func (r *LabInstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbv1alpha1.LabInstance{}).
		Owns(&corev1.Pod{}).
		Owns(&kubevirtv1.VirtualMachine{}).
		Complete(r)
}
