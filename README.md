# Sonar Operator

Get acquainted with the Sonar Operator and the installation process as well as the local development, 
and architecture scheme.
## Overview

Sonar Operator is an EDP operator that is responsible for installing and configuring SonarQube.

_**NOTE:** Operator is platform-independent, that is why there is a unified instruction for deploying._

## Prerequisites
1. Linux machine or Windows Subsystem for Linux instance with [Helm 3](https://helm.sh/docs/intro/install/) installed;
2. Cluster admin access to the cluster;
3. EDP project/namespace is deployed by following one of the instructions: [edp-install-openshift](https://github.com/epmd-edp/edp-install/blob/master/documentation/openshift_install_edp.md#edp-project) or [edp-install-kubernetes](https://github.com/epmd-edp/edp-install/blob/master/documentation/kubernetes_install_edp.md#edp-namespace).

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
     epamedp/sonar-operator  v2.4.0                          Helm chart for Golang application/service deplo...
     ```

    _**NOTE:** It is highly recommended to use the latest released version._
    
3. Create secrets:

Openshift
```bash
oc -n <global.edpName> create secret generic sonar-db --from-literal=database-user=admin --from-literal=database-password=<password>
oc -n <global.edpName> create secret generic sonar-admin-password --from-literal=username=admin --from-literal=password=<password>
```

Kubernetes: 
```bash
kubectl -n <global.edpName> create secret generic sonar-db --from-literal=database-user=admin --from-literal=database-password=<password>
kubectl -n <global.edpName> create secret generic sonar-admin-password --from-literal=username=admin --from-literal=password=<password>
```

4. Deploy operator:

    Full available chart parameters list:
    ```
    - <chart_version>                        # Helm chart version;
    - global.edpName                         # a namespace or a project name (in case of OpenShift);
    - global.platform                        # a platform type that can be "kubernetes" or "openshift";
    - global.dnsWildCard                     # a cluster DNS wildcard name;
    - image.name                             # EDP sonar-oprator Docker image name. The released image can be found on https://hub.docker.com/r/epamedp/sonar-operator;
    - image.version                          # EDP sonar-oprator Docker image tag. The released image can be found on https://hub.docker.com/r/epamedp/sonar-operator/tags;
    - sonar.deploy                           # If true Sonarqube CR will be added and Sonarqube instance will be deployed
    - sonar.name                             # Sonar custom resource name
    - sonar.image                            # Sonarqube Docker image name. Default supported is "sonarqube";
    - sonar.version                          # Sonarqube Docker image tag. Default supported is "7.9-community";
    - sonar.initImage                        # Init Docker image for Sonarqube deployment. Default is "busybox";
    - sonar.dbImage                          # Docker image name for Sonarqube Database. Default in "postgres:9.6";
    - sonar.dataVolumeStorageClass           # Storageclass for Sonarqube data volume. Default is "gp2";
    - sonar.dataVolumeCapacity               # Sonarqube data volume capacity. Default is "1Gi";
    - sonar.dbVolumeStorageClass             # Storageclass for Sonarqube database volume. Default is "gp2";
    - sonar.dbVolumeCapacity                 # Sonarqube database volume capacity. Default is "1Gi".
    - sonar.imagePullSecrets                 # Secrets to pull from private Docker registry;
    - sonar.basePath                         # Base path for Sonarqube URL;
    - sonar.storage.data.class               # Storageclass for Sonar data volume. Default is "gp2";
    - sonar.storage.data.size                # Sonar data volume size. Default is "1Gi";
    - sonar.storage.database.class           # Storageclass for Sonar database volume. Default is "gp2";
    - sonar.storage.database.size            # Sonar database volume size. Default is "1Gi".
    ```
   
4. Install operator in the <edp_cicd_project> namespace with the helm command; find below the installation command example:
    ```bash
    helm install sonar-operator epamedp/sonar-operator --version <chart_version> --namespace <edp_cicd_project> --set name=sonar-operator --set global.edpName=<edp_cicd_project> --set global.platform=<platform_type> --set global.dnsWildCard=<cluster_DNS_wildcard>
    ```

5. Check the <edp_cicd_project> namespace that should contain operator deployment with your operator in a running status.

## Local Development
In order to develop the operator, first set up a local environment. For details, please refer to the [Local Development](documentation/local-development.md) page.

### Related Articles

- [Architecture Scheme of Sonar Operator](documentation/arch.md)