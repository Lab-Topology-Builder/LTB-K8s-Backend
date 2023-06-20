package controllers

import (
	"context"

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

var _ = Describe("LabInstance Controller", func() {

	var r *LabInstanceReconciler

	Describe("Reconcile", func() {
		var (
			ctx context.Context
			req ctrl.Request
		)
		BeforeEach(func() {
			req = ctrl.Request{}
			ctx = context.Background()
			fakeClient = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType).Build()
			r = &LabInstanceReconciler{Client: fakeClient, Scheme: scheme.Scheme}
		})
		Context("Empty request", func() {
			It("should return NotFound error", func() {
				result, err := r.Reconcile(ctx, req)
				Expect(apiErrors.IsNotFound(err)).To(BeFalse())
				Expect(result).To(Equal(ctrl.Result{Requeue: false}))
			})
		})
		Context("Namespaced request with wrong name", func() {
			It("should return NotFound error", func() {
				req.NamespacedName = types.NamespacedName{Namespace: namespace, Name: "test"}
				result, err := r.Reconcile(ctx, req)
				Expect(apiErrors.IsNotFound(err)).To(BeFalse())
				Expect(result).To(Equal(ctrl.Result{Requeue: false}))
			})
		})

		Describe("Namespace request with correct name of available lab instance", func() {
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testPodNodeType, testNodeVMType).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testRoleBinding, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydService, testService, testVMIngress, testPodIngress).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testService, testVMIngress, testPodIngress).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType, testPodNetworkAttachmentDefinition, testVM, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType, testPodNetworkAttachmentDefinition, testVM, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testVMIngress, testPodIngress).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType, testPodNetworkAttachmentDefinition, testVM, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testService).Build()
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
					r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType, testNodeVMType, testPod, testVM, testPodNetworkAttachmentDefinition, testVMNetworkAttachmentDefinition, testServiceAccount, testRoleBinding, testRole, testTtydPod, testTtydService, testService, testVMIngress, testPodIngress).Build()
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

	Describe("ReconcileNetwork", func() {
		var (
			ctx context.Context
		)
		BeforeEach(func() {
			ctx = context.Background()
		})
		Context("labInstance nil", func() {
			It("should return error", func() {
				returnValue := r.ReconcileNetwork(ctx, nil)
				Expect(returnValue.result).To(Equal(ctrl.Result{}))
				Expect(returnValue.err).To(Equal(apiErrors.NewBadRequest("labInstance is nil")))
				Expect(returnValue.shouldReturn).To(BeTrue())
			})
		})
		Context("Valid lab instance provided", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should create a network attachment definition", func() {
				returnValue := r.ReconcileNetwork(ctx, testLabInstance)
				Expect(returnValue.result).To(Equal(ctrl.Result{Requeue: true}))
				Expect(returnValue.err).To(BeNil())
				Expect(returnValue.shouldReturn).To(BeTrue())

			})
		})
	})

	Describe("ReconcileResource", func() {
		Context("LabInstance nil", func() {
			It("Should not try to create a resource", func() {
				returnValue := r.ReconcileResource(nil, &corev1.Pod{}, nil, "test-pod")
				Expect(returnValue.result).To(Equal(ctrl.Result{}))
				Expect(returnValue.err).To(Equal(apiErrors.NewBadRequest("labInstance is nil")))
				Expect(returnValue.shouldReturn).To(BeTrue())
			})
		})
		Context("Resource already exists", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testPod).Build()
			})
			It("should not create a resource, but retrieve it", func() {
				resource := &corev1.Pod{}
				resource.Name = testLabInstance.Name + "-" + testPodNode.Name
				returnValue := r.ReconcileResource(testLabInstance, resource, nil, "")
				Expect(resource.GetName()).To(Equal(testPod.Name))
				Expect(resource.GetNamespace()).To(Equal(testPod.Namespace))
				Expect(resource.GetAnnotations()).To(Equal(testPod.Annotations))
				Expect(returnValue.result).To(Equal(ctrl.Result{}))
				Expect(returnValue.err).To(BeNil())
				Expect(returnValue.shouldReturn).To(BeFalse())
			})
		})
		Context("Resource doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("Resource can't be created, because it's the wrong type", func() {
				secret := &corev1.Secret{}
				returnValue := r.ReconcileResource(testLabInstance, secret, nil, "")
				Expect(returnValue.result).To(Equal(ctrl.Result{}))
				Expect(returnValue.err).To(HaveOccurred())
				Expect(returnValue.shouldReturn).To(BeTrue())
			})
			It("Resource can be created", func() {
				resource := &corev1.Pod{}
				resource.Name = testLabInstance.Name + "-" + testPodNode.Name
				returnValue := r.ReconcileResource(testLabInstance, &corev1.Pod{}, nil, "")
				Expect(resource).NotTo(BeNil())
				Expect(returnValue.result).To(Equal(ctrl.Result{Requeue: true}))
				Expect(returnValue.err).To(BeNil())
				Expect(returnValue.shouldReturn).To(BeTrue())
			})
		})
	})

	Describe("CreateResource", func() {
		Context("Unsupport resource type", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should return error", func() {
				secret := &corev1.Secret{}
				resource, err := CreateResource(testLabInstance, nil, secret, "")
				Expect(resource).To(Equal(secret))
				Expect(err).To(HaveOccurred())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Resource creation succeeds", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should create a VM successfully", func() {
				resource, err := CreateResource(testLabInstance, testVMNode, &kubevirtv1.VirtualMachine{}, "vm")
				Expect(resource.GetName()).To(Equal(testVM.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create a Pod successfully", func() {
				resource, err := CreateResource(testLabInstance, testPodNode, &corev1.Pod{}, "pod")
				Expect(resource.GetName()).To(Equal(testPod.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create a Service successfully", func() {
				resource, err := CreateResource(testLabInstance, testVMNode, &corev1.Service{}, "")
				Expect(resource.GetName()).To(Equal(testService.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create an Ingress successfully", func() {
				resource, err := CreateResource(testLabInstance, testVMNode, &networkingv1.Ingress{}, "vm")
				Expect(err).NotTo(HaveOccurred())
				Expect(resource.GetName()).To(Equal(testVMIngress.Name))
			})
			It("should create a service account successfully", func() {
				resource, err := CreateResource(testLabInstance, nil, &corev1.ServiceAccount{}, "")
				Expect(resource.GetName()).To(Equal(testServiceAccount.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create a role successfully", func() {
				resource, err := CreateResource(testLabInstance, nil, &rbacv1.Role{}, "")
				Expect(resource.GetName()).To(Equal(testRole.Name))
				Expect(err).NotTo(HaveOccurred())
			})
			It("should create a role binding successfully", func() {
				resource, err := CreateResource(testLabInstance, nil, &rbacv1.RoleBinding{}, "")
				Expect(resource.GetName()).To(Equal(testRoleBinding.Name))
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("ResourceExists", func() {
		Context("Resource exists", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testVM).Build()
			})
			It("should return true", func() {
				exists, err := r.ResourceExists(testVM)
				Expect(exists).To(BeTrue())
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("Resource doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should return false", func() {
				exists, err := r.ResourceExists(&corev1.Pod{})
				Expect(exists).To(BeFalse())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})

	Describe("GetLabTemplate", func() {
		var (
			ctx context.Context
		)
		BeforeEach(func() {
			ctx = context.Background()
		})
		Context("LabTemplate doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should return error", func() {
				returnValue := r.GetLabTemplate(ctx, testLabInstance, testLabTemplateWithRenderedNodeSpec)
				Expect(returnValue.result).To(Equal(ctrl.Result{}))
				Expect(apiErrors.IsNotFound(returnValue.err)).To(BeTrue())
				Expect(returnValue.shouldReturn).To(BeTrue())
			})
		})
		Context("LabTemplate exists", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec).Build()
			})
			It("should return nil error", func() {
				returnValue := r.GetLabTemplate(ctx, testLabInstance, testLabTemplateWithRenderedNodeSpec)
				Expect(returnValue.result).To(Equal(ctrl.Result{}))
				Expect(returnValue.err).To(BeNil())
				Expect(returnValue.shouldReturn).To(BeFalse())
			})
		})
	})

	Describe("GetNodeType", func() {
		var (
			ctx context.Context
		)
		BeforeEach(func() {
			ctx = context.Background()
		})
		Context("NodeType doesn't exist", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec).Build()
			})
			It("should return error", func() {
				returnValue := r.GetNodeType(ctx, &testPodNode.NodeTypeRef, testPodNodeType)
				Expect(returnValue.result).To(Equal(ctrl.Result{}))
				Expect(apiErrors.IsNotFound(returnValue.err)).To(BeTrue())
				Expect(returnValue.shouldReturn).To(BeTrue())
			})
		})
		Context("NodeType exists", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplateWithRenderedNodeSpec, testPodNodeType).Build()
			})
			It("should not return error", func() {
				returnValue := r.GetNodeType(ctx, &testPodNode.NodeTypeRef, testPodNodeType)
				Expect(returnValue.result).To(Equal(ctrl.Result{}))
				Expect(returnValue.err).To(BeNil())
				Expect(returnValue.shouldReturn).To(BeFalse())
			})
		})
	})

	Describe("MapTemplateToPod", func() {
		Context("Invalid lab instance", func() {
			It("Lab instance nil should return error", func() {
				pod, err := MapTemplateToPod(nil, testPodNode)
				Expect(pod).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Valid lab instance", func() {
			It("Node nil should return error", func() {
				pod, err := MapTemplateToPod(testLabInstance, nil)
				Expect(pod).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
			It("Error in YAML definition, should fail to unmarshal YAML", func() {
				pod, err := MapTemplateToPod(testLabInstance, podNodeYAMLProblem)
				Expect(pod).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
			It("Valid PodYaml, mapping should succeed", func() {
				pod, err := MapTemplateToPod(testLabInstance, testPodNode)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.Name).To(Equal(testPod.Name))
				Expect(pod.Namespace).To(Equal(testPod.Namespace))
				Expect(pod.Labels).To(Equal(testPod.Labels))
			})
		})
	})

	Describe("MapTemplateToVM", func() {
		Context("LabInstance nil", func() {
			It("should return error", func() {
				vm, err := MapTemplateToVM(nil, testVMNode)
				Expect(vm).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Valid lab instance", func() {
			It("Node nil should return error", func() {
				vm, err := MapTemplateToVM(testLabInstance, nil)
				Expect(vm).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
			It("Error in YAML definition, should fail to unmarshal YAML", func() {
				vm, err := MapTemplateToVM(testLabInstance, vmNodeYAMLProblem)
				Expect(vm).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
			It("Valid PodYaml, mapping should succeed", func() {
				vm, err := MapTemplateToVM(testLabInstance, testVMNode)
				Expect(err).NotTo(HaveOccurred())
				Expect(vm.Name).To(Equal(testVM.Name))
				Expect(vm.Namespace).To(Equal(testVM.Namespace))
				Expect(vm.Labels).To(Equal(testVM.Labels))
			})
		})
	})

	Describe("CreateIngress", func() {
		It("Node nil should return error", func() {
			ingress, err := CreateIngress(testLabInstance, nil, "pod")
			Expect(ingress).To(BeNil())
			Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
		})
		It("Invalid kind, should return error", func() {
			ingress, err := CreateIngress(testLabInstance, testPodNode, "invalid")
			Expect(ingress).To(BeNil())
			Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
		})
		It("Valid arguments, creation should succeed", func() {
			ingress, err := CreateIngress(testLabInstance, testPodNode, "pod")
			Expect(err).NotTo(HaveOccurred())
			Expect(ingress.Name).To(Equal(testPodIngress.Name))
			Expect(ingress.Namespace).To(Equal(testPodIngress.Namespace))
			Expect(ingress.Annotations).To(Equal(testPodIngress.Annotations))
		})
	})

	Describe("CreatePod", func() {
		Context("LabInstance nil", func() {
			It("should return error", func() {
				pod, err := CreatePod(nil, testPodNode)
				Expect(pod).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("LabInstance valid, Node nil", func() {
			It("should create ttyd pod successfully", func() {
				pod, err := CreatePod(testLabInstance, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.Name).To(Equal(testTtydPod.Name))
				Expect(pod.Namespace).To(Equal(testTtydPod.Namespace))
				Expect(pod.Labels).To(Equal(testTtydPod.Labels))
			})
		})
		Context("LabInstance and Node valid", func() {
			It("should create a pod according to the node information", func() {
				pod, err := CreatePod(testLabInstance, testPodNode)
				Expect(err).NotTo(HaveOccurred())
				Expect(pod.Name).To(Equal(testPod.Name))
				Expect(pod.Namespace).To(Equal(testPod.Namespace))
				Expect(pod.Labels).To(Equal(testPod.Labels))
			})
		})
	})

	Describe("CreateService", func() {
		Context("LabInstance nil", func() {
			It("should not create service and return BadRequest error", func() {
				service, err := CreateService(nil, testPodNode)
				Expect(service).To(BeNil())
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("Node nil, ttyd case", func() {
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
		Context("Node not nil", func() {
			It("should create a service for remote access", func() {
				service, err := CreateService(testLabInstance, testVMNode)
				Expect(err).NotTo(HaveOccurred())
				Expect(service.Name).To(Equal(testService.Name))
				Expect(service.Namespace).To(Equal(testService.Namespace))
				Expect(service.Spec.Type).To(Equal(testService.Spec.Type))
			})
		})
	})

	Describe("CreateSvcAccRoleRoleBind", func() {
		Context("LabInstance valid", func() {
			It("should create SvcAcc, role, rolebinding", func() {
				svcAcc, role, roleBind := CreateSvcAccRoleRoleBind(testLabInstance)
				Expect(svcAcc.Name).To(Equal(testServiceAccount.Name))
				Expect(role.Name).To(Equal(testRole.Name))
				Expect(roleBind.Name).To(Equal(testRoleBinding.Name))

			})
		})
	})

	Describe("UpdateLabInstanceStatus", func() {
		Context("No VMs and Pods are provided", func() {
			BeforeEach(func() {
				r.Client = fake.NewClientBuilder().WithObjects(testLabInstance).Build()
			})
			It("should fail", func() {
				err := UpdateLabInstanceStatus(nil, nil, testLabInstance)
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("LabInstance nil", func() {
			pods := []*corev1.Pod{testPod}
			vms := []*kubevirtv1.VirtualMachine{testVM}
			It("should fail", func() {
				err := UpdateLabInstanceStatus(pods, vms, nil)
				Expect(apiErrors.IsBadRequest(err)).To(BeTrue())
			})
		})
		Context("LabInstanceStatus update succeeds", func() {
			It("should have a running status", func() {
				pods := []*corev1.Pod{testPod}
				vms := []*kubevirtv1.VirtualMachine{testVM}
				err := UpdateLabInstanceStatus(pods, vms, testLabInstance)
				Expect(err).NotTo(HaveOccurred())
				Expect(testLabInstance.Status.Status).To(Equal("Running"))
				Expect(testLabInstance.Status.NumPodsRunning).To(Equal("1/1"))
				Expect(testLabInstance.Status.NumVMsRunning).To(Equal("1/1"))
			})
			It("should have a pending status", func() {
				pods := []*corev1.Pod{testPod, testPodUndefinedNode}
				vms := []*kubevirtv1.VirtualMachine{testVM}
				err := UpdateLabInstanceStatus(pods, vms, testLabInstance)
				Expect(err).NotTo(HaveOccurred())
				Expect(testLabInstance.Status.Status).To(Equal("Pending"))
				Expect(testLabInstance.Status.NumPodsRunning).To(Equal("1/2"))
				Expect(testLabInstance.Status.NumVMsRunning).To(Equal("1/1"))
			})
			It("should have a not ready status", func() {
				pods := []*corev1.Pod{testPod}
				vms := []*kubevirtv1.VirtualMachine{testVM, testVM2}
				err := UpdateLabInstanceStatus(pods, vms, testLabInstance)
				Expect(err).NotTo(HaveOccurred())
				Expect(testLabInstance.Status.Status).To(Equal("Not Ready"))
				Expect(testLabInstance.Status.NumPodsRunning).To(Equal("1/1"))
				Expect(testLabInstance.Status.NumVMsRunning).To(Equal("1/2"))
			})
		})
	})

	Describe("SetupWithManager", func() {
		It("should fail", func() {
			Expect(r.SetupWithManager(nil)).ToNot(Succeed())
		})
	})

})
