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

// TODO move to util
type ReturnToReconciler struct {
	ShouldReturn bool
	Result       ctrl.Result
	Err          error
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
			return ctrl.Result{}, err
		}
		log.Error(err, "Failed to get LabInstance")
		return ctrl.Result{}, err
	}

	labTemplate := &ltbv1alpha1.LabTemplate{}
	retValue := r.GetLabTemplate(ctx, labInstance, labTemplate)
	if retValue.ShouldReturn {
		return retValue.Result, retValue.Err
	}

	// Reconcile Network
	retValue = r.ReconcileNetwork(ctx, labInstance)
	if retValue.ShouldReturn {
		return retValue.Result, retValue.Err
	}

	// Reconcile TTYD Service Account
	_, retValue = ReconcileResource(r, labInstance, &corev1.ServiceAccount{}, nil, labInstance.Name+"-ttyd-svcacc")
	if retValue.ShouldReturn {
		return retValue.Result, retValue.Err
	}

	// Reconcile TTYD Role
	_, retValue = ReconcileResource(r, labInstance, &rbacv1.Role{}, nil, labInstance.Name+"-ttyd-role")
	if retValue.ShouldReturn {
		return retValue.Result, retValue.Err
	}

	// Reconcile TTYD Role Binding
	_, retValue = ReconcileResource(r, labInstance, &rbacv1.RoleBinding{}, nil, labInstance.Name+"-ttyd-rolebind")
	if retValue.ShouldReturn {
		return retValue.Result, retValue.Err
	}

	// Reconcile TTYD Service
	_, retValue = ReconcileResource(r, labInstance, &corev1.Service{}, nil, labInstance.Name+"-ttyd-service")
	if retValue.ShouldReturn {
		return retValue.Result, retValue.Err
	}

	// Reconcile TTYD Pod
	_, retValue = ReconcileResource(r, labInstance, &corev1.Pod{}, nil, labInstance.Name+"-ttyd-pod")
	if retValue.ShouldReturn {
		return retValue.Result, retValue.Err
	}

	nodes := labTemplate.Spec.Nodes
	pods := []*corev1.Pod{}
	vms := []*kubevirtv1.VirtualMachine{}
	for _, node := range nodes {
		nodeType := &ltbv1alpha1.NodeType{}
		retValue = r.GetNodeType(ctx, &node.NodeTypeRef, nodeType)
		if retValue.ShouldReturn {
			return retValue.Result, retValue.Err
		}
		if nodeType.Spec.Kind == "vm" {
			vm, retValue := ReconcileResource(r, labInstance, &kubevirtv1.VirtualMachine{}, &node, labInstance.Name+"-"+node.Name)
			if retValue.ShouldReturn {
				return retValue.Result, retValue.Err
			}
			vms = append(vms, vm.(*kubevirtv1.VirtualMachine))
		} else {
			pod, retValue := ReconcileResource(r, labInstance, &corev1.Pod{}, &node, labInstance.Name+"-"+node.Name)
			if retValue.ShouldReturn {
				return retValue.Result, retValue.Err
			}
			pods = append(pods, pod.(*corev1.Pod))
		}

		// Reconcile Remote Access Service
		if len(node.Ports) > 0 {
			_, retValue = ReconcileResource(r, labInstance, &corev1.Service{}, &node, labInstance.Name+"-"+node.Name+"-remote-access")
			if retValue.ShouldReturn {
				return retValue.Result, retValue.Err
			}
		}

		// Reconcile Ingress
		_, retValue = ReconcileResource(r, labInstance, &networkingv1.Ingress{}, &node, labInstance.Name+"-"+node.Name+"-ingress")
		if retValue.ShouldReturn {
			return retValue.Result, retValue.Err
		}

	}

	UpdateLabInstanceStatus(pods, vms, labInstance)

	err = r.Status().Update(ctx, labInstance)
	if err != nil {
		log.Error(err, "Failed to update LabInstance status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// TODO: Refactor to use ReconcileResource or at least to use the same pattern/splitting
func (r *LabInstanceReconciler) ReconcileNetwork(ctx context.Context, labInstance *ltbv1alpha1.LabInstance) ReturnToReconciler {
	log := log.FromContext(ctx)
	retValue := ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{}, Err: nil}
	if labInstance == nil {
		retValue.Err = errors.NewBadRequest("labInstance is nil")
		return retValue
	}
	podNetworkDefinitionName := labInstance.Name + "-pod"
	vmNetworkDefinitionName := labInstance.Name + "-vm"
	networkdefinitionNames := []string{podNetworkDefinitionName, vmNetworkDefinitionName}
	for _, networkDefinitionName := range networkdefinitionNames {
		foundNetworkAttachmentDefinition := &network.NetworkAttachmentDefinition{}
		err := r.Get(ctx, types.NamespacedName{Name: networkDefinitionName, Namespace: labInstance.Namespace}, foundNetworkAttachmentDefinition)
		if errors.IsNotFound(err) {
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
				retValue.Err = err
				log.Error(err, "Failed to create NetworkAttachmentDefinition")
				return retValue
			}
			retValue.Result = ctrl.Result{Requeue: true}
			return retValue
		}
		if err != nil {
			retValue.Err = err
			log.Error(err, "Failed to get NetworkAttachmentDefinition")
			return retValue
		}
	}
	retValue.ShouldReturn = false
	return retValue
}

func ReconcileResource(r *LabInstanceReconciler, labInstance *ltbv1alpha1.LabInstance, resource client.Object, node *ltbv1alpha1.LabInstanceNodes, resourceName string) (client.Object, ReturnToReconciler) {
	ctx := context.Context(context.Background())
	log := log.FromContext(ctx)
	retValue := ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{}, Err: nil}
	if labInstance == nil {
		retValue.Err = errors.NewBadRequest("labInstance is nil")
		return nil, retValue
	}
	resourceExists, err := r.ResourceExists(resource, resourceName, labInstance.Namespace)
	if err != nil && !resourceExists {
		createdResource, err := CreateResource(labInstance, node, resource)
		if err != nil {
			retValue.Err = err
			log.Error(err, "Failed to create new resource", "resource.Namespace", labInstance.Namespace, "resource.Name", resourceName)
			return nil, retValue
		}
		log.Info("Creating a new resource", "resource.Namespace", labInstance.Namespace, "resource.Name", reflect.ValueOf(createdResource).Elem().FieldByName("Name"))
		ctrl.SetControllerReference(labInstance, createdResource, r.Scheme)

		err = r.Create(ctx, createdResource)
		if err != nil {
			retValue.Err = err
			log.Error(err, "Failed to create new resource", "resource.Namespace", labInstance.Namespace, "resource.Name", reflect.ValueOf(createdResource).Elem().FieldByName("Name"))
			return nil, retValue
		}
		retValue.Result = ctrl.Result{Requeue: true}
		return createdResource, retValue
	} else if err != nil {
		retValue.Err = err
		log.Error(err, "Failed to get resource")
		return resource, retValue
	}
	retValue.ShouldReturn = false
	return resource, retValue
}

// TODO: Remove return value use pointers, maybe remove resourceName and add flag for ttyd
func CreateResource(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes, resource client.Object) (client.Object, error) {
	ctx := context.Context(context.Background())
	log := log.FromContext(ctx)
	switch reflect.TypeOf(resource).Elem().Name() {
	case "Pod":
		return CreatePod(labInstance, node)
	case "VirtualMachine":
		return MapTemplateToVM(labInstance, node)
	case "Service":
		return CreateService(labInstance, node)
	case "Ingress":
		return CreateIngress(labInstance, node)
	case "Role":
		_, role, _, err := CreateSvcAccRoleRoleBind(labInstance)
		return role, err
	case "ServiceAccount":
		svcAcc, _, _, err := CreateSvcAccRoleRoleBind(labInstance)
		return svcAcc, err
	case "RoleBinding":
		_, _, roleBind, err := CreateSvcAccRoleRoleBind(labInstance)
		return roleBind, err
	default:
		log.Error(fmt.Errorf("resource type not supported"), "Unsupported", "ResourceKind", reflect.TypeOf(resource).Elem().Name())
		return nil, errors.NewBadRequest(fmt.Sprintf("Resource type not supported: %s", reflect.TypeOf(resource).Elem().Name()))
	}

}

// TODO: Refactor probably does not need to be a function as it is only called once
func (r *LabInstanceReconciler) ResourceExists(resource client.Object, resourceName string, nameSpace string) (bool, error) {
	ctx := context.Context(context.Background())
	err := r.Get(ctx, types.NamespacedName{Name: resourceName, Namespace: nameSpace}, resource)
	if errors.IsNotFound(err) {
		return false, err
	} else if err != nil {
		return true, err
	}
	return true, nil
}

func (r *LabInstanceReconciler) GetLabTemplate(ctx context.Context, labInstance *ltbv1alpha1.LabInstance, labTemplate *ltbv1alpha1.LabTemplate) ReturnToReconciler {
	err := r.Get(ctx, types.NamespacedName{Name: labInstance.Spec.LabTemplateReference, Namespace: labInstance.Namespace}, labTemplate)
	returnValue := ErrorMsg(ctx, err, "LabTemplate")
	return returnValue
}

func (r *LabInstanceReconciler) GetNodeType(ctx context.Context, nodeTypeRef *ltbv1alpha1.NodeTypeRef, nodeType *ltbv1alpha1.NodeType) ReturnToReconciler {
	err := r.Get(ctx, types.NamespacedName{Name: nodeTypeRef.Type, Namespace: nodeType.Namespace}, nodeType)
	returnValue := ErrorMsg(ctx, err, "NodeType")
	return returnValue
}

func MapTemplateToPod(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (*corev1.Pod, error) {
	log := log.FromContext(context.Background())
	if node == nil {
		return nil, errors.NewBadRequest("Node is nil")
	}
	if labInstance == nil {
		return nil, errors.NewBadRequest("LabInstance is nil")
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
	podSpec := &corev1.PodSpec{}
	err := yaml.Unmarshal([]byte(node.RenderedNodeSpec), podSpec)
	if err != nil {
		log.Error(err, "Failed to unmarshal node spec")
		return nil, err
	}
	pod := &corev1.Pod{
		ObjectMeta: metadata,
		Spec:       *podSpec,
	}
	return pod, nil
}

func MapTemplateToVM(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (*kubevirtv1.VirtualMachine, error) {
	log := log.FromContext(context.Background())
	if node == nil {
		return nil, errors.NewBadRequest("Node is nil")
	}
	if labInstance == nil {
		return nil, errors.NewBadRequest("LabInstance is nil")
	}
	metadata := metav1.ObjectMeta{
		Name:      labInstance.Name + "-" + node.Name,
		Namespace: labInstance.Namespace,
	}
	vmSpec := &kubevirtv1.VirtualMachineSpec{}
	err := yaml.Unmarshal([]byte(node.RenderedNodeSpec), vmSpec)
	if err != nil {
		log.Error(err, "Failed to unmarshal node spec")
		return nil, err
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
	vm := &kubevirtv1.VirtualMachine{
		ObjectMeta: metadata,
		Spec:       *vmSpec,
	}
	return vm, nil
}

func CreateIngress(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (*networkingv1.Ingress, error) {
	// TODO: hack to determine if node is a vm or pod, need to improve
	var resourceType string
	if labInstance == nil {
		return nil, errors.NewBadRequest("LabInstance is nil")
	} else if node == nil {
		return nil, errors.NewBadRequest("Node is nil")
	}
	if node != nil && strings.Contains(node.RenderedNodeSpec, "template:") {
		resourceType = "vm"
	} else {
		resourceType = "pod"
	}
	name := labInstance.Name + "-" + node.Name
	ingressName := name + "-ingress"
	metadata := metav1.ObjectMeta{
		Name:      ingressName,
		Namespace: labInstance.Namespace,
		Annotations: map[string]string{
			"nginx.ingress.kubernetes.io/rewrite-target": "/?arg=" + resourceType + "&arg=" + name + "&arg=bash",
		},
	}
	className := "nginx"
	// TODO: ingress dns address should be configurable
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
	return ingress, nil
}

func CreatePod(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	var err error
	if labInstance == nil {
		return nil, errors.NewBadRequest("LabInstance is nil")
	}

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
		err = nil
	} else {
		pod, err = MapTemplateToPod(labInstance, node)
	}
	return pod, err
}

func CreateService(labInstance *ltbv1alpha1.LabInstance, node *ltbv1alpha1.LabInstanceNodes) (*corev1.Service, error) {
	var serviceName string
	ports := []corev1.ServicePort{}
	serviceType := corev1.ServiceTypeLoadBalancer
	if labInstance == nil {
		return nil, errors.NewBadRequest("LabInstance is nil")
	}

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
	return service, nil
}

func CreateSvcAccRoleRoleBind(labInstance *ltbv1alpha1.LabInstance) (*corev1.ServiceAccount, *rbacv1.Role, *rbacv1.RoleBinding, error) {
	if labInstance == nil {
		return nil, nil, nil, errors.NewBadRequest("LabInstance is nil")
	}
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

	return serviceAccount, role, roleBinding, nil

}

func UpdateLabInstanceStatus(pods []*corev1.Pod, vms []*kubevirtv1.VirtualMachine, labInstance *ltbv1alpha1.LabInstance) error {
	var podStatus corev1.PodPhase
	var vmStatus kubevirtv1.VirtualMachinePrintableStatus
	var numVMsRunning, numPodsRunning int
	if pods == nil && vms == nil {
		return errors.NewBadRequest("No resources found")
	}
	if labInstance == nil {
		return errors.NewBadRequest("LabInstance is nil")
	}
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
	return nil
}

// TODO: move this function to utils or remove it
func ErrorMsg(ctx context.Context, err error, resource string) ReturnToReconciler {
	log := log.FromContext(ctx)
	returnValue := ReturnToReconciler{ShouldReturn: false, Result: ctrl.Result{}, Err: nil}
	if errors.IsNotFound(err) {
		log.Info("Resource not found.")
		returnValue.ShouldReturn = true
		returnValue.Err = err
		return returnValue
	}
	if err != nil {
		returnValue.ShouldReturn = true
		returnValue.Err = err
		log.Error(err, "Resource: "+resource)
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
