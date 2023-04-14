package controllers_test

import (
	"context"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	controller "github.com/Lab-Topology-Builder/LTB-K8s-Backend/controllers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		client = fake.NewClientBuilder().WithObjects(testLabInstance, testLabTemplate).Build()
		r = &controller.LabInstanceReconciler{Client: client}
	})

	Context("LabInstance controller functions", func() {

		It("should get the correct labtemplate", func() {
			requeue, result, err = r.GetLabTemplate(ctx, testLabInstance, testLabTemplate)
			Expect(err).NotTo(HaveOccurred())
			Expect(requeue).To(BeFalse())
			Expect(result).To(Equal(ctrl.Result{}))
			Expect(testLabTemplate.Name).To(Equal("test-labtemplate"))
		})
	})
})
