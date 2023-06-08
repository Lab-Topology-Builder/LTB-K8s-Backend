package controllers

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	kubevirtv1 "kubevirt.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("LabInstance Controller", func() {

	var (
		ctx context.Context
		req ctrl.Request
		r   *LabInstanceReconciler
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
				result, err := r.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
				Expect(result).To(Equal(ctrl.Result{}))
			})
		})

		Context("Namespaced request", func() {
			It("should return NotFound error", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: "test"}
				result, err := r.Reconcile(ctx, req)
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
				Expect(result).To(Equal(ctrl.Result{}))
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

	Describe("GetNodeType", func() {
		var (
			ctx context.Context
		)
		Context("NodeType doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate).Build()
			})
			It("should return error", func() {
				returnValue := r.GetNodeType(ctx, &normalPodNode.NodeTypeRef, testNodeTypePod)
				Expect(returnValue.Result).To(Equal(ctrl.Result{}))
				Expect(apiErrors.IsNotFound(returnValue.Err)).To(BeTrue())
				Expect(returnValue.ShouldReturn).To(BeTrue())
			})
		})
		Context("NodeType exists", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate, testNodeTypePod).Build()
			})
			It("should return nil error", func() {
				returnValue := r.GetNodeType(ctx, &normalPodNode.NodeTypeRef, testNodeTypePod)
				Expect(returnValue.Result).To(Equal(ctrl.Result{}))
				Expect(returnValue.Err).To(BeNil())
				Expect(returnValue.ShouldReturn).To(BeFalse())
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("MapTemplateToPod", func() {
		Context("Node nil", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should return error", func() {
				pod, err := MapTemplateToPod(testLabInstance, nil)
				Expect(pod).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Mapping fails", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should fail unmarshal problem", func() {
				pod, err := MapTemplateToPod(testLabInstance, podYAMLProblemNode)
				Expect(pod).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
		Context("Mapping succeeds", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should succeed", func() {
				pod, err := MapTemplateToPod(testLabInstance, normalPodNode)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.Name).To(Equal(testPod.Name))
				Expect(pod.Namespace).To(Equal(testPod.Namespace))
				Expect(pod.Labels).To(Equal(testPod.Labels))
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("MapTemplateToVM", func() {
		Context("LabInstance nil", func() {
			It("should return error", func() {
				vm, err := MapTemplateToVM(nil, normalVMNode)
				Expect(vm).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Node nil", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should return error", func() {
				vm, err := MapTemplateToVM(testLabInstance, nil)
				Expect(vm).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Mapping fails", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should fail unmarshal problem", func() {
				vm, err := MapTemplateToVM(testLabInstance, vmYAMLProblemNode)
				Expect(vm).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
		Context("Mapping succeeds", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should succeed", func() {
				vm, err := MapTemplateToVM(testLabInstance, normalVMNode)
				Expect(err).NotTo(HaveOccurred())
				Expect(vm.Name).To(Equal(testVM.Name))
				Expect(vm.Namespace).To(Equal(testVM.Namespace))
				Expect(vm.Labels).To(Equal(testVM.Labels))
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("CreatePod", func() {
		Context("LabInstance nil", func() {
			It("should return error", func() {
				pod, err := CreatePod(nil, normalPodNode)
				Expect(pod).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Node nil, ttyd pod creation successful", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should create ttyd pod", func() {
				pod, err := CreatePod(testLabInstance, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.Name).To(Equal(testTtydPod.Name))
				Expect(pod.Namespace).To(Equal(testTtydPod.Namespace))
				Expect(pod.Labels).To(Equal(testTtydPod.Labels))
			})
		})
		Context("Node not nil, a pod creation successful", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should succeed", func() {
				pod, err := CreatePod(testLabInstance, normalPodNode)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.Name).To(Equal(testPod.Name))
				Expect(pod.Namespace).To(Equal(testPod.Namespace))
				Expect(pod.Labels).To(Equal(testPod.Labels))
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("CreateIngress", func() {
		Context("LabInstance nil", func() {
			It("should return error", func() {
				ingress, err := CreateIngress(nil, normalPodNode)
				Expect(ingress).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Node nil", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should return error", func() {
				ingress, err := CreateIngress(testLabInstance, nil)
				Expect(ingress).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Ingress creation succeeds", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should succeed", func() {
				ingress, err := CreateIngress(testLabInstance, normalPodNode)
				Expect(err).NotTo(HaveOccurred())
				Expect(ingress.Name).To(Equal(testPodIngress.Name))
				Expect(ingress.Namespace).To(Equal(testPodIngress.Namespace))
				Expect(ingress.Annotations).To(Equal(testPodIngress.Annotations))
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("CreateService", func() {
		Context("LabInstance nil", func() {
			It("should return error", func() {
				service, err := CreateService(nil, normalPodNode)
				Expect(service).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Node nil, ttyd service creation succeeds", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should create a ttyd service", func() {
				service, err := CreateService(testLabInstance, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(service.Name).To(Equal(testTtydService.Name))
				Expect(service.Namespace).To(Equal(testTtydService.Namespace))
				Expect(service.Spec.Ports).To(Equal(testTtydService.Spec.Ports))
				Expect(service.Spec.Selector).To(Equal(testTtydService.Spec.Selector))
				Expect(service.Spec.Type).To(Equal(testTtydService.Spec.Type))
			})
		})
		Context("Node not nil, service creation succeeds", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should create a service for remote access", func() {
				service, err := CreateService(testLabInstance, normalVMNode)
				Expect(err).NotTo(HaveOccurred())
				Expect(service.Name).To(Equal(testService.Name))
				Expect(service.Namespace).To(Equal(testService.Namespace))
				Expect(service.Spec.Type).To(Equal(testService.Spec.Type))
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("CreateSvcAccRoleRoleBind", func() {
		Context("LabInstance nil", func() {
			It("should return error", func() {
				svcAcc, role, roleBind, err := CreateSvcAccRoleRoleBind(nil)
				Expect(svcAcc).To(BeNil())
				Expect(role).To(BeNil())
				Expect(roleBind).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("SvcAcc, role, rolebinding creation succeeds", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should succeed", func() {
				svcAcc, role, roleBind, err := CreateSvcAccRoleRoleBind(testLabInstance)
				Expect(err).NotTo(HaveOccurred())
				Expect(svcAcc.Name).To(Equal(testServiceAccount.Name))
				Expect(role.Name).To(Equal(testRole.Name))
				Expect(roleBind.Name).To(Equal(testRoleBinding.Name))

			})
		})

		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("UpdateLabInstanceStatus", func() {
		var (
			ctx context.Context
		)
		Context("No VMs and Pods are provided", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should fail", func() {
				err := UpdateLabInstanceStatus(ctx, nil, nil, testLabInstance)
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("LabInstance nil", func() {
			pods := []*corev1.Pod{testPod}
			vms := []*kubevirtv1.VirtualMachine{testVM}
			It("should fail", func() {
				err := UpdateLabInstanceStatus(ctx, pods, vms, nil)
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})

		Context("LabInstanceStatus update succeeds", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should have a running status", func() {
				pods := []*corev1.Pod{testPod}
				vms := []*kubevirtv1.VirtualMachine{testVM}
				err := UpdateLabInstanceStatus(ctx, pods, vms, testLabInstance)
				Expect(err).NotTo(HaveOccurred())
				Expect(testLabInstance.Status.Status).To(Equal("Running"))
				Expect(testLabInstance.Status.NumPodsRunning).To(Equal("1/1"))
				Expect(testLabInstance.Status.NumVMsRunning).To(Equal("1/1"))
			})
			It("should have a pending status", func() {
				pods := []*corev1.Pod{testPod, testNodePod}
				vms := []*kubevirtv1.VirtualMachine{testVM}
				err := UpdateLabInstanceStatus(ctx, pods, vms, testLabInstance)
				Expect(err).NotTo(HaveOccurred())
				Expect(testLabInstance.Status.Status).To(Equal("Pending"))
				Expect(testLabInstance.Status.NumPodsRunning).To(Equal("1/2"))
				Expect(testLabInstance.Status.NumVMsRunning).To(Equal("1/1"))
			})
			It("should have a not ready status", func() {
				pods := []*corev1.Pod{testPod}
				vms := []*kubevirtv1.VirtualMachine{testVM, testNodeVM}
				err := UpdateLabInstanceStatus(ctx, pods, vms, testLabInstance)
				Expect(err).NotTo(HaveOccurred())
				Expect(testLabInstance.Status.Status).To(Equal("Not Ready"))
				Expect(testLabInstance.Status.NumPodsRunning).To(Equal("1/1"))
				Expect(testLabInstance.Status.NumVMsRunning).To(Equal("1/2"))
			})
		})
		AfterEach(func() {
			r.Client = nil
		})
	})

	Describe("ErrorMsg", func() {
		var (
			ctx context.Context
		)
		Context("No error", func() {
			It("should nil error", func() {
				returnValue := ErrorMsg(ctx, nil, "Test-resource")
				Expect(returnValue.Err).To(BeNil())
				Expect(returnValue.ShouldReturn).To(BeFalse())
				Expect(returnValue.Result).To(Equal(ctrl.Result{}))
			})
		})
		Context("Error", func() {
			It("should return an error", func() {
				returnValue := ErrorMsg(ctx, apiErrors.NewNotFound(schema.GroupResource{}, "test"), "Test-resource")
				Expect(returnValue.Err).To(Equal(apiErrors.NewNotFound(schema.GroupResource{}, "test")))
				Expect(returnValue.ShouldReturn).To(BeTrue())
				Expect(returnValue.Result).To(Equal(ctrl.Result{}))
			})
		})
	})

	Describe("SetupWithManager", func() {
		It("should fail", func() {
			Expect(r.SetupWithManager(nil)).ToNot(Succeed())
		})
	})

})
