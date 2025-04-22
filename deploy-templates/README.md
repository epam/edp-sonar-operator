# sonar-operator

![Version: 3.2.0](https://img.shields.io/badge/Version-3.2.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 3.2.0](https://img.shields.io/badge/AppVersion-3.2.0-informational?style=flat-square)

A Helm chart for KubeRocketCI Sonar Operator

**Homepage:** <https://docs.kuberocketci.io/>

## Overview

Sonar Operator is a KubeRocketCI operator that is responsible for configuring SonarQube.

_**NOTE:** Operator is platform-independent, that is why there is a unified instruction for deploying._

## Prerequisites

1. Linux machine or Windows Subsystem for Linux instance with [Helm 3](https://helm.sh/docs/intro/install/) installed;
2. Cluster admin access to the cluster;

## Installation

In order to install the KubeRocketCI Sonar Operator, follow the steps below:

1. To add the Helm EPAMEDP Charts for local client, run "helm repo add":

     ```bash
     helm repo add epamedp https://epam.github.io/edp-helm-charts/stable
     ```

2. Choose available Helm chart version:

     ```bash
     helm search repo epamedp/sonar-operator -l
     NAME                    CHART VERSION   APP VERSION     DESCRIPTION
     epamedp/sonar-operator  3.2.0           3.2.0           A Helm chart for KubeRocketCI Sonar Operator
     ```

    _**NOTE:** It is highly recommended to use the latest released version._

3. Deploy operator:

    Full available chart parameters available in [deploy-templates/README.md](deploy-templates/README.md):

4. Install operator in the arbitrary (`sonar-operator`) namespace with the helm command; find below the installation command example:

    ```bash
    helm install sonar-operator epamedp/sonar-operator --version <chart_version> --namespace sonar
    ```

5. Check the `sonar` namespace that should contain operator deployment with your operator in a running status.

## Quick Start

1. Login into Sonarqube and create user. Attach permissions to user such as quality gates, profiles, user managment etc. Insert user credentials into Kubernetes secret.

    ```yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name:  sonar-access
    type: Opaque
    data:
      username: dXNlcg==  # base64-encoded value of "user"
      password: cGFzcw==  # base64-encoded value of "pass"
    ```

2. Create Custom Resource `kind: Sonar` with Sonar instance URL and secret created on the previous step:

    ```yaml
    apiVersion: edp.epam.com/v1alpha1
    kind: Sonar
    metadata:
      name: sonar-sample
    spec:
      url: https://sonar.example.com   # Sonar URL
      secret: sonar-access             # Secret name
    ```

    Wait for the `.status` field with  `status.connected: true`

4. Create Quality Gate using Custom Resources SonarQualityGate:

   ```yaml
   apiVersion: edp.epam.com/v1alpha1
    kind: SonarQualityGate
    metadata:
      name: qualityGate-sample
    spec:
      sonarRef:
        name: sonar-sample # the name of `kind: Sonar`
      name: qualityGate-sample
      default: true
      conditions:
        new_coverage:
        op: LT
        error: "80"
    ```

    ```yaml
    apiVersion: edp.epam.com/v1alpha1
    kind: SonarQualityProfile
    metadata:
      name: qualityProfile-sample
    spec:
      sonarRef:
        name: sonar-sample # the name of `kind: Sonar`
      name: qualityProfile-sample
      language: java
      default: true
      rules:
        checkstyle:com.puppycrawl.tools.checkstyle.checks.OuterTypeFilenameCheck:
        severity: 'MAJOR'
    ```

    Inspect [CR templates folder](./deploy-templates/_crd_examples/) for more examples

## Local Development

In order to develop the operator, first set up a local environment. For details, please refer to the [Local Development](https://docs.kuberocketci.io/docs/developer-guide/local-development) page.

Development versions are also available, please refer to the [snapshot helm chart repository](https://epam.github.io/edp-helm-charts/snapshot/) page.

### Related Articles

- [Install KubeRocketCI](https://docs.kuberocketci.io/docs/operator-guide/install-kuberocketci)

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| epmd-edp | <SupportEPMD-EDP@epam.com> | <https://solutionshub.epam.com/solution/kuberocketci> |
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
| image.repository | string | `"epamedp/sonar-operator"` | KubeRocketCI sonar-operator Docker image name. The released image can be found on [Dockerhub](https://hub.docker.com/r/epamedp/sonar-operator) |
| image.tag | string | `nil` | KubeRocketCI sonar-operator Docker image tag. The released image can be found on [Dockerhub](https://hub.docker.com/r/epamedp/sonar-operator/tags) |
| imagePullPolicy | string | `"IfNotPresent"` |  |
| name | string | `"sonar-operator"` | component name |
| nodeSelector | object | `{}` |  |
| resources.limits.memory | string | `"192Mi"` |  |
| resources.requests.cpu | string | `"50m"` |  |
| resources.requests.memory | string | `"64Mi"` |  |
| tolerations | list | `[]` |  |
