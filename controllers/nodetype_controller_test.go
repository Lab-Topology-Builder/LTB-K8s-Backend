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

var _ = Describe("NodeTye Controller", func() {
	var (
		ctx context.Context
		req ctrl.Request
		ln  *NodeTypeReconciler
	)

	Describe("Reconcile", func() {
		BeforeEach(func() {
			req = ctrl.Request{}
			fakeClient = fake.NewClientBuilder().WithObjects(testNodeTypePod, testNodeVM).Build()
			ln = &NodeTypeReconciler{Client: fakeClient, Scheme: scheme.Scheme}
		})
		Context("NodeType doesn't exists", func() {
			BeforeEach(func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: "test"}
			})
			It("should return NotFound error", func() {
				result, err := ln.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).To(BeNil())
			})
		})

		Context("NodeType exists, but unmarshalling fails", func() {
			BeforeEach(func() {
				ln.Client = fake.NewClientBuilder().WithObjects(failingNodeType).Build()
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: "failingNodeType"}
			})
			It("should return error", func() {
				result, err := ln.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).To(BeNil())
			})
		})
	})
})
