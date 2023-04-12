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

const namespace = "test-namespace"

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
		return output, fmt.Errorf("error running %s: %v\n", command, err)
	}
	return output, nil
}

func LoadImageToClusterWithName(imageName string) error {
	cluster := "kind"
	if value, ok := os.LookupEnv("KIND_CLUSTER_NAME"); ok {
		cluster = value
	}
	kindOptions := []string{"load", "docker-image", imageName, "--name", cluster}
	cmd := exec.Command("kind", kindOptions...)
	_, err := RunCommand(cmd)
	return err
}

func GetNonEmptyLines(result string) []string {
	var lines []string
	elements := strings.Split(result, "\n")
	for _, element := range elements {
		if element != "" {
			lines = append(lines, element)
		}
	}
	return lines
}

var _ = Describe("LTB Operator", Ordered, func() {
	BeforeAll(func() {
		By("By creating a namespace")
		cmd := exec.Command("kubectl", "create", "namespace", namespace)
		_, err := RunCommand(cmd)
		Expect(err).NotTo(HaveOccurred())
		//By("By labeling all namespaces to warn about PodSecurityPolicy violations")
		//cmd = exec.Command("kubectl", "label", "--overwrite", "ns", "--all",
		//	"pod-security.kubernetes.io/audit=restricted",
		//	"pod-security.kubernetes.io/warn=restricted")
		//_, err = RunCommand(cmd)
		//Expect(err).NotTo(HaveOccurred())
		//
		//By("By labeling the namespace the operator will be deployed in to enforce the PodSecurityPolicy")
		//cmd = exec.Command("kubectl", "label", "--overwrite", "ns", namespace,
		//	"pod-security.kubernetes.io/audit=restricted",
		//	"pod-security.kubernetes.io/enforce=restricted")
		//_, err = RunCommand(cmd)
		//Expect(err).NotTo(HaveOccurred())

	})

	Context("LTB Operator", func() {
		It("should be running", func() {
			var controllerPodName string
			var err error
			directory, _ := GetProjectDir()
			dockerImage := "ltb/operator:0.1.0"

			//By("By building the operator image")
			//cmd := exec.Command("make", "docker-build", fmt.Sprintf("IMG=%s", dockerImage))
			//_, err = RunCommand(cmd)
			//ExpectWithOffset(1, err).NotTo(HaveOccurred())
			//
			//By("By loading the operator image into the kind cluster")
			//err = LoadImageToClusterWithName(dockerImage)
			//ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("By installing CRDs")
			cmd := exec.Command("make", "install")
			_, err = RunCommand(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("By deploying the controller manager")
			cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", dockerImage))
			output, err := RunCommand(cmd)
			ExpectWithOffset(1, err).NotTo(HaveOccurred())

			By("By validating the manager pod is restricted")
			ExpectWithOffset(1, output).NotTo(ContainSubstring("Warning: would violate PodSecurityPolicy"))

			By("By checking that the controller manager pod is running")
			controllerRunning := func() error {
				cmd = exec.Command("kubectl", "get", "pods", "-n", namespace, "-l", "control-plane=controller-manager", "-o", "jsonpath={.items[*].status.phase}")
				podOutput, err := RunCommand(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				podNames := GetNonEmptyLines(string(podOutput))
				if len(podNames) != 1 {
					return fmt.Errorf("expected 1 pod, got %d", len(podNames))
				}
				controllerPodName = podNames[0]
				ExpectWithOffset(2, controllerPodName).Should(ContainSubstring("controller-manager"))

				cmd = exec.Command("kubectl", "get", "pods", "-n", namespace, "-o", "jsonpath={.status.phase}")
				status, err := RunCommand(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != "Running" {
					return fmt.Errorf("expected controller pod to be running, got %s", string(status))
				}
				return nil
			}
			EventuallyWithOffset(1, controllerRunning, time.Minute, time.Second).Should(Succeed())

			By("Creating a labinstance")
			EventuallyWithOffset(1, func() error {
				cmd = exec.Command("kubectl", "apply", "-f", directory+"/config/samples/ltb_v1alpha1_labtemplate.yaml")
				_, err = RunCommand(cmd)
				cmd = exec.Command("kubectl", "apply", "-f", directory+"/config/samples/ltb_v1alpha1_labinstance.yaml")
				_, err = RunCommand(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking that the labinstance is running")
			getLabInstanceStatus := func() error {
				cmd = exec.Command("kubectl", "get", "labinstance", "-o", "jsonpath={.items[*].status.status}")
				status, err := RunCommand(cmd)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				if string(status) != "Running" {
					return fmt.Errorf("expected labinstance to be running, got %s", string(status))
				}
				return nil
			}
			EventuallyWithOffset(1, getLabInstanceStatus, time.Minute, time.Second).Should(Succeed())

		})
	})
	AfterAll(func() {
		By("By deleting the namespace")
		cmd := exec.Command("kubectl", "delete", "namespace", namespace)
		_, err := RunCommand(cmd)
		Expect(err).NotTo(HaveOccurred())
	})
})
