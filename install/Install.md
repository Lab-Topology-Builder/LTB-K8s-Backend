# Getting Started

## Install

1. Install OLM

```sh
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.25.0/install.sh | bash -s v0.25.0
```

1. Install the CatalogSource

```sh
kubectl apply -f https://raw.githubusercontent.com/<Path to our catalog source>
```

1. Install operator by creating a subscription

```sh
kubectl apply -f https://raw.githubusercontent.com/<Path to our subscription>
```


## Uninstall

1. Deleting the subscription

```sh
kubectl delete subscriptions.operators.coreos.com -n operators ltb-subscription
```

1. Delete the CSV

```sh
kubectl delete csv -n operators ltb-operator.<version>
```

1. Delete the CRDs

```sh
kubectl delete crd labinstances.ltb-backend.ltb labtemplates.ltb-backend.ltb nodetypes.ltb-backend.ltb
```

1. Delete operator

```sh
kubectl delete operator ltb-operator.operators
```

1. Delete the CatalogSource

```sh
kubectl delete catalogsource.operators.coreos.com -n operators ltb-catalog
```
