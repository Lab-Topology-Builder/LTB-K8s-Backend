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

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("LabTemplate Controller", func() {
	var lr *LabTemplateReconciler

	Describe("Reconcile", func() {
		var (
			ctx context.Context
			req ctrl.Request
		)
		BeforeEach(func() {
			ctx = context.Background()
			req = ctrl.Request{}
			fakeClient = fake.NewClientBuilder().WithObjects(testLabTemplateWithoutRenderedNodeSpec).Build()
			lr = &LabTemplateReconciler{Client: fakeClient, Scheme: scheme.Scheme}

		})
		Context("LabTemplate doesn't exists", func() {
			BeforeEach(func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: "test"}
			})
			It("should return NotFound error", func() {
				result, err := lr.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).To(BeNil())
			})
		})
		Context("LabTemplate exists, but nodetypes don't exist", func() {
			BeforeEach(func() {
				req.NamespacedName = types.NamespacedName{Name: testLabTemplateWithoutRenderedNodeSpec.Name, Namespace: testLabTemplateWithoutRenderedNodeSpec.Namespace}
			})
			It("should return nil error", func() {
				result, err := lr.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).To(BeNil())
			})
		})
		Context("All resources exist, and successfully renders", func() {
			BeforeEach(func() {
				lr.Client = fake.NewClientBuilder().WithObjects(testLabTemplateWithoutRenderedNodeSpec, testNodeTypePod, testNodeTypeVM).Build()
				req.NamespacedName = types.NamespacedName{Name: testLabTemplateWithoutRenderedNodeSpec.Name, Namespace: testLabTemplateWithoutRenderedNodeSpec.Namespace}
			})
			It("should return nil error and render nodespec", func() {
				result, err := lr.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).To(BeNil())
				labtemplate := &ltbv1alpha1.LabTemplate{}
				err = lr.Get(ctx, req.NamespacedName, labtemplate)
				Expect(err).To(BeNil())
				Expect(labtemplate.Spec.Nodes).To(HaveLen(2))
				Expect(labtemplate.Spec.Nodes[0].RenderedNodeSpec).To(MatchYAML(testLabTemplateWithRenderedNodeSpec.Spec.Nodes[0].RenderedNodeSpec))
				Expect(labtemplate.Spec.Nodes[1].RenderedNodeSpec).ToNot(MatchYAML(testLabTemplateWithRenderedNodeSpec.Spec.Nodes[1].RenderedNodeSpec))
			})
		})
		Context("All resources exist, but fails to render", func() {
			BeforeEach(func() {
				lr.Client = fake.NewClientBuilder().WithObjects(testLabTemplateWithRenderedNodeSpec, failingPodNodeType, testNodeTypeVM).Build()
			})
			It("should return error", func() {
				result, err := lr.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).To((BeNil()))
			})
		})
		AfterEach(func() {
			lr.Client = nil
		})
	})

	Describe("SetupWithManager", func() {
		It("should return error", func() {
			err := lr.SetupWithManager(nil)
			Expect(err).To(HaveOccurred())
		})
	})
})
