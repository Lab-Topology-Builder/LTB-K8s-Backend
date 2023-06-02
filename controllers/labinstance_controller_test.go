package controllers

import (
	"context"
	"errors"
	"os"
	"testing"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/types"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestMain(m *testing.M) {
	initialize()
	code := m.Run()
	os.Exit(code)
}

func TestLabInstanceReconciler_Reconcile(t *testing.T) {
	type args struct {
		ctx context.Context
		req ctrl.Request
	}
	tests := []struct {
		name    string
		args    args
		want    ctrl.Result
		want1   error
		wantErr bool
	}{

		{
			name: "Empty request",
			args: args{
				ctx: context.Background(),
				req: ctrl.Request{},
			},
			want:    ctrl.Result{},
			want1:   nil,
			wantErr: false,
		},
		{
			name: "Namespaced request",
			args: args{
				ctx: context.Background(),
				req: ctrl.Request{NamespacedName: types.NamespacedName{Namespace: "test", Name: "test"}},
			},
			want:    ctrl.Result{},
			want1:   nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("LabInstanceReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, err)
		})
	}
}

func TestLabInstanceReconciler_ReconcileNetwork(t *testing.T) {
	type args struct {
		ctx         context.Context
		labInstance *ltbv1alpha1.LabInstance
	}
	tests := []struct {
		name string
		args args
		want ReturnToReconciler
	}{
		{
			name: "Network will be created",
			args: args{
				ctx:         context.Background(),
				labInstance: testLabInstance,
			},
			want: ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{Requeue: true}, Err: nil},
		},
		// TODO: See how the other test cases can be implemented, but we might have the issue to test the call to the client functions (eg. Get, Create, etc.)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.ReconcileNetwork(tt.args.ctx, tt.args.labInstance)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabInstanceReconciler_GetLabTemplate(t *testing.T) {
	type args struct {
		ctx         context.Context
		labInstance *ltbv1alpha1.LabInstance
		labTemplate *ltbv1alpha1.LabTemplate
	}
	tests := []struct {
		name string
		args args
		want ReturnToReconciler
	}{
		{
			name: "LabTemplate exists",
			args: args{
				ctx:         context.Background(),
				labInstance: testLabInstance,
				labTemplate: testLabTemplate,
			},
			want: ReturnToReconciler{ShouldReturn: false, Result: ctrl.Result{}, Err: nil},
		},
		{
			name: "LabTemplate can't be nil",
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
			got := r.GetLabTemplate(tt.args.ctx, tt.args.labInstance, tt.args.labTemplate)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLabInstanceReconciler_GetNodeType(t *testing.T) {
	type args struct {
		ctx         context.Context
		nodeTypeRef *ltbv1alpha1.NodeTypeRef
		nodeType    *ltbv1alpha1.NodeType
	}
	tests := []struct {
		name    string
		args    args
		want    ReturnToReconciler
		wantErr bool
	}{
		{
			name: "Node type exists",
			args: args{
				ctx:         context.Background(),
				nodeTypeRef: &normalPodNode.NodeTypeRef,
				nodeType:    testNodeTypePod,
			},
			want:    ReturnToReconciler{ShouldReturn: false, Result: ctrl.Result{}, Err: nil},
			wantErr: false,
		},
		{
			name: "Node type couldn't be retrieved",
			args: args{
				ctx:         context.Background(),
				nodeTypeRef: &nodeUndefinedNodeType.NodeTypeRef,
				nodeType:    testNodeTypePod,
			},
			want:    ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{}, Err: nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		name    string
		args    args
		want    *corev1.Pod
		wantErr bool
	}{
		{
			name: "Pod will be created",
			args: args{
				labInstance: testLabInstance,
				node:        normalPodNode,
			},
			want:    testPod,
			wantErr: false,
		},
		{
			name: "Pod couldn't be created",
			args: args{
				labInstance: testLabInstance,
				node:        nil,
			},
			want:    testPod,
			wantErr: true,
		},
		{
			name: "Unmarshal error",
			args: args{
				labInstance: testLabInstance,
				node:        podYAMLProblemNode,
			},
			want:    testPod,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapTemplateToPod(tt.args.labInstance, tt.args.node)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.Equal(t, tt.want.GetName(), got.GetName())
			assert.Equal(t, tt.want.GetNamespace(), got.GetNamespace())
			assert.Equal(t, tt.want.GetLabels(), got.GetLabels())
			assert.Equal(t, tt.want.GetAnnotations(), got.GetAnnotations())
			assert.Equal(t, nil, err)
		})
	}
}

// TODO: this functions needs to be checked
func TestMapTemplateToVM(t *testing.T) {
	type args struct {
		labInstance *ltbv1alpha1.LabInstance
		node        *ltbv1alpha1.LabInstanceNodes
	}
	tests := []struct {
		name    string
		args    args
		want    *kubevirtv1.VirtualMachine
		wantErr bool
	}{
		{
			name: "VM mapping successful",
			args: args{
				labInstance: testLabInstance,
				node:        normalVMNode,
			},
			want:    testVM,
			wantErr: false,
		},
		{
			name: "Failure - Unmarshal error",
			args: args{
				labInstance: testLabInstance,
				node:        vmYAMLProblemNode,
			},
			want:    testVM,
			wantErr: true,
		},
		{
			name: "Failure - Empty LabInstance",
			args: args{
				labInstance: nil,
				node:        normalVMNode,
			},
			want:    testVM,
			wantErr: true,
		},
		{
			name: "Failure - Empty node",
			args: args{
				labInstance: testLabInstance,
				node:        nil,
			},
			want:    testVM,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapTemplateToVM(tt.args.labInstance, tt.args.node)
			if tt.wantErr {
				assert.Error(t, err)
				return
			} else {
				assert.Equal(t, nil, err)
				assert.Equal(t, tt.want.GetName(), got.GetName())
				assert.Equal(t, tt.want.GetNamespace(), got.GetNamespace())
				assert.Equal(t, tt.want.GetLabels(), got.GetLabels())
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
		name          string
		args          args
		runningStatus bool
		isPodRunning  bool
		isVMReady     bool
	}{
		{
			name: "Running status",
			args: args{
				ctx:         context.Background(),
				pods:        []*corev1.Pod{testPod},
				vms:         []*kubevirtv1.VirtualMachine{testVM},
				labInstance: testLabInstance,
			},
			runningStatus: true,
			isPodRunning:  true,
			isVMReady:     true,
		},
		{
			name: "Pending status",
			args: args{
				ctx:         context.Background(),
				pods:        []*corev1.Pod{testPod, testNodePod},
				vms:         []*kubevirtv1.VirtualMachine{testVM},
				labInstance: testLabInstance,
			},
			runningStatus: false,
			isPodRunning:  false,
			isVMReady:     true,
		},
		{
			name: "Not Ready status",
			args: args{
				ctx:         context.Background(),
				pods:        []*corev1.Pod{testPod},
				vms:         []*kubevirtv1.VirtualMachine{testVM, testNodeVM},
				labInstance: testLabInstance,
			},
			runningStatus: false,
			isPodRunning:  true,
			isVMReady:     false,
		},
		{
			name: "Not Running status",
			args: args{
				ctx:         context.Background(),
				pods:        []*corev1.Pod{testPod, testNodePod},
				vms:         []*kubevirtv1.VirtualMachine{testVM, testNodeVM},
				labInstance: testLabInstance,
			},
			runningStatus: false,
			isPodRunning:  false,
			isVMReady:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateLabInstanceStatus(tt.args.ctx, tt.args.pods, tt.args.vms, tt.args.labInstance)
			t.Log(tt.args.labInstance.Status)
			if tt.runningStatus && tt.isPodRunning && tt.isVMReady {
				assert.Equal(t, "Running", tt.args.labInstance.Status.Status)
				assert.Equal(t, "1/1", tt.args.labInstance.Status.NumPodsRunning)
				assert.Equal(t, "1/1", tt.args.labInstance.Status.NumVMsRunning)
			} else if !tt.runningStatus && !tt.isPodRunning && tt.isVMReady {
				assert.NotEqual(t, "Running", tt.args.labInstance.Status.Status)
				assert.Equal(t, "1/2", tt.args.labInstance.Status.NumPodsRunning)
				assert.Equal(t, "1/1", tt.args.labInstance.Status.NumVMsRunning)
			} else if !tt.runningStatus && tt.isPodRunning && !tt.isVMReady {
				assert.NotEqual(t, "Running", tt.args.labInstance.Status.Status)
				assert.Equal(t, "1/1", tt.args.labInstance.Status.NumPodsRunning)
				assert.Equal(t, "1/2", tt.args.labInstance.Status.NumVMsRunning)
			} else {
				assert.NotEqual(t, "Running", tt.args.labInstance.Status.Status)
				assert.Equal(t, "1/2", tt.args.labInstance.Status.NumPodsRunning)
				assert.Equal(t, "1/2", tt.args.labInstance.Status.NumVMsRunning)
			}
		})
	}
}

func TestLabInstanceReconciler_SetupWithManager(t *testing.T) {
	type args struct {
		mgr ctrl.Manager
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Error case",
			args:    args{mgr: nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		{
			name: "Resource already exists and retrieval is successful",
			args: args{
				r:            r,
				labInstance:  testLabInstance,
				resource:     &corev1.Pod{},
				node:         normalPodNode,
				resourceName: testLabInstance.Name + "-" + normalPodNode.Name,
			},
			want:  testPod,
			want1: ReturnToReconciler{ShouldReturn: false, Result: ctrl.Result{}, Err: nil},
		},
		{
			name: "Resource created successfully",
			args: args{
				r:            r,
				labInstance:  testLabInstance,
				resource:     &corev1.Service{},
				node:         nodeUndefinedNodeType,
				resourceName: testLabInstance.Name + "-" + nodeUndefinedNodeType.Name + "-remote-access",
			},
			want:  testService,
			want1: ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{Requeue: true}, Err: nil},
		},
		// {
		// 	name: "Resource could not be created",
		// 	args: args{
		// 		r:            r,
		// 		labInstance:  testLabInstance,
		// 		resource:     &networkingv1.Ingress{},
		// 		node:         testNode,
		// 		resourceName: testLabInstance.Name + "-" + testNode.Name,
		// 	},
		// 	want:  nil,
		// 	want1: ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{}, Err: errors.New("failed to create resource")},
		// },

		// TODO: There are other two cases to test:
		// 1.resource already exists, but could not be retrieved
		// 2.resource does not exist, but could not be created
		// Those are the calls to the reconciler functions that return an error (e.g. r.Get(), r.Create())
		// {
		// 	name: "Resource already exists, but could not be retrieved",
		// 	args: args{
		// 		r:            r,
		// 		labInstance:  testLabInstance,
		// 		resource:     &corev1.Pod{},
		// 		node:         testNode,
		// 		resourceName: testLabInstance.Name + "-" + testNodePod.Name,
		// 	},
		// 	want:  testNodePod,
		// 	want1: ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{Requeue: true}, Err: fmt.Errorf("error")},
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ReconcileResource(tt.args.r, tt.args.labInstance, tt.args.resource, tt.args.node, tt.args.resourceName)
			assert.Equal(t, tt.want.GetName(), got.GetName())
			assert.Equal(t, tt.want.GetNamespace(), got.GetNamespace())
			assert.Equal(t, tt.want.GetLabels(), got.GetLabels())
			assert.Equal(t, tt.want1, got1)
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
		name    string
		args    args
		want    client.Object
		wantErr bool
	}{
		{
			name: "Resource not supported",
			args: args{
				labInstance: testLabInstance,
				node:        nil,
				resource:    &corev1.Secret{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "VM creation successful",
			args: args{
				labInstance: testLabInstance,
				node:        normalVMNode,
				resource:    &kubevirtv1.VirtualMachine{},
			},
			want:    testVM,
			wantErr: false,
		},

		{
			name: "Pod creation successful",
			args: args{
				labInstance: testLabInstance,
				node:        nodeUndefinedNodeType,
				resource:    &corev1.Pod{},
			},
			want:    testNodePod,
			wantErr: false,
		},
		{
			name: "Ingress creation successful",
			args: args{
				labInstance: testLabInstance,
				node:        normalVMNode,
				resource:    &networkingv1.Ingress{},
			},
			want:    testVMIngress,
			wantErr: false,
		},
		{
			name: "Service Account creation successful",
			args: args{
				labInstance: testLabInstance,
				node:        normalVMNode,
				resource:    &corev1.ServiceAccount{},
			},
			want:    testServiceAccount,
			wantErr: false,
		},
		{
			name: "Role creation successful",
			args: args{
				labInstance: testLabInstance,
				node:        normalVMNode,
				resource:    &rbacv1.Role{},
			},
			want:    testRole,
			wantErr: false,
		},
		{
			name: "RoleBinding creation successful",
			args: args{
				labInstance: testLabInstance,
				node:        normalVMNode,
				resource:    &rbacv1.RoleBinding{},
			},
			want:    testRoleBinding,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateResource(tt.args.labInstance, tt.args.node, tt.args.resource)
			if (err != nil) && tt.wantErr {
				assert.Error(t, err)
			} else if err == nil && !tt.wantErr {
				assert.Equal(t, tt.want.GetName(), got.GetName())
				assert.Equal(t, tt.want.GetNamespace(), got.GetNamespace())
			} else {
				t.Errorf("CreateResource() error = %v, wantErr %v", err, tt.wantErr)
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
		{
			name: "Resource exists",
			args: args{
				r:            r,
				resource:     &corev1.Pod{},
				resourceName: testLabInstance.Name + "-" + normalPodNode.Name,
				nameSpace:    namespace,
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "Resource does not exist",
			args: args{
				r:            r,
				resource:     &corev1.Pod{},
				resourceName: testLabInstance.Name + "-" + nodeUndefinedNodeType.Name + "-1",
				nameSpace:    namespace,
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "Error occurred",
			args: args{
				r:            r,
				resource:     nil,
				resourceName: testLabInstance.Name + "-nil",
				nameSpace:    namespace,
			},
			want:    true,
			wantErr: true,
		},
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
		{
			name: "Ingress will be created for a pod",
			args: args{
				labInstance: testLabInstance,
				node:        nodeUndefinedNodeType,
			},
			want: testPodIngress,
		},
		{
			name: "Ingress will be created for a vm",
			args: args{
				labInstance: testLabInstance,
				node:        normalVMNode,
			},
			want: testVMIngress,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateIngress(tt.args.labInstance, tt.args.node)
			assert.Equal(t, tt.want.GetName(), got.GetName())
			assert.Equal(t, tt.want.GetNamespace(), got.GetNamespace())
			assert.Equal(t, tt.want.GetLabels(), got.GetLabels())
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
		{
			name: "Ttyd Pod will be created",
			args: args{
				labInstance: testLabInstance,
				node:        nil,
			},
			want: testTtydPod,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreatePod(tt.args.labInstance, tt.args.node)
			assert.Equal(t, tt.want.GetName(), got.GetName())
			assert.Equal(t, tt.want.GetNamespace(), got.GetNamespace())
			assert.Equal(t, tt.want.GetLabels(), got.GetLabels())
			assert.Equal(t, nil, err)
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
		{
			name: "Service will be created",
			args: args{

				labInstance: testLabInstance,
				node:        nodeUndefinedNodeType,
			},
			want: testService,
		},
		{
			name: "Ttyd Service will be created",
			args: args{

				labInstance: testLabInstance,
				node:        nil,
			},
			want: testTtydService,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateService(tt.args.labInstance, tt.args.node)
			assert.Equal(t, tt.want.GetName(), got.GetName())
			assert.Equal(t, tt.want.GetNamespace(), got.GetNamespace())
			assert.Equal(t, tt.want.GetLabels(), got.GetLabels())
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
		{
			name: "Error when checking if resource exists",
			args: args{
				ctx:      context.Background(),
				err:      errors.New("error"),
				resource: "Pod",
			},
			want: ReturnToReconciler{ShouldReturn: true, Result: ctrl.Result{}, Err: errors.New("error")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ErrorMsg(tt.args.ctx, tt.args.err, tt.args.resource)
			assert.Equal(t, tt.want, got)
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
		{
			name: "Service Account, Role, Rolebinding will be created",
			args: args{
				labInstance: testLabInstance,
			},
			want:  testServiceAccount,
			want1: testRole,
			want2: testRoleBinding,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := CreateSvcAccRoleRoleBind(tt.args.labInstance)
			assert.Equal(t, tt.want.GetName(), got.GetName())
			assert.Equal(t, tt.want.GetNamespace(), got.GetNamespace())
			assert.Equal(t, tt.want.GetLabels(), got.GetLabels())
			assert.Equal(t, tt.want1.GetName(), got1.GetName())
			assert.Equal(t, tt.want1.GetNamespace(), got1.GetNamespace())
			assert.Equal(t, tt.want1.GetLabels(), got1.GetLabels())
			assert.Equal(t, tt.want2.GetName(), got2.GetName())
			assert.Equal(t, tt.want2.GetNamespace(), got2.GetNamespace())
			assert.Equal(t, tt.want2.GetLabels(), got2.GetLabels())
		})
	}
}
