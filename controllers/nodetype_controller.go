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
	"errors"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/yaml"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	util "github.com/Lab-Topology-Builder/LTB-K8s-Backend/util"
	kubevirtv1 "kubevirt.io/api/core/v1"
)

// NodeTypeReconciler reconciles a NodeType object
type NodeTypeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

var (
	// TestData for Template rendering
	TestNodeData = ltbv1alpha1.LabInstanceNodes{
		Name: "test",
		NodeTypeRef: ltbv1alpha1.NodeTypeRef{
			Type:    "testnodetype",
			Image:   "ubuntu",
			Version: "latest",
		},
		Config: `
		#cloud-config
		password: ubuntu
		chpasswd: { expire: False }
		ssh_authorized_keys:
			- <your-ssh-pub-key>
		packages:
			- qemu-guest-agent
		runcmd:
			- [ systemctl, start, qemu-guest-agent ]`,
	}
	TestInterfaces = []ltbv1alpha1.NodeInterface{
		{
			IPv4: "192.168.0.1/24",
		},
		{
			IPv4: "172.16.0.1/24",
		},
		{
			IPv4: "10.0.0.1/24",
		},
	}
	Data = util.TemplateData{
		Node:       TestNodeData,
		Interfaces: TestInterfaces,
	}
)

//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=nodetypes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=nodetypes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=nodetypes/finalizers,verbs=update

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *NodeTypeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	nodetype := &ltbv1alpha1.NodeType{}
	err := r.Get(ctx, req.NamespacedName, nodetype)
	if err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	var renderedNodeSpec strings.Builder
	if err = util.ParseAndRenderTemplate(nodetype, &renderedNodeSpec, Data); err != nil {
		l.Error(err, "Failed to render template")
		return ctrl.Result{}, err
	}
	nodeSpecBytes := []byte(renderedNodeSpec.String())
	if nodetype.Spec.Kind == "vm" {
		// check if valid vm spec
		vmSpec := kubevirtv1.VirtualMachineSpec{}
		err := yaml.Unmarshal(nodeSpecBytes, &vmSpec)
		if err != nil {
			l.Error(err, "Failed to unmarshal NodeSpec to VMSpec")
			return ctrl.Result{}, err
		}
		l.Info("Decoded VM Spec", "Spec", vmSpec)
	} else if nodetype.Spec.Kind == "pod" {
		// check if valid pod spec
		podSpec := corev1.PodSpec{}
		err := yaml.Unmarshal(nodeSpecBytes, &podSpec)
		if err != nil {
			l.Error(err, "Failed to unmarshal NodeSpec to PodSpec")
			return ctrl.Result{}, err
		}
		l.Info("Decoded Pod Spec", "Spec", podSpec)
	} else {
		// invalid kind
		return ctrl.Result{}, errors.New("invalid Kind")
	}

	return ctrl.Result{}, nil
}

// Move to utils

// SetupWithManager sets up the controller with the Manager.
func (r *NodeTypeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbv1alpha1.NodeType{}).
		Complete(r)
}
