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
				lr.Client = fake.NewClientBuilder().WithObjects(testLabTemplateWithoutRenderedNodeSpec, testPodNodeType, testNodeVMType).Build()
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
				lr.Client = fake.NewClientBuilder().WithObjects(testLabTemplateWithRenderedNodeSpec, failingPodNodeType, testNodeVMType).Build()
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

		Context("Rendering template fails", func() {
			BeforeEach(func() {
				lr.Client = fake.NewClientBuilder().WithObjects(testLabTemplateWithoutRenderedNodeSpec2, renderInvalidNodeType, testPodRenderSpecProblem).Build()
				req.NamespacedName = types.NamespacedName{Name: testLabTemplateWithoutRenderedNodeSpec2.Name}
			})
			It("should return error", func() {
				result, err := lr.Reconcile(ctx, req)
				Expect(result).To(Equal(ctrl.Result{}))
				Expect(err).ToNot((BeNil()))
			})
		})
	})

	Describe("SetupWithManager", func() {
		It("should return error", func() {
			err := lr.SetupWithManager(nil)
			Expect(err).To(HaveOccurred())
		})
	})
})
