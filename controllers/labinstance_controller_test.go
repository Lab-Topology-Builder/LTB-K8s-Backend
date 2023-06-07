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
	Describe("Reconcile", func() {
		BeforeEach(func() {
			req = ctrl.Request{}
			ctx = context.Background()
			fakeClient = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM).Build()
			r = &LabInstanceReconciler{Client: fakeClient, Scheme: scheme.Scheme}
		})
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

		Context("Network attachment doesn't exist", func() {
			It("should return nil error", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("Lab template doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testNodeTypePod, testNodeTypeVM).Build()
			})
			It("should return not found error", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
				Expect(result).To(Equal(ctrl.Result{}))
			})
		})

		Context("Ttyd service account doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testRoleBinding, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
			})
			It("should create ttyd service account", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("Ttyd role doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
			})
			It("should create a ttyd role", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("Ttyd role binding doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
			})
			It("should create a ttyd rolebinding", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("Ttyd pod doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydService, testService, testVMIngress, testPodIngress).Build()
			})
			It("should create a ttyd pod", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("Ttyd service doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testService, testVMIngress, testPodIngress).Build()
			})
			It("should create a ttyd service", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("Node type not found", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
			})
			It("should return not found error", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
				Expect(result).To(Equal(ctrl.Result{}))
			})
		})

		Context("VM doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
			})
			It("should create a VM", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("Pod doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPodNetworkAttachmentDefinition, testVM, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
			})
			It("should create a Pod", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("Service for remote access doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPodNetworkAttachmentDefinition, testVM, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testVMIngress, testPodIngress).Build()
			})
			It("should create a Service for remote access", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("Ingress for remote access doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod, testNodeTypeVM, testPodNetworkAttachmentDefinition, testVM, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testService).Build()
			})
			It("should create an Ingress for remote access", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: testLabInstance.Name}
				result, err := r.Reconcile(ctx, req)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{Requeue: true}))
			})
		})

		Context("All resources exists", func() {
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
