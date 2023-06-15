package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
			fakeClient = fake.NewClientBuilder().WithObjects(testPodNodeType, testVM2).Build()
			ln = &NodeTypeReconciler{Client: fakeClient, Scheme: scheme.Scheme}
		})
		Context("NodeType doesn't exists", func() {
			BeforeEach(func() {
				req.NamespacedName = types.NamespacedName{Name: "test"}
			})
			It("should return NotFound error", func() {
				result, err := ln.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(apiErrors.IsNotFound(err)).To(BeFalse())
			})
		})
		Context("NodeType exists, but VM YAML is invalid", func() {
			BeforeEach(func() {
				ln.Client = fake.NewClientBuilder().WithObjects(invalidNodeSpecVMNodeType).Build()
				req.NamespacedName = types.NamespacedName{Name: invalidNodeSpecVMNodeType.Name}
			})
			It("should return error while unmarshaling, YAML is invalid", func() {
				result, err := ln.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("error converting YAML to JSON"))
			})
		})
		Context("NodeType exists, but Pod YAML is invalid", func() {
			BeforeEach(func() {
				ln.Client = fake.NewClientBuilder().WithObjects(invalidNodeSpecPodNodeType).Build()
				req.NamespacedName = types.NamespacedName{Name: invalidNodeSpecPodNodeType.Name}
			})
			It("should return error while unmarshaling, YAML is invalid", func() {
				result, err := ln.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(ContainSubstring("error converting YAML to JSON"))
			})
		})
		Context("NodeType exists, with correct YAML but wrong content", func() {
			BeforeEach(func() {
				ln.Client = fake.NewClientBuilder().WithObjects(failingVMNodeType, failingPodNodeType).Build()
				req.NamespacedName = types.NamespacedName{Name: failingVMNodeType.Name}
			})
			It("should return error while unmarshaling to VMSpec", func() {
				result, err := ln.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).ToNot(BeNil())
			})
			It("should return error while unmarshaling to PodSpec", func() {
				req.NamespacedName = types.NamespacedName{Name: failingPodNodeType.Name}
				result, err := ln.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).ToNot(BeNil())
			})
		})
		Context("Rendering NodeSpec for VM works", func() {
			BeforeEach(func() {
				ln.Client = fake.NewClientBuilder().WithObjects(testNodeVMType).Build()
				req.NamespacedName = types.NamespacedName{Name: testNodeVMType.Name}
			})
			It("should render the VMSpec successfully", func() {
				result, err := ln.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("Rendering NodeSpec for pod works", func() {
			BeforeEach(func() {
				ln.Client = fake.NewClientBuilder().WithObjects(testPodNodeType).Build()
				req.NamespacedName = types.NamespacedName{Name: testPodNodeType.Name}
			})
			It("should render the PodSpec successfully", func() {
				result, err := ln.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("Invalid nodetype kind", func() {
			BeforeEach(func() {
				ln.Client = fake.NewClientBuilder().WithObjects(invalidKindNodeType).Build()
				req.NamespacedName = types.NamespacedName{Name: invalidKindNodeType.Name}
			})
			It("should return error", func() {
				result, err := ln.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).ToNot(BeNil())
			})
		})
	})

	Describe("SetupWithManager", func() {
		It("should return error", func() {
			err := ln.SetupWithManager(nil)
			Expect(err).ToNot(BeNil())
		})
	})
})
