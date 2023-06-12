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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("LabTemplate Controller", func() {
	var (
		ctx context.Context // TODO we only use empty context, change?
		req ctrl.Request
		lr  *LabTemplateReconciler
	)

	// TODO: I couldn't test the rendering labtemplate because of the dependency on nodetype

	Describe("Reconcile", func() {
		BeforeEach(func() {
			req = ctrl.Request{}
			fakeClient = fake.NewClientBuilder().WithObjects(testLabTemplate).Build()
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
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: "test-labtemplate"}
			})
			It("should return nil error", func() {
				result, err := lr.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).To(BeNil())
			})
		})

		Context("All resources exist, and successfully renders", func() {
			BeforeEach(func() {
				lr.Client = fake.NewClientBuilder().WithObjects(testLabTemplate, testNodeTypePod, testNodeTypeVM).Build()
			})
			It("should return nil error", func() {
				result, err := lr.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).To(BeNil())
			})
			// TODO check more, like labtemplate content of rendered nodespec
		})

		Context("All resources exist, but fails to render", func() {
			BeforeEach(func() {
				lr.Client = fake.NewClientBuilder().WithObjects(testLabTemplate, failingPodNodeType, testNodeTypeVM).Build()
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
