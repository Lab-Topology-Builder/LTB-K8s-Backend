package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	kubevirtv1 "kubevirt.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TODO: Test the status of the pods and VMs
const (
	Timeout  = time.Second * 10
	Interval = time.Millisecond * 250
)

var _ = Describe("LabInstance Reconcile", func() {

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

		AfterEach(func() {
			r.Client = nil
		})

	})
	Describe("ReconcileNetwork", func() {
		var (
			ctx context.Context
		)
		Context("Network Attachment couldn't be created, labInstance nil", func() {
			It("should return error", func() {
				returnValue := r.ReconcileNetwork(ctx, nil)
				Expect(returnValue.Result).To(Equal(ctrl.Result{}))
				Expect(returnValue.Err).To(Equal(apiErrors.NewBadRequest("labInstance is nil")))
				Expect(returnValue.ShouldReturn).To(BeTrue())
			})
		})
		// TODO: See how the rest of the tests for this function can be written
		Context("Network Attachment gets created", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should create a network attachment", func() {
				returnValue := r.ReconcileNetwork(ctx, testLabInstance)
				Expect(returnValue.Result).To(Equal(ctrl.Result{Requeue: true}))
				Expect(returnValue.Err).To(BeNil())
				Expect(returnValue.ShouldReturn).To(BeTrue())
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("ReconcileResource", func() {
		Context("Resource couldn't be created, labInstance nil", func() {
			It("should return error", func() {
				resource, returnValue := ReconcileResource(r, nil, &corev1.Pod{}, nil, "test-pod")
				Expect(resource).To(BeNil())
				Expect(returnValue.Result).To(Equal(ctrl.Result{}))
				Expect(returnValue.Err).To(Equal(apiErrors.NewBadRequest("labInstance is nil")))
				Expect(returnValue.ShouldReturn).To(BeTrue())
			})
		})
		Context("Resource already exists", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testPod).Build()
			})

			It("should not create a resource, but retrieve it", func() {
				resource, returnValue := ReconcileResource(r, testLabInstance, &corev1.Pod{}, nil, testLabInstance.Name+"-"+normalPodNode.Name)
				Expect(resource.GetName()).To(Equal(testPod.Name))
				Expect(resource.GetNamespace()).To(Equal(testPod.Namespace))
				Expect(resource.GetAnnotations()).To(Equal(testPod.Annotations))
				Expect(returnValue.Result).To(Equal(ctrl.Result{}))
				Expect(returnValue.Err).To(BeNil())
				Expect(returnValue.ShouldReturn).To(BeFalse())
			})
		})
		Context("Resource doesn't exist, but couldn't created", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should return error", func() {
				resource, returnValue := ReconcileResource(r, testLabInstance, &corev1.Secret{}, nil, "test-secret")
				Expect(resource).To(BeNil())
				Expect(returnValue.Result).To(Equal(ctrl.Result{}))
				Expect(returnValue.Err).To(HaveOccurred())
				Expect(returnValue.ShouldReturn).To(BeTrue())
			})
		})

		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("CreateResource", func() {
		Context("Unsupport resource type", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should return error", func() {
				resource, err := CreateResource(testLabInstance, nil, &corev1.Secret{})
				Expect(resource).To(BeNil())
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Resource creation succeeds", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should create a VM successfully", func() {
				resource, err := CreateResource(testLabInstance, normalVMNode, &kubevirtv1.VirtualMachine{})
				Expect(resource.GetName()).To(Equal(testVM.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create a Pod successfully", func() {
				resource, err := CreateResource(testLabInstance, normalPodNode, &corev1.Pod{})
				Expect(resource.GetName()).To(Equal(testPod.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create a Service successfully", func() {
				resource, err := CreateResource(testLabInstance, normalVMNode, &corev1.Service{})
				Expect(resource.GetName()).To(Equal(testService.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create an Ingress successfully", func() {
				resource, err := CreateResource(testLabInstance, normalVMNode, &networkingv1.Ingress{})
				Expect(resource.GetName()).To(Equal(testVMIngress.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create a service account successfully", func() {
				resource, err := CreateResource(testLabInstance, nil, &corev1.ServiceAccount{})
				Expect(resource.GetName()).To(Equal(testServiceAccount.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create a role successfully", func() {
				resource, err := CreateResource(testLabInstance, nil, &rbacv1.Role{})
				Expect(resource.GetName()).To(Equal(testRole.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create a role binding successfully", func() {
				resource, err := CreateResource(testLabInstance, nil, &rbacv1.RoleBinding{})
				Expect(resource.GetName()).To(Equal(testRoleBinding.Name))
				Expect(err).NotTo(HaveOccurred())
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("ResourceExists", func() {
		Context("Resource exists", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testVM).Build()
			})
			It("should return true", func() {
				exists, err := ResourceExists(r, &kubevirtv1.VirtualMachine{}, testLabInstance.Name+"-"+normalVMNode.Name, testLabInstance.Namespace)
				Expect(exists).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("Resource doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should return false", func() {
				exists, err := ResourceExists(r, &corev1.Pod{}, testLabInstance.Name+"-"+normalPodNode.Name, testLabInstance.Namespace)
				Expect(exists).To(BeFalse())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("GetLabTemplate", func() {
		var (
			ctx context.Context
		)
		Context("LabTemplate doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should return error", func() {
				returnValue := r.GetLabTemplate(ctx, testLabInstance, testLabTemplate)
				Expect(returnValue.Result).To(Equal(ctrl.Result{}))
				Expect(apiErrors.IsNotFound(returnValue.Err)).To(BeTrue())
				Expect(returnValue.ShouldReturn).To(BeTrue())
			})
		})
		Context("LabTemplate exists", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate).Build()
			})
			It("should return nil error", func() {
				returnValue := r.GetLabTemplate(ctx, testLabInstance, testLabTemplate)
				Expect(returnValue.Result).To(Equal(ctrl.Result{}))
				Expect(returnValue.Err).To(BeNil())
				Expect(returnValue.ShouldReturn).To(BeFalse())
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

})
