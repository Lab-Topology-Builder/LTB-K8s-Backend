package e2e

import (
	"fmt"
	"os"
	"path/filepath"

	"strings"

	. "github.com/onsi/ginkgo/v2"
	kbutil "sigs.k8s.io/kubebuilder/v3/pkg/plugin/util"
	kbtestutils "sigs.k8s.io/kubebuilder/v3/test/e2e/utils"
)

const BinaryName = "operator-sdk"

const scorecardImage = "quay.io/operator-framework/scorecard-test:.*"
const scorecardImageReplace = "quay.io/operator-framework/scorecard-test:dev"

const customScorecardPatch = `
- op: add
  path: /stages/0/tests/-
  value:
    entrypoint:
	- custom-scorecard-tests
	- customtest
	image: quay.io/operator-framework/custom-scorecard-tests:dev
	labels:
		suite: custom
		test: customtest
`

const customScorecardKustomize = `
- path: patches/basic.config.yaml
  target:
    group: scorecard.operatorframework.io
	kind: Configuration
	version: v1alpha3
	name: config
`

type TestContext struct {
	*kbtestutils.TestContext
	BundleImageName string
	ProjectName     string
}

func NewTestContext(binaryName string, env ...string) (tc TestContext, err error) {
	if tc.TestContext, err = kbtestutils.NewTestContext(binaryName, env...); err != nil {
		return tc, err
	}
	tc.ProjectName = strings.ToLower(filepath.Base(tc.Dir))
	tc.ImageName = fmt.Sprintf("quay.io/%s:v0.1.0", tc.ProjectName)
	tc.BundleImageName = fmt.Sprintf("quay.io/%s-bundle:v0.1.0", tc.ProjectName)
	return tc, nil
}

func (tc TestContext) IsRunningOnKind() (bool, error) {
	kubectx, err := tc.Kubectl.Command("config", "current-context")
	if err != nil {
		return false, err
	}
	return strings.Contains(kubectx, "kind"), nil
}

func (tc TestContext) AddScorecardCustomPathFile() error {
	customScorecardPatchFile := filepath.Join(tc.Dir, "config", "scorecard", "patches", "custom.config.yaml")
	patchBytes := []byte(customScorecardPatch)
	err := os.WriteFile(customScorecardPatchFile, patchBytes, 0777)
	if err != nil {
		return err
	}

	kustomizeFile := filepath.Join(tc.Dir, "config", "scorecard", "kustomization.yaml")
	file, err := os.OpenFile(kustomizeFile, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = file.WriteString(customScorecardKustomize); err != nil {
		return err
	}
	return nil
}

func (tc TestContext) ReplaceScorecardImageForDev() error {
	err := kbutil.ReplaceRegexInFile(filepath.Join(tc.Dir, "config", "scorecard", "patches", "basic.config.yaml"), scorecardImage, scorecardImageReplace)
	if err != nil {
		return err
	}

	return nil
}

func WrapWarnOutput(_ string, err error) {
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "WARN: %s", err)
	}
}
