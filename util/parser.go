package util

import (
	// "errors"
	"fmt"
	"strings"
	"text/template"

	"sigs.k8s.io/controller-runtime/pkg/log"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
)

// TODO might use an interface like this later
// type TemplateData struct {
// 	Node       ltbv1alpha1.LabInstanceNodes
// 	Interfaces []ltbv1alpha1.NodeInterface
// }

func ParseAndRenderTemplate(nodetype *ltbv1alpha1.NodeType, renderedNodeSpec *strings.Builder, data ltbv1alpha1.LabInstanceNodes) error {
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