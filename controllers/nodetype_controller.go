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
	"fmt"
	"strings"
	"text/template"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
)

// NodeTypeReconciler reconciles a NodeType object
type NodeTypeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

type TemplateData struct {
	Node        ltbv1alpha1.LabInstanceNodes
	Connections []ltbv1alpha1.Connection
}

var (
	// TestData for Template rendering
	TestNodeData = ltbv1alpha1.LabInstanceNodes{
		Name: "test",
		NodeTypeRef: ltbv1alpha1.NodeTypeRef{
			Type:    "testnodetype",
			Version: "v1",
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
	TestConnections = []ltbv1alpha1.Connection{
		{
			Neighbors: []string{
				"test1",
				"test2",
			},
		},
	}
	Data = TemplateData{
		Node:        TestNodeData,
		Connections: TestConnections,
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
	if nodetype.Spec.Kind == "vm" {
		tmplt, err := template.New("vmTemplate").Parse(nodetype.Spec.NodeSpec)
		if err != nil {
			l.Error(err, "Failed to parse template")
		}
		var renderedVMTemplate strings.Builder
		err = tmplt.Execute(&renderedVMTemplate, Data)
		if err != nil {
			l.Error(err, "Failed to render template")
		}
		l.Info(fmt.Sprintf("Rendered VM Template: %s", renderedVMTemplate.String()))
		// check if valid vm spec

	} else if nodetype.Spec.Kind == "pod" {
		tmplt, err := template.New("vmTemplate").Parse(nodetype.Spec.NodeSpec)
		if err != nil {
			l.Error(err, "Failed to parse template")
		}
		var renderedVMTemplate strings.Builder
		err = tmplt.Execute(&renderedVMTemplate, Data)
		if err != nil {
			l.Error(err, "Failed to render template")
		}
		l.Info(fmt.Sprintf("Rendered VM Template: %s", renderedVMTemplate.String()))
		// check if valid pod spec
	} else {
		// invalid kind
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeTypeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbv1alpha1.NodeType{}).
		Complete(r)
}
