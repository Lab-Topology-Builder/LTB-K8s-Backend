package controllers_test

import (
	"context"
	"time"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
	var testLabInstance *ltbv1alpha1.LabInstance

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
	})

	Context("When creating a new LabInstance", func() {
		It("Should not exist before creation", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme.Scheme).Build()
			err := fakeClient.Get(context.Background(), types.NamespacedName{Name: LabInstanceName, Namespace: LabInstanceNamespace}, testLabInstance)
			By("Expecting an error when trying to get the LabInstance which does not exist")
			Expect(err).To(HaveOccurred())
			Expect(client.IgnoreNotFound(err)).To(Succeed())

			By("Expecting the LabInstance to be created")
			err = fakeClient.Create(context.Background(), testLabInstance)
			Expect(err).NotTo(HaveOccurred())

			By("Expecting the LabInstance to exist after creation")
			err = fakeClient.Get(context.Background(), types.NamespacedName{Name: LabInstanceName, Namespace: LabInstanceNamespace}, testLabInstance)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
