# Kubernetes Admission Webhook Template

This scaffold helps in building a Kubernetes Admission Webhook.

```shell
make
```

Running `make` will install a webhook server in the cluster, which:
- validates any verbs on Pod resources. If any container in the Pod contains the environment variable `DENY`, the action will be rejected.
- mutates the `CREATE` verb on Pod resources by appending the environment variable `APPEND_BY_MUTATING_WEBHOOK=yes` to all containers in the Pod.

Fork (and star) the code to start your journey!