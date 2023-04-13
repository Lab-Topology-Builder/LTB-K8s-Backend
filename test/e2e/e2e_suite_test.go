package e2e

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

var (
	tc TestContext
)

var _ = BeforeSuite(func() {
	var err error
	By("By creating a new test context")
	tc, err = NewTestContext(BinaryName, "GO111MODULE=on")
	Expect(err).NotTo(HaveOccurred())

	tc.Domain = "ltb"
	tc.Group = "ltb-backend"
	tc.Version = "v1alpha1"
	tc.Kind = "LTBBackend"
	tc.Resources = "ltbbackends"
	tc.ProjectName = "operator"
	tc.Kubectl.Namespace = fmt.Sprintf("%s-system", tc.ProjectName)
	tc.Kubectl.ServiceAccount = fmt.Sprintf("%s-controller-manager", tc.ProjectName)
	dir, err := os.Getwd()
	tc.Dir = dir + "/../../"
	//tc.ImageName = fmt.Sprintf("docker.io/tsigereda/ltb-operator-test:0.1.0")

	//By("By copying the project to a temporary directory")
	//Expect(exec.Command("cp", "-r", "../../test_data", tc.Dir).Run()).To(Succeed())

	By("By adding the scorecard custom path file")
	err = tc.AddScorecardCustomPathFile()
	Expect(err).NotTo(HaveOccurred())

	//By("By using the dev image for scorecard-test")
	//err = tc.ReplaceScorecardImageForDev()
	//Expect(err).NotTo(HaveOccurred())

	By("By building the project image")
	err = tc.Make("docker-build", "IMG="+tc.ImageName)
	Expect(err).NotTo(HaveOccurred())

	onKind, err := tc.IsRunningOnKind()
	Expect(err).NotTo(HaveOccurred())
	if onKind {
		By("By loading the project image to the Kind cluster")
		Expect(tc.LoadImageToKindCluster()).To(Succeed())
		Expect(tc.LoadImageToKindClusterWithName("quay.io/operator-framework/scorecard-test:dev")).To(Succeed())
		Expect(tc.LoadImageToKindClusterWithName("quay.io/operator-framework/custom-scorecard-tests:dev")).To(Succeed())
	}

})

var _ = AfterSuite(func() {

})
