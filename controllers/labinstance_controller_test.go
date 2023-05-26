package controllers

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	network "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	fakeClient                      client.Client
	testPod                         *corev1.Pod
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

	// vmConfig := `
	//   #cloud-config
	//   password: ubuntu
	//   chpasswd: { expire: False }
	//   ssh_pwauth: True
	//   packages:
	//     - qemu-guest-agent
	//   runcmd:
	//     - [ systemctl, start, qemu-guest-agent ]
	// `

	nodeSpecYAMLPod := `
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
	`

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
				},
				{
					Name: "test-node-1",
					NodeTypeRef: ltbv1alpha1.NodeTypeRef{
						Type:    testNodeTypePod.Name,
						Image:   "ubuntu",
						Version: "20.04",
					},
					RenderedNodeSpec: nodeSpecYAMLPod,
				},
				{
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
				},
			},
		},
	}
	vmNode = &testLabTemplate.Spec.Nodes[0]
	podNode = &testLabTemplate.Spec.Nodes[1]
	testNode = &testLabTemplate.Spec.Nodes[2]

	testPod = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLabInstance.Name + "-" + podNode.Name,
			Namespace: namespace,
			Annotations: map[string]string{
				"k8s.v1.cni.cncf.io/networks": testLabInstance.Name + "-pod",
			},
			Labels: map[string]string{
				"app": testLabInstance.Name + "-" + podNode.Name + "-remote-access",
			},
		},
	}

	// TODO: Need to check if this
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
	expectedReturnValue = ReturnToReconciler{ShouldReturn: false, Result: ctrl.Result{}, Err: nil}

	fakeClient = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM).Build()
	r = &LabInstanceReconciler{Client: fakeClient, Scheme: scheme.Scheme}
}

// func TestGetTemplate(t *testing.T) {
// 	returnValue = r.GetLabTemplate(ctx, testLabInstance, testLabTemplate)
// 	assert.Equal(t, expectedReturnValue, returnValue)
// }

// func TestGetNodeType(t *testing.T) {
// 	returnValue = r.GetNodeType(ctx, &podNode.NodeTypeRef, testNodeTypePod)
// 	assert.Equal(t, expectedReturnValue, returnValue)
// }

// func TestCreatePod(t *testing.T) {
// 	createdPod := CreatePod(testLabInstance, podNode)
// 	assert.Equal(t, testLabInstance.Name+"-"+podNode.Name, createdPod.Name)
// 	assert.Equal(t, testLabInstance.Namespace, createdPod.Namespace)
// 	createdTtydPod := CreatePod(testLabInstance, nil)
// 	assert.Equal(t, testLabInstance.Name+"-ttyd-pod", createdTtydPod.Name)
// }

// // TODO: Need to figure out how to test this
// // func TestMapTemplateToVM(t *testing.T) {
// // 	t.Log(vmNode)
// // 	mappedVM := MapTemplateToVM(testLabInstance, vmNode)
// // 	t.Log(mappedVM)
// // 	t.Log(vmNode)
// // 	assert.Equal(t, testLabInstance.Name+"-"+vmNode.Name, mappedVM.Name)
// // }

// func TestCreateIngress(t *testing.T) {
// 	// Pod ingress
// 	createdPodIngress := CreateIngress(testLabInstance, podNode)
// 	assert.Equal(t, testLabInstance.Name+"-"+podNode.Name+"-ingress", createdPodIngress.Name)
// 	assert.Equal(t, testLabInstance.Namespace, createdPodIngress.Namespace)

// 	// VM ingress
// 	createdVMIngress := CreateIngress(testLabInstance, vmNode)
// 	assert.Equal(t, testLabInstance.Name+"-"+vmNode.Name+"-ingress", createdVMIngress.Name)
// 	assert.Equal(t, testLabInstance.Namespace, createdVMIngress.Namespace)
// }

// func TestCreateService(t *testing.T) {
// 	// Pod service
// 	createdPodService := CreateService(testLabInstance, podNode)
// 	assert.Equal(t, testLabInstance.Name+"-"+podNode.Name+"-remote-access", createdPodService.Name)
// 	assert.Equal(t, testLabInstance.Namespace, createdPodService.Namespace)
// 	assert.Equal(t, "LoadBalancer", string(createdPodService.Spec.Type))
// 	assert.Equal(t, 0, len(createdPodService.Spec.Ports))

// 	// VM service
// 	createdVMService := CreateService(testLabInstance, vmNode)
// 	assert.Equal(t, testLabInstance.Name+"-"+vmNode.Name+"-remote-access", createdVMService.Name)
// 	assert.Equal(t, testLabInstance.Namespace, createdVMService.Namespace)
// 	assert.Equal(t, "LoadBalancer", string(createdVMService.Spec.Type))
// 	assert.Equal(t, 1, len(createdVMService.Spec.Ports))

// 	// TTYD service
// 	createdTTYDService := CreateService(testLabInstance, nil)
// 	assert.Equal(t, testLabInstance.Name+"-ttyd-service", createdTTYDService.Name)
// 	assert.Equal(t, testLabInstance.Namespace, createdTTYDService.Namespace)
// 	assert.Equal(t, "ClusterIP", string(createdTTYDService.Spec.Type))
// 	assert.Equal(t, 1, len(createdTTYDService.Spec.Ports))
// 	assert.Equal(t, 7681, int(createdTTYDService.Spec.Ports[0].Port))
// }

// func TestCreateSvcAccRoleRoleBind(t *testing.T) {
// 	svcAcc, role, roleBind := CreateSvcAccRoleRoleBind(testLabInstance)

// 	// Service Account
// 	assert.Equal(t, testLabInstance.Name+"-ttyd-svcacc", svcAcc.Name)
// 	assert.Equal(t, testLabInstance.Namespace, svcAcc.Namespace)

// 	// Role
// 	assert.Equal(t, testLabInstance.Name+"-ttyd-role", role.Name)
// 	assert.Equal(t, testLabInstance.Namespace, role.Namespace)
// 	assert.NotEqual(t, 0, len(role.Rules))

// 	// Role Binding
// 	assert.Equal(t, testLabInstance.Name+"-ttyd-rolebind", roleBind.Name)
// 	assert.Equal(t, testLabInstance.Namespace, roleBind.Namespace)
// 	assert.Equal(t, testLabInstance.Name+"-ttyd-svcacc", roleBind.Subjects[0].Name)
// 	assert.Equal(t, testLabInstance.Namespace, roleBind.Subjects[0].Namespace)
// 	assert.Equal(t, testLabInstance.Name+"-ttyd-role", roleBind.RoleRef.Name)
// }

// func TestErrorMsg(t *testing.T) {
// 	err := errors.New("Resource not found")
// 	returnValue = ErrorMsg(ctx, err, testLabInstance.Name+"-test-node-3")
// 	assert.Equal(t, true, returnValue.ShouldReturn)
// 	assert.Equal(t, ctrl.Result{}, returnValue.Result)
// 	assert.Equal(t, err, returnValue.Err)
// }

// func TestResourceExists(t *testing.T) {
// 	shouldReturn, err := ResourceExists(r, &corev1.Pod{}, testLabInstance.Name+"-test-node-3", testLabInstance.Namespace)
// 	assert.Equal(t, false, shouldReturn)
// 	assert.NotEqual(t, nil, err)
// }

// func TestCreateResource(t *testing.T) {
// 	createdPod := CreateResource(testLabInstance, podNode, &corev1.Pod{})
// 	assert.Equal(t, testLabInstance.Name+"-"+podNode.Name, createdPod.GetName())
// 	assert.Equal(t, testLabInstance.Namespace, createdPod.GetNamespace())
// }

// func TestReconcileResource(t *testing.T) {
// 	expectedReturnValue = ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{Requeue: true}}
// 	// Non-Existing resource
// 	createdPod, returnValue := ReconcileResource(r, testLabInstance, &corev1.Pod{}, podNode, testLabInstance.Name+"-"+podNode.Name)
// 	assert.Equal(t, expectedReturnValue, returnValue)
// 	assert.Equal(t, testLabInstance.Name+"-"+podNode.Name, createdPod.GetName())
// 	assert.Equal(t, testLabInstance.Namespace, createdPod.GetNamespace())

// 	// Existing resource
// 	createdPod, returnValue = ReconcileResource(r, testLabInstance, &corev1.Pod{}, podNode, testLabInstance.Name+"-"+podNode.Name)
// 	assert.Equal(t, ReturnToReconciler{ShouldReturn: false, Result: ctrl.Result{}, Err: nil}, returnValue)

// 	// Error

// }

// func TestReconcileNetwork(t *testing.T) {
// 	returnValue = r.ReconcileNetwork(ctx, testLabInstance)
// 	t.Log(returnValue)
// 	assert.Equal(t, ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{Requeue: true}}, returnValue)
// }

func TestLabInstanceReconciler_Reconcile(t *testing.T) {
	type fields struct {
		Client client.Client
		Scheme *runtime.Scheme
	}
	type args struct {
		ctx context.Context
		req ctrl.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ctrl.Result
		wantErr bool
	}{
		{
			name: "Happy Case with empty request",
			fields: fields{
				Client: fakeClient,
				Scheme: scheme.Scheme,
			},
			args: args{
				ctx: context.Background(),
				req: ctrl.Request{},
			},
			want:    ctrl.Result{},
			wantErr: false,
		},
		{
			name: "Happy case with namespaced request",
			fields: fields{
				Client: fakeClient,
				Scheme: scheme.Scheme,
			},
			args: args{
				ctx: context.Background(),
				req: ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "test", Name: "test"}},
			},
			want:    ctrl.Result{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LabInstanceReconciler{
				Client: tt.fields.Client,
				Scheme: tt.fields.Scheme,
			}
			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("LabInstanceReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabInstanceReconciler_ReconcileNetwork(t *testing.T) {
	type fields struct {
		Client client.Client
		Scheme *runtime.Scheme
	}
	type args struct {
		ctx         context.Context
		labInstance *ltbv1alpha1.LabInstance
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   ReturnToReconciler
	}{
		{
			name: "Happy Case",
			fields: fields{
				Client: fakeClient,
				Scheme: scheme.Scheme,
			},
			args: args{
				ctx:         context.Background(),
				labInstance: testLabInstance,
			},
			want: ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{Requeue: true}, Err: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LabInstanceReconciler{
				Client: tt.fields.Client,
				Scheme: tt.fields.Scheme,
			}
			got := r.ReconcileNetwork(tt.args.ctx, tt.args.labInstance)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabInstanceReconciler_GetLabTemplate(t *testing.T) {
	type fields struct {
		Client client.Client
		Scheme *runtime.Scheme
	}
	type args struct {
		ctx         context.Context
		labInstance *ltbv1alpha1.LabInstance
		labTemplate *ltbv1alpha1.LabTemplate
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   ReturnToReconciler
	}{
		{
			name: "Happy Case",
			fields: fields{
				Client: fakeClient,
				Scheme: scheme.Scheme,
			},
			args: args{
				ctx:         context.Background(),
				labInstance: testLabInstance,
				labTemplate: testLabTemplate,
			},
			want: ReturnToReconciler{ShouldReturn: false, Result: ctrl.Result{}, Err: nil},
		},
		{
			name: "Error Case",
			fields: fields{
				Client: fakeClient,
				Scheme: scheme.Scheme,
			},
			args: args{
				ctx:         context.Background(),
				labInstance: testLabInstance,
				labTemplate: nil,
			},
			want: ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{}, Err: errors.New("expected pointer, but got nil")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LabInstanceReconciler{
				Client: tt.fields.Client,
				Scheme: tt.fields.Scheme,
			}
			got := r.GetLabTemplate(tt.args.ctx, tt.args.labInstance, tt.args.labTemplate)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabInstanceReconciler_GetNodeType(t *testing.T) {
	type fields struct {
		Client client.Client
		Scheme *runtime.Scheme
	}
	type args struct {
		ctx         context.Context
		nodeTypeRef *ltbv1alpha1.NodeTypeRef
		nodeType    *ltbv1alpha1.NodeType
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ReturnToReconciler
		wantErr bool
	}{
		{
			name: "Happy Case",
			fields: fields{
				Client: fakeClient,
				Scheme: scheme.Scheme,
			},
			args: args{
				ctx:         context.Background(),
				nodeTypeRef: &podNode.NodeTypeRef,
				nodeType:    testNodeTypePod,
			},
			want:    ReturnToReconciler{ShouldReturn: false, Result: ctrl.Result{}, Err: nil},
			wantErr: false,
		},
		{
			name: "Error Case",
			fields: fields{
				Client: fakeClient,
				Scheme: scheme.Scheme,
			},
			args: args{
				ctx:         context.Background(),
				nodeTypeRef: &testNode.NodeTypeRef,
				nodeType:    testNodeTypePod,
			},
			want:    ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{}, Err: errors.New("Node type doesn't exist")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LabInstanceReconciler{
				Client: tt.fields.Client,
				Scheme: tt.fields.Scheme,
			}
			got := r.GetNodeType(tt.args.ctx, tt.args.nodeTypeRef, tt.args.nodeType)
			if tt.wantErr {
				assert.Error(t, got.Err)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapTemplateToPod(t *testing.T) {
	type args struct {
		labInstance *ltbv1alpha1.LabInstance
		node        *ltbv1alpha1.LabInstanceNodes
	}
	tests := []struct {
		name string
		args args
		want *corev1.Pod
	}{
		// {
		// 	name: "Error Case",
		// 	args: args{
		// 		labInstance: testLabInstance,
		// 		node:        podNode,
		// 	},
		// 	want: testPod,
		// },
		// {},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapTemplateToPod(tt.args.labInstance, tt.args.node)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapTemplateToVM(t *testing.T) {
	type args struct {
		labInstance *ltbv1alpha1.LabInstance
		node        *ltbv1alpha1.LabInstanceNodes
	}
	tests := []struct {
		name string
		args args
		want *kubevirtv1.VirtualMachine
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MapTemplateToVM(tt.args.labInstance, tt.args.node); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapTemplateToVM() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateLabInstanceStatus(t *testing.T) {
	type args struct {
		ctx         context.Context
		pods        []*corev1.Pod
		vms         []*kubevirtv1.VirtualMachine
		labInstance *ltbv1alpha1.LabInstance
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateLabInstanceStatus(tt.args.ctx, tt.args.pods, tt.args.vms, tt.args.labInstance)
		})
	}
}

func TestLabInstanceReconciler_SetupWithManager(t *testing.T) {
	type fields struct {
		Client client.Client
		Scheme *runtime.Scheme
	}
	type args struct {
		mgr ctrl.Manager
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LabInstanceReconciler{
				Client: tt.fields.Client,
				Scheme: tt.fields.Scheme,
			}
			if err := r.SetupWithManager(tt.args.mgr); (err != nil) != tt.wantErr {
				t.Errorf("LabInstanceReconciler.SetupWithManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReconcileResource(t *testing.T) {
	type args struct {
		r            *LabInstanceReconciler
		labInstance  *ltbv1alpha1.LabInstance
		resource     client.Object
		node         *ltbv1alpha1.LabInstanceNodes
		resourceName string
	}
	tests := []struct {
		name  string
		args  args
		want  client.Object
		want1 ReturnToReconciler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ReconcileResource(tt.args.r, tt.args.labInstance, tt.args.resource, tt.args.node, tt.args.resourceName)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ReconcileResource() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("ReconcileResource() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestCreateResource(t *testing.T) {
	type args struct {
		labInstance *ltbv1alpha1.LabInstance
		node        *ltbv1alpha1.LabInstanceNodes
		resource    client.Object
	}
	tests := []struct {
		name string
		args args
		want client.Object
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateResource(tt.args.labInstance, tt.args.node, tt.args.resource); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceExists(t *testing.T) {
	type args struct {
		r            *LabInstanceReconciler
		resource     client.Object
		resourceName string
		nameSpace    string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResourceExists(tt.args.r, tt.args.resource, tt.args.resourceName, tt.args.nameSpace)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResourceExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ResourceExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateIngress(t *testing.T) {
	type args struct {
		labInstance *ltbv1alpha1.LabInstance
		node        *ltbv1alpha1.LabInstanceNodes
	}
	tests := []struct {
		name string
		args args
		want *networkingv1.Ingress
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateIngress(tt.args.labInstance, tt.args.node); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateIngress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreatePod(t *testing.T) {
	type args struct {
		labInstance *ltbv1alpha1.LabInstance
		node        *ltbv1alpha1.LabInstanceNodes
	}
	tests := []struct {
		name string
		args args
		want *corev1.Pod
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreatePod(tt.args.labInstance, tt.args.node); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreatePod() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateService(t *testing.T) {
	type args struct {
		labInstance *ltbv1alpha1.LabInstance
		node        *ltbv1alpha1.LabInstanceNodes
	}
	tests := []struct {
		name string
		args args
		want *corev1.Service
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateService(tt.args.labInstance, tt.args.node); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateSvcAccRoleRoleBind(t *testing.T) {
	type args struct {
		labInstance *ltbv1alpha1.LabInstance
	}
	tests := []struct {
		name  string
		args  args
		want  *corev1.ServiceAccount
		want1 *rbacv1.Role
		want2 *rbacv1.RoleBinding
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := CreateSvcAccRoleRoleBind(tt.args.labInstance)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateSvcAccRoleRoleBind() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("CreateSvcAccRoleRoleBind() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("CreateSvcAccRoleRoleBind() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestErrorMsg(t *testing.T) {
	type args struct {
		ctx      context.Context
		err      error
		resource string
	}
	tests := []struct {
		name string
		args args
		want ReturnToReconciler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorMsg(tt.args.ctx, tt.args.err, tt.args.resource); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ErrorMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}
