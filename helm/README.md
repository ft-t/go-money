## Values

| key                            | type                   | Description                                                                                                                               |
|--------------------------------|------------------------|-------------------------------------------------------------------------------------------------------------------------------------------|
| `env`                          | map[string]interface{} | dictionary of key-value env variables                                                                                                     |
| `envFrom.secrets`              | []string               | string array of secrets that should be mounted on pod startup                                                                             |
| `envFrom.configMaps`           | []string               | string array of config maps that should be mounted on pod startup                                                                         |
| `image.repository`             | string                 | image repository                                                                                                                          |
| `image.tag`                    | string                 | image tag                                                                                                                                 |
| `serviceMonitoring.enabled`    | bool                   | prometheus metrics exporter. disabled by default                                                                                          |
| `serviceMonitoring.apiVersion` | string                 | prometheus metrics exporter. api version. If you dont use prometheus for metrics, but lets say VictoriaMetrics, swap to proper apiVersion |
| `serviceMonitoring.kind`       | string                 | prometheus metrics exporter. kind. If you dont use prometheus for metrics, but lets say VictoriaMetrics, swap to proper kind              |
