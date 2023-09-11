[![codecov](https://codecov.io/gh/epam/edp-sonar-operator/branch/master/graph/badge.svg?token=ILSDY1GF7W)](https://codecov.io/gh/epam/edp-sonar-operator)

# Sonar Operator

| :heavy_exclamation_mark: Please refer to [EDP documentation](https://epam.github.io/edp-install/) to get the notion of the main concepts and guidelines. |
| --- |

Get acquainted with the Sonar Operator and the installation process as well as the local development, and architecture scheme.

## Overview

Sonar Operator is an EDP operator that is responsible for configuring SonarQube.

_**NOTE:** Operator is platform-independent, that is why there is a unified instruction for deploying._

## Prerequisites

1. Linux machine or Windows Subsystem for Linux instance with [Helm 3](https://helm.sh/docs/intro/install/) installed;
2. Cluster admin access to the cluster;
3. EDP project/namespace is deployed by following the [Install EDP](https://epam.github.io/edp-install/operator-guide/install-edp/) instruction.

## Installation

In order to install the EDP Sonar Operator, follow the steps below:

1. To add the Helm EPAMEDP Charts for local client, run "helm repo add":

     ```bash
     helm repo add epamedp https://epam.github.io/edp-helm-charts/stable
     ```

2. Choose available Helm chart version:

     ```bash
     helm search repo epamedp/sonar-operator -l
     NAME                    CHART VERSION   APP VERSION     DESCRIPTION
     epamedp/sonar-operator  3.1.0           3.1.0           A Helm chart for EDP Sonar Operator
     epamedp/sonar-operator  3.0.0           3.0.0           A Helm chart for EDP Sonar Operator
     ```

    _**NOTE:** It is highly recommended to use the latest released version._

3. Deploy operator:

    Full available chart parameters available in [deploy-templates/README.md](deploy-templates/README.md):

4. Install operator in the arbitrary (`sonar-operator`) namespace with the helm command; find below the installation command example:

    ```bash
    helm install sonar-operator epamedp/sonar-operator --version <chart_version> --namespace sonar-operator
    ```

5. Check the `sonar-operator` namespace that should contain operator deployment with your operator in a running status.

## Quick Start

1. Insert newly created user credentials into Kubernetes secret:

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
      url: https://sonar.example.com/ # Sonar URL
      secret: sonar-access  # Secret name
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

In order to develop the operator, first set up a local environment. For details, please refer to the [Local Development](https://epam.github.io/edp-install/developer-guide/local-development/) page.

Development versions are also available, please refer to the [snapshot helm chart repository](https://epam.github.io/edp-helm-charts/snapshot/) page.

### Related Articles

- [Install EDP](https://epam.github.io/edp-install/operator-guide/install-edp/)
