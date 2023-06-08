package controllers

import (
	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	network "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type fields struct {
	Client client.Client
	Scheme *runtime.Scheme
}

const namespace = "test-namespace"

var (
	r                                                                                         *LabInstanceReconciler
	testLabInstance                                                                           *ltbv1alpha1.LabInstance
	testLabTemplate                                                                           *ltbv1alpha1.LabTemplate
	testNodeTypeVM, testNodeTypePod                                                           *ltbv1alpha1.NodeType
	err                                                                                       error
	normalPodNode, normalVMNode, nodeUndefinedNodeType, vmYAMLProblemNode, podYAMLProblemNode *ltbv1alpha1.LabInstanceNodes
	fakeClient                                                                                client.Client
	testPod, testNodePod, testTtydPod                                                         *corev1.Pod
	field                                                                                     fields
	testVM, testNodeVM                                                                        *kubevirtv1.VirtualMachine
	testPodIngress, testVMIngress                                                             *networkingv1.Ingress
	testService, testTtydService                                                              *corev1.Service
	testRole                                                                                  *rbacv1.Role
	testRoleBinding                                                                           *rbacv1.RoleBinding
	testServiceAccount                                                                        *corev1.ServiceAccount
	testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition                     *network.NetworkAttachmentDefinition
	req                                                                                       ctrl.Request
)

func initialize() {
	testLabInstance = &ltbv1alpha1.LabInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-labinstance",
			Namespace: namespace,
		},
		Spec: ltbv1alpha1.LabInstanceSpec{
			LabTemplateReference: "test-labtemplate",
		},
		Status: ltbv1alpha1.LabInstanceStatus{
			Status: "Running",
		},
	}

	testNodeTypeVM = &ltbv1alpha1.NodeType{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testNodeVM",
			Namespace: "",
		},
		Spec: ltbv1alpha1.NodeTypeSpec{
			Kind: "vm",
			NodeSpec: `
    running: true
    template:
      spec:
        domain:
          resources:
            requests:
              memory: 4096M
          cpu:
            cores: 2
          devices:
            disks:
              - name: containerdisk
                disk:
                  bus: virtio
        terminationGracePeriodSeconds: 0
        volumes:
          - name: containerdisk
            containerDisk:
              image: quay.io/containerdisks/ubuntu:22.04
`,
		},
	}

	testNodeTypePod = &ltbv1alpha1.NodeType{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testNodePod",
			Namespace: "",
		},
		Spec: ltbv1alpha1.NodeTypeSpec{
			Kind: "pod",
			NodeSpec: `
	containers:
	  - name: {{ .Name }}
	    image: {{ .NodeTypeRef.Image}}:{{ .NodeTypeRef.Version }}
	    command: ["/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"]
	    ports:
	      {{- range $index, $port := .Ports }}
	      - name: {{ $port.Name }}
	        containerPort: {{ $port.Port }}
	        protocol: {{ $port.Protocol }}
	      {{- end }}
`,
		},
	}

	normalVMNode = &ltbv1alpha1.LabInstanceNodes{
		Name: "test-node-0",
		NodeTypeRef: ltbv1alpha1.NodeTypeRef{
			Type:    testNodeTypeVM.Name,
			Image:   "ubuntu",
			Version: "22.04",
		},
		Ports: []ltbv1alpha1.Port{
			{
				Name:     "test-ssh-port",
				Protocol: "TCP",
				Port:     22,
			},
		},
		RenderedNodeSpec: `
running: true
template:
  spec:
    domain:
      resources:
        requests:
          memory: 4096M
      cpu:
        cores: 2
      devices:
        disks:
          - name: containerdisk
            disk:
              bus: virtio
    terminationGracePeriodSeconds: 0
    volumes:
      - name: containerdisk
        containerDisk:
          image: quay.io/containerdisks/ubuntu:22.04
`,
	}
	vmYAMLProblemNode = &ltbv1alpha1.LabInstanceNodes{
		Name: "test-node-3",
		NodeTypeRef: ltbv1alpha1.NodeTypeRef{
			Type:    testNodeTypeVM.Name,
			Image:   "ubuntu",
			Version: "22.04",
		},
		RenderedNodeSpec: `
running: true
template:
  spec:
    domain:
      resources:
        requests:
          memory: 4096M
      cpu:
        cores: 2
      devices:
        disks:
          - name: containerdisk
            disk:
              bus: virtio
    terminationGracePeriodSeconds: 0
    volumes:
      - name: containerdisk
        containerDisk:
          image: quay.io/containerdisks/ubuntu:22.04
	`,
	}
	normalPodNode = &ltbv1alpha1.LabInstanceNodes{
		Name: "test-node-1",
		NodeTypeRef: ltbv1alpha1.NodeTypeRef{
			Type:    testNodeTypePod.Name,
			Image:   "ubuntu",
			Version: "20.04",
		},
		RenderedNodeSpec: `
    containers:
      - name: testnode
        image: ubuntu:22.04
        command: ["/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"]
        ports:
          - name: testsshport
            containerPort: 22
            protocol: tcp
`,
	}
	nodeUndefinedNodeType = &ltbv1alpha1.LabInstanceNodes{
		Name: "test-node-2",
		NodeTypeRef: ltbv1alpha1.NodeTypeRef{
			Type:    "test",
			Image:   "ubuntu",
			Version: "20.04",
		},
		Ports: []ltbv1alpha1.Port{
			{
				Name:     "test-ssh-port",
				Protocol: "TCP",
				Port:     22,
			},
		},
	}
	podYAMLProblemNode = &ltbv1alpha1.LabInstanceNodes{
		Name: "test-node-4",
		NodeTypeRef: ltbv1alpha1.NodeTypeRef{
			Type:    "testNodePod",
			Image:   "ubuntu",
			Version: "20.04",
		},
		RenderedNodeSpec: `
					containers:
					- name: testnode
					  image: ubuntu:22.04
					  command: ["/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"]
					  ports:
						- name: testsshport
						  containerPort: 22
						  protocol: tcp
			  	`,
	}

	testLabTemplate = &ltbv1alpha1.LabTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-labtemplate",
			Namespace: namespace,
		},
		Spec: ltbv1alpha1.LabTemplateSpec{
			Nodes: []ltbv1alpha1.LabInstanceNodes{
				*normalVMNode,
				*normalPodNode,
			},
		},
	}

	testPod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-" + normalPodNode.Name,
			Namespace: namespace,
			Annotations: map[string]string{
				"k8s.v1.cni.cncf.io/networks": testLabInstance.Name + "-pod",
			},
			Labels: map[string]string{
				"app": testLabInstance.Name + "-" + normalPodNode.Name + "-remote-access",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}

	testVM = &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-" + normalVMNode.Name,
			Namespace: testLabInstance.Namespace,
		},
		Spec: kubevirtv1.VirtualMachineSpec{
			Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": testLabInstance.Name + "-" + normalVMNode.Name + "-remote-access",
					},
				},
				Spec: kubevirtv1.VirtualMachineInstanceSpec{
					Volumes: []kubevirtv1.Volume{
						{
							Name: "cloudinitdisk",
							VolumeSource: kubevirtv1.VolumeSource{
								CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{
									UserData: normalVMNode.Config,
								},
							},
						},
					},
				},
			},
		},
		Status: kubevirtv1.VirtualMachineStatus{
			Ready:           true,
			PrintableStatus: "VM Ready",
		},
	}

	testNodeVM = &kubevirtv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-" + normalVMNode.Name + "-2",
			Namespace: testLabInstance.Namespace,
		},
		Status: kubevirtv1.VirtualMachineStatus{
			Ready: false,
		},
	}

	testNodePod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-" + nodeUndefinedNodeType.Name,
			Namespace: namespace,
			Annotations: map[string]string{
				"k8s.v1.cni.cncf.io/networks": testLabInstance.Name + "-pod",
			},
			Labels: map[string]string{
				"app": testLabInstance.Name + "-" + nodeUndefinedNodeType.Name + "-remote-access",
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPending,
		},
	}

	testTtydPod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-ttyd-pod",
			Namespace: namespace,
			Labels: map[string]string{
				"app": testLabInstance.Name + "-ttyd-service",
			},
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: testLabInstance.Name + "-ttyd-svcacc",
			Containers: []corev1.Container{
				{
					Name:  testLabInstance.Name + "-ttyd-container",
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

	testTtydService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-ttyd-service",
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": testLabInstance.Name + "-ttyd-service",
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "ttyd",
					Port:       7681,
					TargetPort: intstr.FromInt(7681),
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		},
	}

	testService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-" + normalVMNode.Name + "-remote-access",
			Namespace: namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{
				"app": testLabInstance.Name + "-" + normalVMNode.Name + "-remote-access",
			},
			Ports: []corev1.ServicePort{
				{
					Name:     "test-ssh-port",
					Port:     22,
					Protocol: "TCP",
				},
			},
			Type: corev1.ServiceTypeLoadBalancer,
		},
	}

	testPodIngress = &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-" + normalPodNode.Name + "-ingress",
			Namespace: namespace,
			Annotations: map[string]string{
				"nginx.ingress.kubernetes.io/rewrite-target": "/?arg=pod&arg=" + testLabInstance.Name + "-" + normalPodNode.Name + "&arg=bash",
			},
		},
	}

	testVMIngress = &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-" + normalVMNode.Name + "-ingress",
			Namespace: namespace,
		},
	}

	testServiceAccount = &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-ttyd-svcacc",
			Namespace: namespace,
		},
	}

	testRole = &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-ttyd-role",
			Namespace: namespace,
		},
	}

	testRoleBinding = &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-ttyd-rolebind",
			Namespace: namespace,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      testLabInstance.Name + "-ttyd-svcacc",
				Namespace: namespace,
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     testLabInstance.Name + "-ttyd-role",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}

	testPodNetworkAttachmentDefinition = &network.NetworkAttachmentDefinition{}
	testPodNetworkAttachmentDefinition.Name = testLabInstance.Name + "-pod"
	testPodNetworkAttachmentDefinition.Namespace = testLabInstance.Namespace
	testPodNetworkAttachmentDefinition.Spec.Config = `{
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
	testVMNetworkAttachmentDefinition = &network.NetworkAttachmentDefinition{}
	testVMNetworkAttachmentDefinition.Name = testLabInstance.Name + "-vm"
	testVMNetworkAttachmentDefinition.Namespace = testLabInstance.Namespace
	testVMNetworkAttachmentDefinition.Spec.Config = `{
					"cniVersion": "0.3.1",
					"name": "mynet",
					"type": "bridge",
					"bridge": "mynet0",
					"ipam": {}
				}`

	err = ltbv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}
	err = kubevirtv1.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}

	err = network.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}

	// fakeClient = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPod).Build()
	// r = &LabInstanceReconciler{Client: fakeClient, Scheme: scheme.Scheme}
	field = fields{fakeClient, scheme.Scheme}
}
