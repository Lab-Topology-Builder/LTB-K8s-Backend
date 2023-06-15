package util_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Lab-Topology-Builder/LTB-K8s-Backend/util"
)

var _ = Describe("Parser", func() {
	var (
		nodeType *ltbv1alpha1.NodeType
		data     ltbv1alpha1.LabInstanceNodes
	)
	Context("When parsing valid nodetype and data", func() {
		BeforeEach(func() {
			nodeType = &ltbv1alpha1.NodeType{
				ObjectMeta: metav1.ObjectMeta{
					Name: "GenericPodType",
				},
				Spec: ltbv1alpha1.NodeTypeSpec{
					Kind: "pod",
					NodeSpec: `
containers:
  - name: {{ .Name }}
    image: {{ .NodeTypeRef.Image}}:{{ .NodeTypeRef.Version }}
    command: ["/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"]
    ports:
      {{- range $index, $port := .Ports }}
      - name: {{ $port.Name }}
        containerPort: {{ $port.Port }}
        protocol: {{ $port.Protocol }}
      {{- end }}`,
				},
			}
			data = ltbv1alpha1.LabInstanceNodes{
				Name: "test",
				NodeTypeRef: ltbv1alpha1.NodeTypeRef{
					Type:    "testnodetype",
					Image:   "ubuntu",
					Version: "latest",
				},
				Interfaces: []ltbv1alpha1.NodeInterface{
					{
						IPv4: "192.168.0.1/24",
					},
					{
						IPv4: "172.16.0.1/24",
					},
					{
						IPv4: "10.0.0.1/24",
					},
				},
				Config: `
#cloud-config
password: ubuntu
chpasswd: { expire: False }`,
				Ports: []ltbv1alpha1.Port{
					{
						Name:     "ssh",
						Port:     22,
						Protocol: "TCP",
					},
				},
			}
		})
		It("should return nil", func() {
			err := util.ParseAndRenderTemplate(nodeType, &strings.Builder{}, data)
			Expect(err).To(BeNil())
		})
		It("should parse and render the template correctly", func() {
			var sb strings.Builder
			err := util.ParseAndRenderTemplate(nodeType, &sb, data)
			Expect(err).To(BeNil())
			Expect(sb.String()).To(MatchYAML(`
containers:
  - name: test
    image: ubuntu:latest
    command: ["/bin/bash", "-c", "apt update && apt install -y openssh-server && service ssh start && sleep 365d"]
    ports:
      - name: ssh
        containerPort: 22
        protocol: TCP
`))
		})
	})
	Context("When parsing invalid nodetype", func() {
		BeforeEach(func() {
			nodeType = &ltbv1alpha1.NodeType{
				ObjectMeta: metav1.ObjectMeta{
					Name: "GenericPodType",
				},
				Spec: ltbv1alpha1.NodeTypeSpec{
					Kind: "pod",
					NodeSpec: `
containers:
  - name: {{}{{}} .Name }}
`,
				},
			}
			data = ltbv1alpha1.LabInstanceNodes{}
		})
		It("should return an error", func() {
			err := util.ParseAndRenderTemplate(nodeType, &strings.Builder{}, data)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("Failed to parse template"))
		})
	})
	Context("When parsing empty data", func() {
		BeforeEach(func() {
			nodeType = &ltbv1alpha1.NodeType{
				ObjectMeta: metav1.ObjectMeta{
					Name: "GenericPodType",
				},
				Spec: ltbv1alpha1.NodeTypeSpec{
					Kind: "pod",
					NodeSpec: `
containers:
  - name: {{ .Name }}
    ports:
    {{- range $index, $port := .Ports }}
    - name: {{ $port.Name }}
      containerPort: {{ $port.Port }}
      protocol: {{ $port.Protocol }}
	  {{- end }}`,
				},
			}
			data = ltbv1alpha1.LabInstanceNodes{}
		})
		It("should not return an error", func() {
			err := util.ParseAndRenderTemplate(nodeType, &strings.Builder{}, data)
			Expect(err).To(BeNil())
		})
	})

})
