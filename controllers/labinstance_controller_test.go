package controllers_test

import (
	"context"
	"reflect"
	"time"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	Context("When creating a new LabInstance", func() {
		It("Should create a new labinstance from the labtemplate", func() {
			labInstanceLookUpKey := types.NamespacedName{Name: LabInstanceName, Namespace: LabInstanceNamespace}
			testLabInstance := &ltbv1alpha1.LabInstance{}

			By("By checking if there no labinstance exists")
			Consistently(func() (bool, error) {
				err := k8sClient.Get(context.Background(), labInstanceLookUpKey, testLabInstance)
				if err != nil {
					return false, err
				}
				return true, nil
			}, Timeout, Interval).Should(BeFalse())

			By("By creating a new LabInstance")
			testLabTemplate := &ltbv1alpha1.LabTemplate{
				ObjectMeta: metav1.ObjectMeta{
					Name:      LabTemplateName,
					Namespace: LabInstanceNamespace,
				},
				Spec: ltbv1alpha1.LabTemplateSpec{
					Nodes: []ltbv1alpha1.LabInstanceNodes{
						{
							Name: "test-node",
							Image: ltbv1alpha1.NodeImage{
								Type:    "ubuntu",
								Version: "20.04",
							},
						},
					},
				},
			}
			testPod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-pod",
					Namespace: LabInstanceNamespace,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    testLabTemplate.Spec.Nodes[0].Name,
							Image:   testLabTemplate.Spec.Nodes[0].Image.Type + ":" + testLabTemplate.Spec.Nodes[0].Image.Version,
							Command: []string{"/bin/sleep", "365d"},
						},
					},
				},
			}
			Expect(k8sClient.Create(context.Background(), testPod)).Should(Succeed())

			Eventually(func() bool {
				err := k8sClient.Get(context.Background(), labInstanceLookUpKey, testLabInstance)
				return err == nil
			}, Timeout, Interval).Should(BeTrue())

			Expect(testLabInstance).ShouldNot(BeNil())

			kind := reflect.TypeOf(testLabInstance).Name()
			gvk := ltbv1alpha1.GroupVersion.WithKind(kind)

			controllerRef := metav1.NewControllerRef(testLabInstance, gvk)
			testLabInstance.SetOwnerReferences([]metav1.OwnerReference{*controllerRef})
			Expect(k8sClient.Create(context.Background(), testLabInstance)).Should(Succeed())

			By("By checking if the labinstance is created with a pod")
			Eventually(func() ([]string, error) {
				podList := &corev1.PodList{}
				err := k8sClient.List(context.Background(), podList, client.InNamespace(LabInstanceNamespace))
				if err != nil {
					return nil, err
				}
				var podNames []string
				for _, pod := range podList.Items {
					podNames = append(podNames, pod.Name)
				}
				return podNames, nil
			}, Timeout, Interval).Should(ContainElement(testLabInstance.Name))
		})
	})
})
