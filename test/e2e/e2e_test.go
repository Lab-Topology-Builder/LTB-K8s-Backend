package e2e

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const resourcesNamespace = "test"
const controllerManagerNamespace = "test-system"

func GetProjectDir() (string, error) {
	directory, err := os.Getwd()
	if err != nil {
		return directory, err
	}
	directory = strings.Replace(directory, "test/e2e", "", -1)
	return directory, nil
}

func RunCommand(cmd *exec.Cmd) ([]byte, error) {
	directory, _ := GetProjectDir()
	cmd.Dir = directory
	if err := os.Chdir(cmd.Dir); err != nil {
		fmt.Fprintf(GinkgoWriter, "Error changing directory: %v\n", err)
	}
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	command := strings.Join(cmd.Args, " ")
	fmt.Fprintf(GinkgoWriter, "Running command: %s\n", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("Error running %s: %v\n", command, err)
	}
	return output, nil
}
func GetOperatorPodName(result []byte) (string, error) {
	lines := strings.Split(string(result), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		podName := fields[0]
		if strings.Contains(podName, "controller-manager") {
			return podName, nil
		}
	}
	return "", fmt.Errorf("could not find operator pod name")

}

var _ = Describe("LTB Operator", Ordered, func() {
	Context("LTB Operator", func() {
		// TODO: check this image with Jan
		imageBase := "tsigereda/ltb-operator"
		imageVersion := "0.1.0"
		var err error
		//directory, _ := GetProjectDir()
		labInstanceFile := "../..config/samples/samples_test/ltb_v1alpha1_labinstance.yaml"
		labTemplateFile := "../..config/samples/samples_test/ltb_v1alpha1_labtemplate.yaml"
		BeforeAll(func() {

			By("By creating the controller-manager namespace")
			cmd := exec.Command("kubectl", "create", "namespace", controllerManagerNamespace)
			_, err := RunCommand(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("By creating the resources namespace")
			cmd = exec.Command("kubectl", "create", "namespace", resourcesNamespace)
			_, err = RunCommand(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("By setting the env NAMESPACE")
			err = os.Setenv("NAMESPACE", controllerManagerNamespace)
			Expect(err).NotTo(HaveOccurred())

			By("By setting the env WATCH_NAMESPACE")
			err = os.Setenv("WATCH_NAMESPACE", resourcesNamespace)
			Expect(err).NotTo(HaveOccurred())

			By("By setting the env IMAGE_TAG_BASE")
			err = os.Setenv("IMAGE_TAG_BASE", imageBase)
			Expect(err).NotTo(HaveOccurred())

			By("By setting the env VERSION")
			err = os.Setenv("VERSION", imageVersion)
			Expect(err).NotTo(HaveOccurred())

			By("By setting the env IMG")
			err = os.Setenv("IMG", os.Getenv("IMAGE_TAG_BASE")+":"+os.Getenv("VERSION"))
			Expect(err).NotTo(HaveOccurred())

		})
		AfterAll(func() {
			By("By deleting the env NAMESPACE")
			err = os.Unsetenv("NAMESPACE")
			Expect(err).NotTo(HaveOccurred())

			By("By deleting the env WATCH_NAMESPACE")
			err := os.Unsetenv("WATCH_NAMESPACE")
			Expect(err).NotTo(HaveOccurred())

			By("By deleting the env IMAGE_TAG_BASE")
			err = os.Unsetenv("IMAGE_TAG_BASE")
			Expect(err).NotTo(HaveOccurred())

			By("By deleting the env VERSION")
			err = os.Unsetenv("VERSION")
			Expect(err).NotTo(HaveOccurred())

			By("By deleting the env IMG")
			err = os.Unsetenv("IMG")
			Expect(err).NotTo(HaveOccurred())

			By("By deleting the namespace")
			cmd := exec.Command("kubectl", "delete", "namespace", controllerManagerNamespace)
			_, err = RunCommand(cmd)
			Expect(err).NotTo(HaveOccurred())

			By("By deleting the namespace")
			cmd = exec.Command("kubectl", "delete", "namespace", resourcesNamespace)
			_, err = RunCommand(cmd)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should be running", func() {

			By("By installing CRDs")
			cmd := exec.Command("make", "install")
			_, err = RunCommand(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("By deploying the controller manager")
			cmd = exec.Command("make", "deploy")
			_, err = RunCommand(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("By checking that the controller manager is running")
			controllerManagerRunning := func() error {
				cmd := exec.Command("kubectl", "get", "pods", "-l", "control-plane=controller-manager", "-n", "operator-system")
				podOutput, err := RunCommand(cmd)
				fmt.Fprintf(GinkgoWriter, "podOutput: %s\n", podOutput)
				Expect(err).NotTo(HaveOccurred())

				podName, err := GetOperatorPodName(podOutput)
				fmt.Fprintf(GinkgoWriter, "podName: %s\n", podName)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				status, err := exec.Command("kubectl", "get", "pods", podName, "-n", "operator-system", "-o", "jsonpath={.status.phase}").Output()
				Expect(err).NotTo(HaveOccurred())
				Expect(string(status)).To(Equal("Running"))

				return nil
			}
			Eventually(controllerManagerRunning, 2*time.Minute, time.Second).Should(Succeed())

			By("By creating a lab template")
			Eventually(func() error {
				cmd := exec.Command("kubectl", "apply", "-f", labTemplateFile)
				_, err = RunCommand(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			By("By creating a lab instance")
			Eventually(func() error {
				cmd = exec.Command("kubectl", "apply", "-f", labInstanceFile)
				_, err = RunCommand(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking that the labinstance is running")
			getLabInstanceStatus := func() error {
				cmd := exec.Command("kubectl", "get", "labinstance", "-o", "jsonpath={.items[*].status.status}")
				status, err := RunCommand(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				Expect(string(status)).To(Equal("Running"))
				return nil
			}
			Eventually(getLabInstanceStatus, 2*time.Minute, time.Second).Should(Succeed())

		})
	})
})
