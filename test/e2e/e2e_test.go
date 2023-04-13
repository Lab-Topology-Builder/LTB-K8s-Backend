package e2e

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"time"

	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kbutil "sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
)

const namespace = "operator-system"

var saSecretTemplate = `---
apiVersion: v1
kind: Secret
type: kubernetes.io/service-account-token
metadata:
  name: %s
  annotations:
    kubernetes.io/service-account.name: "%s"
`

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
	var controllerPodName, metricsClusterRoleBindingName string

	Context("LTB Operator", func() {
		BeforeAll(func() {
			metricsClusterRoleBindingName = fmt.Sprintf("metrics-reader-%s", tc.ProjectName)

			By("By deploying the project to the cluster")
			Expect(tc.Make("deploy", "IMG="+tc.ImageName)).To(Succeed())

			//By("By creating a namespace")
			//cmd := exec.Command("kubectl", "create", "namespace", namespace)
			//_, err := RunCommand(cmd)
			//Expect(err).NotTo(HaveOccurred())
			// TODO: check this image with Jan
			//dockerImage := "docker.io/tsigereda/ltb-operator-test:0.1.0"
			//
			//By("By building the operator image")
			//cmd = exec.Command("make", "docker-build", "docker-push", fmt.Sprintf("IMG=%s", dockerImage))
			//_, err = RunCommand(cmd)
			//ExpectWithOffset(1, err).NotTo(HaveOccurred())

			//By("By installing CRDs")
			//cmd = exec.Command("make", "install")
			//_, err = RunCommand(cmd)
			//ExpectWithOffset(1, err).NotTo(HaveOccurred())
			//
			//By("By deploying the controller manager")
			//cmd = exec.Command("make", "deploy", fmt.Sprintf("IMG=%s", dockerImage))
			//_, err = RunCommand(cmd)
			//ExpectWithOffset(1, err).NotTo(HaveOccurred())

			//By("By validating the manager pod is restricted")
			//ExpectWithOffset(1, output).NotTo(ContainSubstring("Warning: would violate PodSecurityPolicy"))
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
		AfterAll(func() {
			By("By deleting the curl pod")
			WrapWarnOutput(tc.Kubectl.Delete(false, "pod", "curl"))

			By("By deleting the metrics ClusterRoleBinding")
			WrapWarnOutput(tc.Kubectl.Command("delete", "clusterrolebinding", metricsClusterRoleBindingName))

			By("By deleting namespace")
			WrapWarnOutput(tc.Kubectl.Wait(false, "namespace", namespace, "--for", "delete", "--timeout", "2m"))
			//By("By deleting the namespace")
			//cmd := exec.Command("kubectl", "delete", "namespace", namespace)
			//_, err := RunCommand(cmd)
			//Expect(err).NotTo(HaveOccurred())
		})

		It("should be running", func() {
			var err error
			directory, _ := GetProjectDir()

			By("By checking that the operator is running")
			operatorRunning := func() error {
				//cmd := exec.Command("kubectl", "get", "pods", "-l", "control-plane=controller-manager", "-n", namespace)
				podOutput, err := tc.Kubectl.Get(true, "pods", "-l", "control-plane=controller-manager", "-n", namespace)
				Expect(err).NotTo(HaveOccurred())

				podNames := kbutil.GetNonEmptyLines(podOutput)
				ExpectWithOffset(2, err).NotTo(HaveOccurred())

				controllerPodName = podNames[0]
				Expect(controllerPodName).To(ContainSubstring("controller-manager"))
				status, err := tc.Kubectl.Get(true, "pods", controllerPodName, "-n", namespace, "-o=jsonpath={.status.phase}")
				Expect(err).NotTo(HaveOccurred())
				Expect(string(status)).To(Equal("Running"))

				if string(status) != "Running" {
					return fmt.Errorf("expected controller pod to be running, got %s", string(status))
				}
				return nil
			}
			Eventually(operatorRunning, 2*time.Minute, time.Second).Should(Succeed())

			By("By ensuring the created ServiceMonitor exists")
			_, err = tc.Kubectl.Get(true, "ServiceMonitor", fmt.Sprintf("%s-controller-manger-metrics-monitor", tc.ProjectName))
			Expect(err).NotTo(HaveOccurred())

			By("By ensuring the created metrics Service exists")
			_, err = tc.Kubectl.Get(true, "Service", fmt.Sprintf("%s-controller-manager-metrics-service", tc.ProjectName))
			Expect(err).NotTo(HaveOccurred())

			By("Creating a labinstance")
			labInstanceFile := directory + "/config/samples/samples_test/ltb_v1alpha1_labinstance.yaml"
			labTemplateFile := directory + "/config/samples/samples_test/ltb_v1alpha1_labtemplate.yaml"

			Eventually(func() error {
				_, err = tc.Kubectl.Apply(true, "-f", labTemplateFile)
				Expect(err).NotTo(HaveOccurred())
				_, err = tc.Kubectl.Apply(true, "-f", labInstanceFile)
				Expect(err).NotTo(HaveOccurred())
				//cmd := exec.Command("kubectl", "apply", "-f", directory+"/config/samples/samples_test/ltb_v1alpha1_labtemplate.yaml", "-n", namespace)
				//_, err = RunCommand(cmd)
				//cmd = exec.Command("kubectl", "apply", "-f", directory+"/config/samples/samples_test/ltb_v1alpha1_labinstance.yaml", "-n", namespace)
				//_, err = RunCommand(cmd)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			By("By granting permissions to access the metrics")
			_, err = tc.Kubectl.Command("create", "clusterrolebinding", metricsClusterRoleBindingName, fmt.Sprintf("--clusterrole=%s-metrics-reader", tc.ProjectName), fmt.Sprintf("--serviceaccount=%s:%s", tc.Kubectl.Namespace, tc.Kubectl.ServiceAccount))
			Expect(err).NotTo(HaveOccurred())

			By("By creating the token")
			secreteName := tc.Kubectl.ServiceAccount + "-secret"
			fileName := directory + "/" + secreteName + ".yaml"
			err = os.WriteFile(fileName, []byte(fmt.Sprintf(saSecretTemplate, secreteName, tc.Kubectl.ServiceAccount)), 0777)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() error {
				_, err = tc.Kubectl.Apply(true, "-f", fileName)
				return err
			}, time.Minute, time.Second).Should(Succeed())

			By("By getting the token")
			query := fmt.Sprintf(`{.items[?(@.metadata.annotations.kubernetes\.io/service-account\.name=="%s")].data.token}`,
				tc.Kubectl.ServiceAccount,
			)
			b64Token, err := tc.Kubectl.Get(true, "secrets", "-o", "jsonpath="+query)
			Expect(err).NotTo(HaveOccurred())
			token, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(b64Token)))
			Expect(err).NotTo(HaveOccurred())
			Expect(len(token)).To(BeNumerically(">", 0))

			By("Checking that the labinstance is running")
			getLabInstanceStatus := func() error {
				//cmd := exec.Command("kubectl", "get", "labinstance", "-o", "jsonpath={.items[*].status.status}")
				//status, err := RunCommand(cmd)
				status, err := tc.Kubectl.Get(true, "labinstance", "-o", "jsonpath={.items[*].status.status}")
				ExpectWithOffset(2, err).NotTo(HaveOccurred())
				Expect(string(status)).To(Equal("Running"))
				return nil
			}
			EventuallyWithOffset(1, getLabInstanceStatus, time.Minute, time.Second).Should(Succeed())

		})
	})
})
