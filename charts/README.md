# kube-admission-webhook app Helm Chart

Installation

```shell
helm install my-admission-webhook --namespace test --create-namespace .
```

Upgrade

```shell
helm upgrade my-admission-webhook --namespace test --debug .
```

Uninstallation

```shell
helm uninstall my-admission-webhook
```