# API Reference

## Packages
- [ltb-backend.ltb/v1alpha1](#ltb-backendltbv1alpha1)


## ltb-backend.ltb/v1alpha1


### Resource Types
- [LabInstance](#labinstance)
- [LabTemplate](#labtemplate)
- [NodeType](#nodetype)



#### LabInstance



TODO: Explain LabInstance



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `ltb-backend.ltb/v1alpha1`
| `kind` _string_ | `LabInstance`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[LabInstanceSpec](#labinstancespec)_ |  |


#### LabInstanceNodes



Configuration for a lab node.

_Appears in:_
- [LabTemplateSpec](#labtemplatespec)

| Field | Description |
| --- | --- |
| `name` _string_ | The name of the lab node. |
| `nodeTypeRef` _[NodeTypeRef](#nodetyperef)_ | The type of the lab node. |
| `interfaces` _[NodeInterface](#nodeinterface) array_ | Array of interface configurations for the lab node. (currently not supported) |
| `config` _string_ | The configuration for the lab node. |
| `ports` _[Port](#port) array_ | Array of ports which should be publicly exposed for the lab node. |


#### LabInstanceSpec



LabInstanceSpec define which LabTemplate should be used for the lab instance and the DNS address.

_Appears in:_
- [LabInstance](#labinstance)

| Field | Description |
| --- | --- |
| `labTemplateReference` _string_ | Reference to the name of a LabTemplate to use for the lab instance. |
| `dnsAddress` _string_ | The DNS address, which will be used to expose the lab instance. It should point to the Kubernetes node where the lab instance is running. |




#### LabTemplate



TODO: add LabTemplate explanation



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `ltb-backend.ltb/v1alpha1`
| `kind` _string_ | `LabTemplate`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[LabTemplateSpec](#labtemplatespec)_ |  |


#### LabTemplateSpec



LabTemplateSpec defines the Lab nodes and their connections.

_Appears in:_
- [LabTemplate](#labtemplate)

| Field | Description |
| --- | --- |
| `nodes` _[LabInstanceNodes](#labinstancenodes) array_ | Array of lab nodes and their configuration. |
| `neighbors` _string array_ | Array of connections between lab nodes. (currently not supported) |




#### NodeInterface



Interface configuration for the lab node (currently not supported)

_Appears in:_
- [LabInstanceNodes](#labinstancenodes)

| Field | Description |
| --- | --- |
| `ipv4` _string_ | IPv4 address of the interface. |
| `ipv6` _string_ | IPv6 address of the interface. |


#### NodeType



NodeType is defines a type of node that can be used in a lab template



| Field | Description |
| --- | --- |
| `apiVersion` _string_ | `ltb-backend.ltb/v1alpha1`
| `kind` _string_ | `NodeType`
| `metadata` _[ObjectMeta](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#objectmeta-v1-meta)_ | Refer to Kubernetes API documentation for fields of `metadata`. |
| `spec` _[NodeTypeSpec](#nodetypespec)_ |  |


#### NodeTypeRef



NodeTypeRef references a NodeType with the possibility to provide additional information to the NodeType.

_Appears in:_
- [LabInstanceNodes](#labinstancenodes)

| Field | Description |
| --- | --- |
| `type` _string_ | Reference to the name of a NodeType. |
| `image` _string_ | Image to use for the NodeType. Is available as variable in the NodeType and functionality depends on its usage. |
| `version` _string_ | Version of the NodeType. Is available as variable in the NodeType and functionality depends on its usage. |


#### NodeTypeSpec



NodeTypeSpec defines the Kind and NodeSpec for a NodeType

_Appears in:_
- [NodeType](#nodetype)

| Field | Description |
| --- | --- |
| `kind` _string_ | Kind can be used to specify if the nodes is either a pod or a vm |
| `nodeSpec` _string_ | NodeSpec is the PodSpec or VirtualMachineSpec configuration for the node with the possibility to use go templating syntax to include LabTemplate variables (see [User Guide](https://lab-topology-builder.github.io/LTB-K8s-Backend/user-guide/#example-node-type)) See [PodSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#podspec-v1-core) and [VirtualMachineSpec](https://kubevirt.io/api-reference/master/definitions.html#_v1_virtualmachinespec) |




#### Port



Port of a lab node which should be publicly exposed.

_Appears in:_
- [LabInstanceNodes](#labinstancenodes)

| Field | Description |
| --- | --- |
| `name` _string_ | Arbitrary name for the port. |
| `protocol` _[Protocol](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.26/#protocol-v1-core)_ | Choose either TCP or UDP. |
| `port` _integer_ | The port number to expose. |


