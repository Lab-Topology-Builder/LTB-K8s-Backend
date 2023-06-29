# API Reference

## Packages
- [ltb-backend.ltb/v1alpha1](#ltb-backendltbv1alpha1)


## ltb-backend.ltb/v1alpha1


### Resource Types
- [LabInstance](#labinstance)
- [LabTemplate](#labtemplate)
- [NodeType](#nodetype)



#### LabInstance







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `ltb-backend.ltb/v1alpha1`
| `kind` _string_ | `LabInstance`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[LabInstanceSpec](#labinstancespec)_ |  |


#### LabInstanceNodes





_Appears in:_
- [LabTemplateSpec](#labtemplatespec)

| Field | Description |
| --- | --- |
| `name` _string_ | The name of the lab node |
| `nodeTypeRef` _[NodeTypeRef](#nodetyperef)_ | The type of the lab node |
| `interfaces` _[NodeInterface](#nodeinterface) array_ | Interface configuration for the lab node (currently not supported) |
| `config` _string_ | The configuration for the lab node |
| `ports` _[Port](#port) array_ | The ports which should be publicly exposed for the lab node |


#### LabInstanceSpec





_Appears in:_
- [LabInstance](#labinstance)

| Field | Description |
| --- | --- |
| `labTemplateReference` _string_ | Reference the name of a LabTemplate |
| `dnsAddress` _string_ | The DNS address, which will be used to expose the lab instance. It should point to the Kubernetes node where the lab instance is running. |




#### LabTemplate







| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `ltb-backend.ltb/v1alpha1`
| `kind` _string_ | `LabTemplate`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[LabTemplateSpec](#labtemplatespec)_ |  |


#### LabTemplateSpec





_Appears in:_
- [LabTemplate](#labtemplate)

| Field | Description |
| --- | --- |
| `nodes` _[LabInstanceNodes](#labinstancenodes) array_ |  |
| `neighbors` _string array_ |  |




#### NodeInterface





_Appears in:_
- [LabInstanceNodes](#labinstancenodes)

| Field | Description |
| --- | --- |
| `ipv4` _string_ | IPv4 address of the interface |
| `ipv6` _string_ | IPv6 address of the interface |


#### NodeType



NodeType is the Schema for the nodetypes API



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `ltb-backend.ltb/v1alpha1`
| `kind` _string_ | `NodeType`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[NodeTypeSpec](#nodetypespec)_ |  |


#### NodeTypeRef





_Appears in:_
- [LabInstanceNodes](#labinstancenodes)

| Field | Description |
| --- | --- |
| `type` _string_ | Reference the name of a NodeType |
| `image` _string_ | Image to use for the NodeType (functionality depends on the NodeType) |
| `version` _string_ | Version of the NodeType (functionality depends on the NodeType) |


#### NodeTypeSpec



NodeTypeSpec defines the desired state of NodeType

_Appears in:_
- [NodeType](#nodetype)

| Field | Description |
| --- | --- |
| `kind` _string_ | Kind can be used to specify if the nodes is either a pod or a vm |
| `nodeSpec` _string_ | NodeSpec is the PodSpec or VirtualMachineSpec for the node See [PodSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podspec-v1-core) and [VirtualMachineSpec](https://kubevirt.io/api-reference/master/definitions.html#_v1_virtualmachinespec) |




#### Port





_Appears in:_
- [LabInstanceNodes](#labinstancenodes)

| Field | Description |
| --- | --- |
| `name` _string_ | Arbitrary name for the port |
| `protocol` _[Protocol](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#protocol-v1-core)_ | TCP or UDP |
| `port` _integer_ | The port number |


