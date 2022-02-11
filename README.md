[![codecov](https://codecov.io/gh/epam/edp-sonar-operator/branch/master/graph/badge.svg?token=ILSDY1GF7W)](https://codecov.io/gh/epam/edp-sonar-operator)

# Sonar Operator

| :heavy_exclamation_mark: Please refer to [EDP documentation](https://epam.github.io/edp-install/) to get the notion of the main concepts and guidelines. |
| --- |

Get acquainted with the Sonar Operator and the installation process as well as the local development, and architecture scheme.

## Overview

Sonar Operator is an EDP operator that is responsible for installing and configuring SonarQube.

_**NOTE:** Operator is platform-independent, that is why there is a unified instruction for deploying._

## Prerequisites

1. Linux machine or Windows Subsystem for Linux instance with [Helm 3](https://helm.sh/docs/intro/install/) installed;
2. Cluster admin access to the cluster;
3. EDP project/namespace is deployed by following the [Install EDP](https://epam.github.io/edp-install/operator-guide/install-edp/) instruction.

## Installation

In order to install the EDP Sonar Operator, follow the steps below:

1. To add the Helm EPAMEDP Charts for local client, run "helm repo add":

     ```bash
     helm repo add epamedp https://chartmuseum.demo.edp-epam.com/
     ```

2. Choose available Helm chart version:

     ```bash
     helm search repo epamedp/sonar-operator
     NAME                    CHART VERSION   APP VERSION     DESCRIPTION
     epamedp/sonar-operator  v2.10.0                          Helm chart for Golang application/service deplo...
     ```

    _**NOTE:** It is highly recommended to use the latest released version._

3. Deploy operator:

    Full available chart parameters available in [deploy-templates/README.md](deploy-templates/README.md):

4. Install operator in the <edp_cicd_project> namespace with the helm command; find below the installation command example:

    ```bash
    helm install sonar-operator epamedp/sonar-operator --version <chart_version> --namespace <edp_cicd_project> --set name=sonar-operator --set global.edpName=<edp_cicd_project> --set global.platform=<platform_type> --set global.dnsWildCard=<cluster_DNS_wildcard>
    ```

5. Check the <edp_cicd_project> namespace that should contain operator deployment with your operator in a running status.

## Local Development

In order to develop the operator, first set up a local environment. For details, please refer to the [Developer Guide](https://epam.github.io/edp-install/developer-guide/local-development/) page.

### Related Articles

- [Architecture Scheme of Sonar Operator](docs/arch.md)
- [Install EDP](https://epam.github.io/edp-install/operator-guide/install-edp/)
