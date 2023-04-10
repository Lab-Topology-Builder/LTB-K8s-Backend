package controllers_test

import (
	"context"
	"time"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	kubevirtv1 "kubevirt.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// TODO: Test the status of the pods and VMs

var _ = Describe("LabInstance Controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		LabTemplateName      = "test-labtemplate"
		LabInstanceName      = "test-labinstance"
		LabInstanceNamespace = "test-namespace"
		Timeout              = time.Second * 10
		Interval             = time.Millisecond * 250
		config               = "#cloud-config\npassword: ubuntu\nchpasswd:\n{ expire: False }\nssh_authorized_keys:\n- <your-ssh-pub-key>\npackages:\n- qemu-guest-agent\nruncmd\n- [ systemctl, start, qemu-guest-agent ]"
	)
	var (
		testLabInstance *ltbv1alpha1.LabInstance
		testLabTemplate *ltbv1alpha1.LabTemplate
		testPod         *corev1.Pod
		testVM          *kubevirtv1.VirtualMachine
		fakeClient      client.Client
		err             error
		running         bool
	)

	schemeBuilder := runtime.NewSchemeBuilder(kubevirtv1.AddToScheme)
	vmScheme := runtime.NewScheme()
	err = schemeBuilder.AddToScheme(vmScheme)
	Expect(err).NotTo(HaveOccurred())

	BeforeEach(func() {

		testLabInstance = &ltbv1alpha1.LabInstance{
			ObjectMeta: metav1.ObjectMeta{
				Name:      LabInstanceName,
				Namespace: LabInstanceNamespace,
			},
			Spec: ltbv1alpha1.LabInstanceSpec{
				LabTemplateReference: LabTemplateName,
			},
		}

		testLabTemplate = &ltbv1alpha1.LabTemplate{
			ObjectMeta: metav1.ObjectMeta{
				Name:      LabTemplateName,
				Namespace: LabInstanceNamespace,
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
						Config: config,
					},
				},
			},
		}

		testPod = &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testLabTemplate.Spec.Nodes[0].Name,
				Namespace: testLabInstance.Namespace,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:    testLabTemplate.Spec.Nodes[0].Name,
						Image:   testLabTemplate.Spec.Nodes[0].Image.Type + ":" + testLabTemplate.Spec.Nodes[0].Image.Version,
						Command: []string{"/bin/sleep", "3600"},
					},
				},
			},
		}

		running = true
		resources := kubevirtv1.ResourceRequirements{
			Requests: corev1.ResourceList{"memory": resource.MustParse("1024M")},
		}
		cpu := &kubevirtv1.CPU{Cores: 1}

		disk := []kubevirtv1.Disk{
			{Name: "containerdisk",
				DiskDevice: kubevirtv1.DiskDevice{
					Disk: &kubevirtv1.DiskTarget{Bus: "virtio"}}},
			{Name: "cloudinitdisk", DiskDevice: kubevirtv1.DiskDevice{
				Disk: &kubevirtv1.DiskTarget{Bus: "virtio"}}},
		}

		volumes := []kubevirtv1.Volume{
			{Name: "containerdisk",
				VolumeSource: kubevirtv1.VolumeSource{
					ContainerDisk: &kubevirtv1.ContainerDiskSource{Image: "quay.io/containerdisks/" + testLabTemplate.Spec.Nodes[1].Image.Type + ":" + testLabTemplate.Spec.Nodes[1].Image.Version}}},
			{Name: "cloudinitdisk", VolumeSource: kubevirtv1.VolumeSource{
				CloudInitNoCloud: &kubevirtv1.CloudInitNoCloudSource{UserData: testLabTemplate.Spec.Nodes[1].Config}}},
		}

		testVM = &kubevirtv1.VirtualMachine{
			ObjectMeta: metav1.ObjectMeta{
				Name:      testLabTemplate.Spec.Nodes[1].Name,
				Namespace: testLabInstance.Namespace,
			},
			Spec: kubevirtv1.VirtualMachineSpec{
				Running: &running,
				Template: &kubevirtv1.VirtualMachineInstanceTemplateSpec{
					Spec: kubevirtv1.VirtualMachineInstanceSpec{
						Domain: kubevirtv1.DomainSpec{
							Resources: resources,
							CPU:       cpu,
							Devices: kubevirtv1.Devices{
								Disks: disk,
							},
						},
						Volumes: volumes,
					},
				},
			},
		}

	})

	Context("Pod within a LabInstance", func() {
		It("Should not exist before creation", func() {
			fakeClient = fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
			err = fakeClient.Get(context.Background(), types.NamespacedName{Name: testPod.Name, Namespace: testLabInstance.Namespace}, testPod)
			By("Expecting an error when trying to get the Pod which does not exist")
			Expect(err).To(HaveOccurred())
			Expect(client.IgnoreNotFound(err)).To(Succeed())
		})

		It("Should exist after creation", func() {
			By("Expecting the Pod to be created")
			err = fakeClient.Create(context.Background(), testPod)
			Expect(err).NotTo(HaveOccurred())
			Expect(err).To(Succeed())
		})

		It("Should have a status after creation", func() {
			By("Expecting the Pod to have a status")
			testPodStatus := testPod.Status.Phase
			Expect(testPodStatus).NotTo(BeNil())
		})

		It("Should be deleted after deletion", func() {
			By("Expecting the Pod to be deleted")
			err = fakeClient.Delete(context.Background(), testPod)
			Expect(err).NotTo(HaveOccurred())
			Expect(err).To(Succeed())
		})
	})

	Context("VirtualMachine within a LabInstance", func() {
		It("Should not exist before creation", func() {
			fakeClient = fake.NewClientBuilder().WithScheme(vmScheme).Build()
			err = fakeClient.Get(context.Background(), types.NamespacedName{Name: testVM.Name, Namespace: testLabInstance.Namespace}, testVM)
			By("Expecting an error when trying to get the VirtualMachine which does not exist")
			Expect(err).To(HaveOccurred())
			Expect(client.IgnoreNotFound(err)).To(Succeed())
		})

		It("Should exist after creation", func() {
			By("Expecting the VirtualMachine to be created")
			err = fakeClient.Create(context.Background(), testVM)
			Expect(err).NotTo(HaveOccurred())
			Expect(err).To(Succeed())
		})

		It("Should have a status after creation", func() {
			By("Expecting the VirtualMachine to have a status")
			testVMStatus := testVM.Status.Ready
			Expect(testVMStatus).NotTo(BeNil())
		})

		It("Should be deleted after deletion", func() {
			By("Expecting the VirtualMachine to be deleted")
			err = fakeClient.Delete(context.Background(), testVM)
			Expect(err).NotTo(HaveOccurred())
			Expect(err).To(Succeed())
		})
	})

})
