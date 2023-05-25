package controllers_test

import (
	"context"
	"errors"
	"os"
	"testing"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	. "github.com/Lab-Topology-Builder/LTB-K8s-Backend/controllers"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// var _ = Describe("LabInstance Controller", func() {
var (
	ctx                             context.Context
	r                               *LabInstanceReconciler
	testLabInstance                 *ltbv1alpha1.LabInstance
	testLabTemplate                 *ltbv1alpha1.LabTemplate
	testNodeTypeVM, testNodeTypePod *ltbv1alpha1.NodeType
	err                             error
	podNode, vmNode, testNode       *ltbv1alpha1.LabInstanceNodes
	running                         bool
	returnValue                     ReturnToReconciler
	expectedReturnValue             ReturnToReconciler
)

const namespace = "test-namespace"

//k8sClient := K8sClient.GetClient()

func TestMain(m *testing.M) {
	initialize()
	code := m.Run()
	os.Exit(code)
}

func initialize() {
	ctx = context.Background()
	testLabInstance = &ltbv1alpha1.LabInstance{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-labinstance",
			Namespace: namespace,
		},
		Spec: ltbv1alpha1.LabInstanceSpec{
			LabTemplateReference: "test-labtemplate",
		},
	}

	nodeSpecYAMLVM := `
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
	`

	vmConfig := `
	  #cloud-config
      password: ubuntu
      chpasswd: { expire: False }
      ssh_pwauth: True
      packages:
        - qemu-guest-agent
      runcmd:
        - [ systemctl, start, qemu-guest-agent ]
	`

	// nodeSpecYAMLPod := `
	// containers:
	//   - name: {{ .Name }}
	//     image: {{ .NodeTypeRef.Image}}:{{ .NodeTypeRef.Version }}
	//     command: ["/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"]
	//     ports:
	//       {{- range $index, $port := .Ports }}
	//       - name: {{ $port.Name }}
	//         containerPort: {{ $port.Port }}
	//         protocol: {{ $port.Protocol }}
	//       {{- end }}
	// `

	testNodeTypeVM = &ltbv1alpha1.NodeType{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testNodeVM",
			Namespace: namespace,
		},
		Spec: ltbv1alpha1.NodeTypeSpec{
			Kind:     "vm",
			NodeSpec: nodeSpecYAMLVM,
		},
	}

	testNodeTypePod = &ltbv1alpha1.NodeType{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testNodePod",
			Namespace: namespace,
		},
		Spec: ltbv1alpha1.NodeTypeSpec{
			Kind: "pod",
			//			NodeSpec: nodeSpecYAMLPod,
		},
	}

	testLabTemplate = &ltbv1alpha1.LabTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-labtemplate",
			Namespace: namespace,
		},
		Spec: ltbv1alpha1.LabTemplateSpec{
			Nodes: []ltbv1alpha1.LabInstanceNodes{
				{
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
					Config: vmConfig,
				},
				{
					Name: "test-node-1",
					NodeTypeRef: ltbv1alpha1.NodeTypeRef{
						Type:    testNodeTypePod.Name,
						Image:   "ubuntu",
						Version: "20.04",
					},
				},
				{
					Name: "test-node-2",
					NodeTypeRef: ltbv1alpha1.NodeTypeRef{
						Type:    testNodeTypePod.Name,
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
				},
			},
		},
	}
	vmNode = &testLabTemplate.Spec.Nodes[0]
	podNode = &testLabTemplate.Spec.Nodes[1]
	testNode = &testLabTemplate.Spec.Nodes[2]

	// TODO: Need to check if this is the right way to add the scheme
	err = ltbv1alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}
	err = kubevirtv1.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}

	expectedReturnValue = ReturnToReconciler{ShouldReturn: false, Result: ctrl.Result{}, Err: nil}

	client := fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM).Build()
	r = &LabInstanceReconciler{Client: client, Scheme: scheme.Scheme}

}

func TestGetTemplate(t *testing.T) {
	returnValue = r.GetLabTemplate(ctx, testLabInstance, testLabTemplate)
	assert.Equal(t, expectedReturnValue, returnValue)
}

func TestGetNodeType(t *testing.T) {
	returnValue = r.GetNodeType(ctx, &podNode.NodeTypeRef, testNodeTypePod)
	assert.Equal(t, expectedReturnValue, returnValue)
}

func TestCreatePod(t *testing.T) {
	createdPod := CreatePod(testLabInstance, podNode)
	assert.Equal(t, testLabInstance.Name+"-"+podNode.Name, createdPod.Name)
	assert.Equal(t, testLabInstance.Namespace, createdPod.Namespace)
	createdTtydPod := CreatePod(testLabInstance, nil)
	assert.Equal(t, testLabInstance.Name+"-ttyd-pod", createdTtydPod.Name)
}

// TODO: Need to figure out how to test this
// func TestMapTemplateToVM(t *testing.T) {
// 	t.Log(vmNode)
// 	mappedVM := MapTemplateToVM(testLabInstance, vmNode)
// 	t.Log(mappedVM)
// 	t.Log(vmNode)
//assert.Equal(t, testLabInstance.Name+"-"+vmNode.Name, mappedVM.Name)
// }

func TestCreateIngress(t *testing.T) {
	// Pod ingress
	createdPodIngress := CreateIngress(testLabInstance, podNode)
	assert.Equal(t, testLabInstance.Name+"-"+podNode.Name+"-ingress", createdPodIngress.Name)
	assert.Equal(t, testLabInstance.Namespace, createdPodIngress.Namespace)

	// VM ingress
	createdVMIngress := CreateIngress(testLabInstance, vmNode)
	assert.Equal(t, testLabInstance.Name+"-"+vmNode.Name+"-ingress", createdVMIngress.Name)
	assert.Equal(t, testLabInstance.Namespace, createdVMIngress.Namespace)
}

func TestCreateService(t *testing.T) {
	// Pod service
	createdPodService := CreateService(testLabInstance, podNode)
	assert.Equal(t, testLabInstance.Name+"-"+podNode.Name+"-remote-access", createdPodService.Name)
	assert.Equal(t, testLabInstance.Namespace, createdPodService.Namespace)
	assert.Equal(t, "LoadBalancer", string(createdPodService.Spec.Type))
	assert.Equal(t, 0, len(createdPodService.Spec.Ports))

	// VM service
	createdVMService := CreateService(testLabInstance, vmNode)
	assert.Equal(t, testLabInstance.Name+"-"+vmNode.Name+"-remote-access", createdVMService.Name)
	assert.Equal(t, testLabInstance.Namespace, createdVMService.Namespace)
	assert.Equal(t, "LoadBalancer", string(createdVMService.Spec.Type))
	assert.Equal(t, 1, len(createdVMService.Spec.Ports))

	// TTYD service
	createdTTYDService := CreateService(testLabInstance, nil)
	assert.Equal(t, testLabInstance.Name+"-ttyd-service", createdTTYDService.Name)
	assert.Equal(t, testLabInstance.Namespace, createdTTYDService.Namespace)
	assert.Equal(t, "ClusterIP", string(createdTTYDService.Spec.Type))
	assert.Equal(t, 1, len(createdTTYDService.Spec.Ports))
	assert.Equal(t, 7681, int(createdTTYDService.Spec.Ports[0].Port))
}

func TestCreateSvcAccRoleRoleBind(t *testing.T) {
	svcAcc, role, roleBind := CreateSvcAccRoleRoleBind(testLabInstance)

	// Service Account
	assert.Equal(t, testLabInstance.Name+"-ttyd-svcacc", svcAcc.Name)
	assert.Equal(t, testLabInstance.Namespace, svcAcc.Namespace)

	// Role
	assert.Equal(t, testLabInstance.Name+"-ttyd-role", role.Name)
	assert.Equal(t, testLabInstance.Namespace, role.Namespace)
	assert.NotEqual(t, 0, len(role.Rules))

	// Role Binding
	assert.Equal(t, testLabInstance.Name+"-ttyd-rolebind", roleBind.Name)
	assert.Equal(t, testLabInstance.Namespace, roleBind.Namespace)
	assert.Equal(t, testLabInstance.Name+"-ttyd-svcacc", roleBind.Subjects[0].Name)
	assert.Equal(t, testLabInstance.Namespace, roleBind.Subjects[0].Namespace)
	assert.Equal(t, testLabInstance.Name+"-ttyd-role", roleBind.RoleRef.Name)
}

func TestErrorMsg(t *testing.T) {
	err := errors.New("Resource not found")
	returnValue = ErrorMsg(ctx, err, testLabInstance.Name+"-test-node-3")
	assert.Equal(t, true, returnValue.ShouldReturn)
	assert.Equal(t, ctrl.Result{}, returnValue.Result)
	assert.Equal(t, err, returnValue.Err)
}

func TestResourceExists(t *testing.T) {
	shouldReturn, err := ResourceExists(r, &corev1.Pod{}, testLabInstance.Name+"-test-node-3", testLabInstance.Namespace)
	assert.Equal(t, false, shouldReturn)
	assert.NotEqual(t, nil, err)
}

func TestCreateResource(t *testing.T) {
	createdPod := CreateResource(testLabInstance, podNode, &corev1.Pod{})
	assert.Equal(t, testLabInstance.Name+"-"+podNode.Name, createdPod.GetName())
	assert.Equal(t, testLabInstance.Namespace, createdPod.GetNamespace())
}

func TestReconcileResource(t *testing.T) {
	expectedReturnValue = ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{Requeue: true}, Err: nil}
	// Non-Existing resource
	createdPod, returnValue := ReconcileResource(r, testLabInstance, &corev1.Pod{}, podNode, testLabInstance.Name+"-"+podNode.Name)
	assert.Equal(t, expectedReturnValue, returnValue)
	assert.Equal(t, testLabInstance.Name+"-"+podNode.Name, createdPod.GetName())
	assert.Equal(t, testLabInstance.Namespace, createdPod.GetNamespace())

	// Existing resource
	createdPod, returnValue = ReconcileResource(r, testLabInstance, &corev1.Pod{}, podNode, testLabInstance.Name+"-"+podNode.Name)
	assert.Equal(t, ReturnToReconciler{ShouldReturn: false, Result: ctrl.Result{}, Err: nil}, returnValue)
}
