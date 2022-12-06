# sonar-operator

![Version: 2.13.0](https://img.shields.io/badge/Version-2.13.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 2.13.0](https://img.shields.io/badge/AppVersion-2.13.0-informational?style=flat-square)

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
| global.dnsWildCard | string | `nil` | a cluster DNS wildcard name |
| global.edpName | string | `""` | namespace or a project name (in case of OpenShift) |
| global.openshift.deploymentType | string | `"deployments"` | Wich type of kind will be deployed to Openshift (values: deployments/deploymentConfigs) |
| global.platform | string | `"openshift"` | platform type that can be "kubernetes" or "openshift" |
| image.repository | string | `"epamedp/sonar-operator"` | EDP sonar-operator Docker image name. The released image can be found on [Dockerhub](https://hub.docker.com/r/epamedp/sonar-operator) |
| image.tag | string | `nil` | EDP sonar-operator Docker image tag. The released image can be found on [Dockerhub](https://hub.docker.com/r/epamedp/sonar-operator/tags) |
| imagePullPolicy | string | `"IfNotPresent"` |  |
| name | string | `"sonar-operator"` | component name |
| nodeSelector | object | `{}` |  |
| resources.limits.memory | string | `"192Mi"` |  |
| resources.requests.cpu | string | `"50m"` |  |
| resources.requests.memory | string | `"64Mi"` |  |
| sonar.affinity | object | `{}` |  |
| sonar.annotations | object | `{}` |  |
| sonar.basePath | string | `""` | Base path for Sonar URL |
| sonar.db.affinity | object | `{}` |  |
| sonar.db.annotations | object | `{}` |  |
| sonar.db.image | string | `"postgres:9.6"` | Database image name |
| sonar.db.imagePullPolicy | string | `"IfNotPresent"` |  |
| sonar.db.nodeSelector | object | `{}` |  |
| sonar.db.resources.limits.memory | string | `"512Mi"` |  |
| sonar.db.resources.requests.cpu | string | `"50m"` |  |
| sonar.db.resources.requests.memory | string | `"64Mi"` |  |
| sonar.db.tolerations | list | `[]` |  |
| sonar.deploy | bool | `true` | Flag to enable/disable Sonar deploy |
| sonar.env | list | `[{"name":"SONAR_TELEMETRY_ENABLE","value":"false"}]` | Environment variables to attach to the sonar pod |
| sonar.image | string | `"sonarqube"` | Define sonar docker image name |
| sonar.imagePullPolicy | string | `"IfNotPresent"` |  |
| sonar.imagePullSecrets | string | `nil` | Secrets to pull from private Docker registry |
| sonar.ingress.annotations | object | `{}` |  |
| sonar.ingress.pathType | string | `"Prefix"` | pathType is only for k8s >= 1.1= |
| sonar.ingress.tls | list | `[]` | See https://kubernetes.io/blog/2020/04/02/improvements-to-the-ingress-api-in-kubernetes-1.18/#specifying-the-class-of-an-ingress ingressClassName: nginx |
| sonar.initContainers.resources | object | `{}` |  |
| sonar.initImage | string | `"busybox:1.35.0"` |  |
| sonar.name | string | `"sonar"` | Sonar name |
| sonar.nodeSelector | object | `{}` |  |
| sonar.plugins | object | `{"install":["https://github.com/vaulttec/sonar-auth-oidc/releases/download/v2.1.1/sonar-auth-oidc-plugin-2.1.1.jar","https://github.com/checkstyle/sonar-checkstyle/releases/download/9.3/checkstyle-sonar-plugin-9.3.jar","https://github.com/spotbugs/sonar-findbugs/releases/download/4.2.0/sonar-findbugs-plugin-4.2.0.jar","https://github.com/jborgers/sonar-pmd/releases/download/3.4.0/sonar-pmd-plugin-3.4.0.jar","https://github.com/sbaudoin/sonar-ansible/releases/download/v2.5.1/sonar-ansible-plugin-2.5.1.jar","https://github.com/sbaudoin/sonar-yaml/releases/download/v1.7.0/sonar-yaml-plugin-1.7.0.jar","https://github.com/Inform-Software/sonar-groovy/releases/download/1.8/sonar-groovy-plugin-1.8.jar"]}` | List of plugins to install. For example: |
| sonar.resources.limits.memory | string | `"3Gi"` |  |
| sonar.resources.requests.cpu | string | `"100m"` |  |
| sonar.resources.requests.memory | string | `"1.5Gi"` |  |
| sonar.sonarqubeFolder | string | `"/opt/sonarqube"` |  |
| sonar.storage.data.class | string | `"gp2"` | Storageclass for Sonar data volume |
| sonar.storage.data.size | string | `"1Gi"` | Size for Sonar data volume |
| sonar.storage.database.class | string | `"gp2"` | Storageclass for database data volume |
| sonar.storage.database.size | string | `"1Gi"` | Size for database data volume |
| sonar.tolerations | list | `[]` |  |
| sonar.version | string | `"8.9.10-community"` | Define sonar docker image tag |
| tolerations | list | `[]` |  |

