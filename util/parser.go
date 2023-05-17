package util

import (
	// "errors"
	"fmt"
	"strings"
	"text/template"

	"sigs.k8s.io/controller-runtime/pkg/log"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
)

type TemplateData struct {
	Node        ltbv1alpha1.LabInstanceNodes
	Connections []ltbv1alpha1.Connection
}

func ParseAndRenderTemplate(nodetype *ltbv1alpha1.NodeType, renderedNodeSpec *strings.Builder, data TemplateData) error {
	tmplt, err := template.New("nodeTemplate").Parse(nodetype.Spec.NodeSpec)
	if err != nil {
		return err
		// return errors.New("ParseAndRenderTemplate: Failed to parse template")
	}
	err = tmplt.Execute(renderedNodeSpec, data)
	if err != nil {
		return err
		// return errors.New("ParseAndRenderTemplate: Failed to render template")
	}
	log.Log.Info(fmt.Sprintf("Rendered Template: %s", renderedNodeSpec.String()))
	return nil
}