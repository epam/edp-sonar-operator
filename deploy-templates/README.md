# sonar-operator

![Version: 3.1.1](https://img.shields.io/badge/Version-3.1.1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 3.1.1](https://img.shields.io/badge/AppVersion-3.1.1-informational?style=flat-square)

A Helm chart for EDP Sonar Operator

**Homepage:** <https://epam.github.io/edp-install/>

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| epmd-edp | <SupportEPMD-EDP@epam.com> | <https://solutionshub.epam.com/solution/epam-delivery-platform> |
| sergk |  | <https://github.com/SergK> |

## Source Code

* <https://github.com/epam/edp-sonar-operator>

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| annotations | object | `{}` |  |
| extraVolumeMounts | list | `[]` | Additional volumeMounts to be added to the container |
| extraVolumes | list | `[]` | Additional volumes to be added to the pod |
| image.repository | string | `"epamedp/sonar-operator"` | EDP sonar-operator Docker image name. The released image can be found on [Dockerhub](https://hub.docker.com/r/epamedp/sonar-operator) |
| image.tag | string | `nil` | EDP sonar-operator Docker image tag. The released image can be found on [Dockerhub](https://hub.docker.com/r/epamedp/sonar-operator/tags) |
| imagePullPolicy | string | `"IfNotPresent"` |  |
| name | string | `"sonar-operator"` | component name |
| nodeSelector | object | `{}` |  |
| resources.limits.memory | string | `"192Mi"` |  |
| resources.requests.cpu | string | `"50m"` |  |
| resources.requests.memory | string | `"64Mi"` |  |
| tolerations | list | `[]` |  |

