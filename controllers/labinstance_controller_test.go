package controllers_test

import (
	"context"
	"time"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("LabInstance Controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		LabTemplateName      = "test-labtemplate"
		LabInstanceName      = "test-labinstance"
		LabInstanceNamespace = "test-namespace"
		Timeout              = time.Second * 10
		Interval             = time.Millisecond * 250
	)
	var (
		testLabInstance *ltbv1alpha1.LabInstance
		testLabTemplate *ltbv1alpha1.LabTemplate
		testPod         *corev1.Pod
		fakeClient      client.Client
		err             error
	)

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
				},
				Connections: []ltbv1alpha1.Connection{
					{
						Neighbors: "Test-Node1:1,Test-Node2:1",
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

})
