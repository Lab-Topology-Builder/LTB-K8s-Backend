package controllers_test

import (
	"context"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	controller "github.com/Lab-Topology-Builder/LTB-K8s-Backend/controllers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubevirtv1 "kubevirt.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("LabInstance Controller", func() {
	var (
		ctx             context.Context
		r               *controller.LabInstanceReconciler
		testLabInstance *ltbv1alpha1.LabInstance
		testLabTemplate *ltbv1alpha1.LabTemplate
		result          ctrl.Result
		err             error
		requeue         bool
		client          client.Client
		testPod         *corev1.Pod
		testVM          *kubevirtv1.VirtualMachine
		podNode, vmNode *ltbv1alpha1.LabInstanceNodes
		running         bool
	)

	BeforeEach(func() {
		ctx = context.Background()
		testLabInstance = &ltbv1alpha1.LabInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-labinstance",
			},
			Spec: ltbv1alpha1.LabInstanceSpec{
				LabTemplateReference: "test-labtemplate",
			},
		}
		testLabTemplate = &ltbv1alpha1.LabTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-labtemplate",
			},
			Spec: ltbv1alpha1.LabTemplateSpec{
				Nodes: []ltbv1alpha1.LabInstanceNodes{
					{
						Name: "test-node-1",
						Image: ltbv1alpha1.NodeImage{
							Type:    "ubuntu",
							Version: "20.04",
						},
					},
					{
						Name: "test-node-2",
						Image: ltbv1alpha1.NodeImage{
							Type:    "ubuntu",
							Version: "20.04",
							Kind:    "vm",
						},
					},
				},
			},
		}

		podNode = &testLabTemplate.Spec.Nodes[0]
		testPod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: podNode.Name,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    podNode.Name,
						Image:   podNode.Image.Type + ":" + podNode.Image.Version,
						Command: []string{"/bin/sleep", "365d"},
					},
				},
			},
		}

		vmNode = &testLabTemplate.Spec.Nodes[1]
		running = true
		resources := kubevirtv1.ResourceRequirements{
			Requests: corev1.ResourceList{"memory": resource.MustParse("2048M")},
		}
		cpu := &kubevirtv1.CPU{Cores: 1}
		metadata := metav1.ObjectMeta{
			Name: vmNode.Name,
		}
		disks := []kubevirtv1.Disk{
			{Name: "containerdisk", DiskDevice: kubevirtv1.DiskDevice{Disk: &kubevirtv1.DiskTarget{Bus: "virtio"}}},
			{Name: "cloudinitdisk", DiskDevice: kubevirtv1.DiskDevice{Disk: &kubevirtv1.DiskTarget{Bus: "virtio"}}},
		}
		volumes := []kubevirtv1.Volume{
			{Name: "containerdisk", VolumeSource: kubevirtv1.VolumeSource{ContainerDisk: &kubevirtv1.ContainerDiskSource{Image: "quay.io/containerdisks/" + vmNode.Image.Type + ":" + vmNode.Image.Version}}},
			{Name: "cloudinitdisk", VolumeSource: kubevirtv1.VolumeSource{CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{UserData: vmNode.Config}}}}

		testVM = &kubevirtv1.VirtualMachine{
			ObjectMeta: metadata,
			Spec: kubevirtv1.VirtualMachineSpec{
				Running: &running,
				Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
					Spec: kubevirtv1.VirtualMachineInstanceSpec{
						Domain: kubevirtv1.DomainSpec{
							Resources: resources,
							CPU:       cpu,
							Devices: kubevirtv1.Devices{
								Disks: disks,
							},
						},
						Volumes: volumes,
					},
				},
			},
		}

		client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate).Build()
		r = &controller.LabInstanceReconciler{Client: client}
	})

	Context("LabInstance controller template functions", func() {

		It("should get the correct labtemplate", func() {
			requeue, result, err = r.GetLabTemplate(ctx, testLabInstance, testLabTemplate)
			Expect(err).NotTo(HaveOccurred())
			Expect(requeue).To(BeFalse())
			Expect(result).To(Equal(ctrl.Result{}))
			Expect(testLabTemplate.Name).To(Equal("test-labtemplate"))
		})

		It("should map labtemplate to pod", func() {
			podNode = &testLabTemplate.Spec.Nodes[0]
			testPod = r.MapTemplateToPod(testLabInstance, podNode)
			Expect(testPod.Name).To(Equal("test-node-1"))
			Expect(testPod.Spec.Containers[0].Name).To(Equal("test-node-1"))
			Expect(testPod.Spec.Containers[0].Image).To(Equal("ubuntu:20.04"))
		})

		It("should map labtemplate to vm", func() {
			vmNode = &testLabTemplate.Spec.Nodes[1]
			testVM = r.MapTemplateToVM(testLabInstance, vmNode)
			Expect(testVM.Name).To(Equal("test-node-2"))
			Expect(testVM.Spec.Template.Spec.Domain.Resources.Requests.Memory().String()).To(Equal("2048M"))
			Expect(testVM.Spec.Template.Spec.Domain.CPU.Cores).To(Equal(uint32(1)))
			Expect(testVM.Spec.Template.Spec.Volumes[0].Name).To(Equal("containerdisk"))
			Expect(testVM.Spec.Template.Spec.Volumes[0].VolumeSource.ContainerDisk.Image).To(Equal("quay.io/containerdisks/ubuntu:20.04"))
			Expect(testVM.Spec.Template.Spec.Volumes[1].Name).To(Equal("cloudinitdisk"))
			Expect(testVM.Spec.Template.Spec.Domain.Resources.Requests.Memory().String()).ToNot(BeEmpty())
			Expect(testVM.Spec.Running).To(Equal(&running))

		})
	})

	// Unable to test reconcile functions because it requires deployment of the pods and vms
})
