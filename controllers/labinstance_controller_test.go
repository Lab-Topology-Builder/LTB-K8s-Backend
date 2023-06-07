package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TODO: Test the status of the pods and VMs
const (
	Timeout  = time.Second * 10
	Interval = time.Millisecond * 250
)

var _ = Describe("LabInstance Reconcile", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals

	var (
		ctx context.Context
		req ctrl.Request
	)
	BeforeEach(func() {
		req = ctrl.Request{}
		ctx = context.Background()
		fakeClient = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM).Build()
		r = &LabInstanceReconciler{Client: fakeClient, Scheme: scheme.Scheme}
	})

	Describe("Reconcile", func() {
		Context("Empty request", func() {
			It("should return NotFound error", func() {
				_, err := r.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
			})
		})

		Context("Namespaced request", func() {
			It("should return NotFound error", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: "test"}
				_, err := r.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
			})
		})

		Context("Network attachment not created", func() {
			It("should return nil error", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("Network attachment exists", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPod, testVM, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
			})
			It("should return nil error", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})
		})
	})
})
