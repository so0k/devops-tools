# DataDog Pod sizer

Query container cpu / memory usage from DataDog queries

Uses DataDog Agent 6.1 with following label whitelisting

```yaml
   - name: DD_KUBERNETES_POD_LABELS_AS_TAGS # whilelist relevant labels
    value: '{"app":"helm_app","release":"helm_release","component":"helm_component","k8s-app":"k8s-app","chart":"helm_chart","heritage":"helm_heritage"}'
```

The `kube_container_name` is automatically captured by the DataDog agent.
