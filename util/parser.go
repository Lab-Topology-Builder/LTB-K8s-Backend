package util

import (
	"fmt"
	"strings"
	"text/template"

	"sigs.k8s.io/controller-runtime/pkg/log"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
)

func ParseAndRenderTemplate(nodetype *ltbv1alpha1.NodeType, renderedNodeSpec *strings.Builder, data ltbv1alpha1.LabInstanceNodes) error {
	tmplt, err := template.New("nodeTemplate").Parse(nodetype.Spec.NodeSpec)

	if err != nil {
		return fmt.Errorf("ParseAndRenderTemplate: Failed to parse template\nErr:%s", err)
	}
	err = tmplt.Execute(renderedNodeSpec, data)
	if err != nil {
		return fmt.Errorf("ParseAndRenderTemplate: Failed to render template\nErr:%s", err)
	}
	log.Log.Info(fmt.Sprintf("Rendered Template: %s", renderedNodeSpec.String()))
	return nil
}
