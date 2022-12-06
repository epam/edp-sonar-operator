<a name="unreleased"></a>
## [Unreleased]


<a name="v2.13.0"></a>
## [v2.13.0] - 2022-11-21
### Features

- Skip jenkins configuration when it is not deployed [EPMDEDP-10650](https://jiraeu.epam.com/browse/EPMDEDP-10650)

### Bug Fixes

- Align SonarQube plugins installation [EPMDEDP-10821](https://jiraeu.epam.com/browse/EPMDEDP-10821)
- Clean up sonar plugins installation folder [EPMDEDP-10821](https://jiraeu.epam.com/browse/EPMDEDP-10821)

### Code Refactoring

- Disable Sonar telemetry for OpenShift deployment [EPMDEDP-10655](https://jiraeu.epam.com/browse/EPMDEDP-10655)
- Disable sonar telemetry by default [EPMDEDP-10655](https://jiraeu.epam.com/browse/EPMDEDP-10655)

### Routine

- Update current development version [EPMDEDP-10274](https://jiraeu.epam.com/browse/EPMDEDP-10274)
- Enable go-lint for operator [EPMDEDP-10628](https://jiraeu.epam.com/browse/EPMDEDP-10628)
- Upgrade sonar to version LTS 8.9.10 [EPMDEDP-10754](https://jiraeu.epam.com/browse/EPMDEDP-10754)


<a name="v2.12.0"></a>
## [v2.12.0] - 2022-08-26
### Features

- Switch to use V1 apis of EDP components [EPMDEDP-10087](https://jiraeu.epam.com/browse/EPMDEDP-10087)
- Download required tools for Makefile targets [EPMDEDP-10105](https://jiraeu.epam.com/browse/EPMDEDP-10105)
- Switch CRDs to v1 version [EPMDEDP-9221](https://jiraeu.epam.com/browse/EPMDEDP-9221)

### Bug Fixes

- Downgrade Sonar version [EPMDEDP-10281](https://jiraeu.epam.com/browse/EPMDEDP-10281)
- Downgrade Sonar plugins to fix compatibility [EPMDEDP-10281](https://jiraeu.epam.com/browse/EPMDEDP-10281)
- Make sure we pass refreshed JWT token from cookies [EPMDEDP-10395](https://jiraeu.epam.com/browse/EPMDEDP-10395)

### Code Refactoring

- Deprecate unused Spec components for Sonar v1 [EPMDEDP-10128](https://jiraeu.epam.com/browse/EPMDEDP-10128)
- Use repository and tag for image reference in chart [EPMDEDP-10389](https://jiraeu.epam.com/browse/EPMDEDP-10389)
- Apply new lint config [EPMDEDP-8066](https://jiraeu.epam.com/browse/EPMDEDP-8066)

### Routine

- Upgrade go version to 1.18 [EPMDEDP-10110](https://jiraeu.epam.com/browse/EPMDEDP-10110)
- Fix Jira Ticket pattern for changelog generator [EPMDEDP-10159](https://jiraeu.epam.com/browse/EPMDEDP-10159)
- Update alpine base image to 3.16.2 version [EPMDEDP-10274](https://jiraeu.epam.com/browse/EPMDEDP-10274)
- Update SonarQube plugins version [EPMDEDP-10274](https://jiraeu.epam.com/browse/EPMDEDP-10274)
- Update SonarQube to 8.9.9-community version [EPMDEDP-10274](https://jiraeu.epam.com/browse/EPMDEDP-10274)
- Update sonar version to 8.9.9-community [EPMDEDP-10281](https://jiraeu.epam.com/browse/EPMDEDP-10281)
- Update sonar version to 8.9.9-community [EPMDEDP-10281](https://jiraeu.epam.com/browse/EPMDEDP-10281)
- Revert: Update sonar version to 8.9.9-community [EPMDEDP-10281](https://jiraeu.epam.com/browse/EPMDEDP-10281)
- Change 'go get' to 'go install' for git-chglog [EPMDEDP-10337](https://jiraeu.epam.com/browse/EPMDEDP-10337)
- Use deployments as default deploymentType for OpenShift [EPMDEDP-10344](https://jiraeu.epam.com/browse/EPMDEDP-10344)
- Remove VERSION file [EPMDEDP-10387](https://jiraeu.epam.com/browse/EPMDEDP-10387)
- Add gcflags for go build artifact [EPMDEDP-10411](https://jiraeu.epam.com/browse/EPMDEDP-10411)
- Update chart annotation [EPMDEDP-8832](https://jiraeu.epam.com/browse/EPMDEDP-8832)
- Update current development version [EPMDEDP-8832](https://jiraeu.epam.com/browse/EPMDEDP-8832)

### Documentation

- Align README.md [EPMDEDP-10274](https://jiraeu.epam.com/browse/EPMDEDP-10274)


<a name="v2.11.0"></a>
## [v2.11.0] - 2022-05-25
### Features

- implement SonarPermissionTemplate custom resource [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)
- implement SonarGroup custom resource [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Use sonar-developers role for OIDC integration [EPMDEDP-7506](https://jiraeu.epam.com/browse/EPMDEDP-7506)
- implment default permission template for sonar CR [EPMDEDP-7633](https://jiraeu.epam.com/browse/EPMDEDP-7633)
- add .golangci-lint config [EPMDEDP-8066](https://jiraeu.epam.com/browse/EPMDEDP-8066)
- Update Makefile changelog target [EPMDEDP-8218](https://jiraeu.epam.com/browse/EPMDEDP-8218)
- Implement validation steps for api/helm docs [EPMDEDP-8329](https://jiraeu.epam.com/browse/EPMDEDP-8329)
- Add ingress tls certificate option when using ingress controller [EPMDEDP-8377](https://jiraeu.epam.com/browse/EPMDEDP-8377)
- Add CRD API documentation [EPMDEDP-8385](https://jiraeu.epam.com/browse/EPMDEDP-8385)

### Bug Fixes

- permission template id in sync groups [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)
- restry retry count [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)
- Ensure CI user has browse permissions [EPMDEDP-7506](https://jiraeu.epam.com/browse/EPMDEDP-7506)
- sonar secrets generation [EPMDEDP-7633](https://jiraeu.epam.com/browse/EPMDEDP-7633)
- Changelog links [EPMDEDP-7847](https://jiraeu.epam.com/browse/EPMDEDP-7847)
- Upgrade SonarQube to the LTS 8.9.4 version [EPMDEDP-8041](https://jiraeu.epam.com/browse/EPMDEDP-8041)
- Upgrade SonarQube to the LTS 8.9.6 version [EPMDEDP-8041](https://jiraeu.epam.com/browse/EPMDEDP-8041)
- save template ID in k8s, immediately after creation in sonar [EPMDEDP-8055](https://jiraeu.epam.com/browse/EPMDEDP-8055)
- QualityGatesListResponse unmarshalling bugfix [EPMDEDP-8224](https://jiraeu.epam.com/browse/EPMDEDP-8224)
- Change ca-certificates in dockerfile [EPMDEDP-8238](https://jiraeu.epam.com/browse/EPMDEDP-8238)
- Fix sonar.plugins.proxy definition [EPMDEDP-8374](https://jiraeu.epam.com/browse/EPMDEDP-8374)
- Fix changelog generation in GH Release Action [EPMDEDP-8468](https://jiraeu.epam.com/browse/EPMDEDP-8468)

### Code Refactoring

- Decrease initial delay for sonar-db readinessProbe [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Remove initial delay for configuration phase [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Provision developers group in sonar controller [EPMDEDP-7506](https://jiraeu.epam.com/browse/EPMDEDP-7506)
- Address golangci-lint issues [EPMDEDP-7945](https://jiraeu.epam.com/browse/EPMDEDP-7945)
- Refactor User and Token management [EPMDEDP-8006](https://jiraeu.epam.com/browse/EPMDEDP-8006)
- refactor json unmarshalling [EPMDEDP-8224](https://jiraeu.epam.com/browse/EPMDEDP-8224)

### Testing

- Align CI flow for GH Actions [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Add fake kubeconfig to fix unit tests [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Add tests [EPMDEDP-7994](https://jiraeu.epam.com/browse/EPMDEDP-7994)
- Add tests [EPMDEDP-7994](https://jiraeu.epam.com/browse/EPMDEDP-7994)
- Add tests [EPMDEDP-7994](https://jiraeu.epam.com/browse/EPMDEDP-7994)
- Add tests [EPMDEDP-7994](https://jiraeu.epam.com/browse/EPMDEDP-7994)
- tests refactoring [EPMDEDP-7994](https://jiraeu.epam.com/browse/EPMDEDP-7994)
- Add tests [EPMDEDP-7994](https://jiraeu.epam.com/browse/EPMDEDP-7994)

### Routine

- Add mocks to sonar exlcusion list [EPMDEDP-7096](https://jiraeu.epam.com/browse/EPMDEDP-7096)
- Update sonar exclusion list [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)
- Update Ingress resources to the newest API version [EPMDEDP-7476](https://jiraeu.epam.com/browse/EPMDEDP-7476)
- Update release CI pipelines [EPMDEDP-7847](https://jiraeu.epam.com/browse/EPMDEDP-7847)
- Update artifacthub tags [EPMDEDP-8049](https://jiraeu.epam.com/browse/EPMDEDP-8049)
- Populate chart with artifacthub tags [EPMDEDP-8049](https://jiraeu.epam.com/browse/EPMDEDP-8049)
- Update changelog [EPMDEDP-8227](https://jiraeu.epam.com/browse/EPMDEDP-8227)
- Update alpine-wget image [EPMDEDP-8331](https://jiraeu.epam.com/browse/EPMDEDP-8331)
- Upgrade SonarQube to version 8.9.7 [EPMDEDP-8332](https://jiraeu.epam.com/browse/EPMDEDP-8332)
- Update base docker image to alpine 3.15.4 [EPMDEDP-8853](https://jiraeu.epam.com/browse/EPMDEDP-8853)
- Upgrade Sonarqube to the latest LTS 8.9.8 [EPMDEDP-8922](https://jiraeu.epam.com/browse/EPMDEDP-8922)
- Update changelog [EPMDEDP-9185](https://jiraeu.epam.com/browse/EPMDEDP-9185)

### Documentation

- Update Arch schema [EPMDEDP-8385](https://jiraeu.epam.com/browse/EPMDEDP-8385)
- Update documentation section [EPMDEDP-8385](https://jiraeu.epam.com/browse/EPMDEDP-8385)

### BREAKING CHANGE:


Ensure keycloak role is aligned when migrating from


<a name="v2.10.3"></a>
## [v2.10.3] - 2022-02-09
### Bug Fixes

- Change ca-certificates in dockerfile [EPMDEDP-8238](https://jiraeu.epam.com/browse/EPMDEDP-8238)

### Routine

- Populate chart with artifacthub tags [EPMDEDP-8049](https://jiraeu.epam.com/browse/EPMDEDP-8049)


<a name="v2.10.2"></a>
## [v2.10.2] - 2022-01-04
### Bug Fixes

- Upgrade SonarQube to the LTS 8.9.6 version [EPMDEDP-8041](https://jiraeu.epam.com/browse/EPMDEDP-8041)


<a name="v2.10.1"></a>
## [v2.10.1] - 2021-12-16
### Bug Fixes

- Upgrade SonarQube to the LTS 8.9.4 version [EPMDEDP-8041](https://jiraeu.epam.com/browse/EPMDEDP-8041)


<a name="v2.10.0"></a>
## [v2.10.0] - 2021-12-09
### Features

- Provide operator's build information [EPMDEDP-7847](https://jiraeu.epam.com/browse/EPMDEDP-7847)

### Bug Fixes

- Redefine roleRef kind in RB from ClusterRole to Role [EPMDEDP-7279](https://jiraeu.epam.com/browse/EPMDEDP-7279)
- Changelog links [EPMDEDP-7847](https://jiraeu.epam.com/browse/EPMDEDP-7847)

### Code Refactoring

- Expand sonar-operator role [EPMDEDP-7279](https://jiraeu.epam.com/browse/EPMDEDP-7279)
- Add namespace field in roleRef in OKD RB [EPMDEDP-7279](https://jiraeu.epam.com/browse/EPMDEDP-7279)
- Replace cluster-wide role/rolebinding to namespaced [EPMDEDP-7279](https://jiraeu.epam.com/browse/EPMDEDP-7279)

### Formatting

- go fmt [EPMDEDP-7943](https://jiraeu.epam.com/browse/EPMDEDP-7943)

### Routine

- Upgrade SonarQube to the LTS 8.9.3 version [EPMDEDP-7409](https://jiraeu.epam.com/browse/EPMDEDP-7409)
- Upgrade SonarQube Scanner version [EPMDEDP-7409](https://jiraeu.epam.com/browse/EPMDEDP-7409)
- Update openssh-client version [EPMDEDP-7469](https://jiraeu.epam.com/browse/EPMDEDP-7469)
- Add changelog generator [EPMDEDP-7847](https://jiraeu.epam.com/browse/EPMDEDP-7847)
- Add codecov report [EPMDEDP-7885](https://jiraeu.epam.com/browse/EPMDEDP-7885)
- Update docker image [EPMDEDP-7895](https://jiraeu.epam.com/browse/EPMDEDP-7895)
- Use custom go build step for operator [EPMDEDP-7932](https://jiraeu.epam.com/browse/EPMDEDP-7932)
- Update go to version 1.17 [EPMDEDP-7932](https://jiraeu.epam.com/browse/EPMDEDP-7932)

### Documentation

- Update the links on GitHub [EPMDEDP-7781](https://jiraeu.epam.com/browse/EPMDEDP-7781)


<a name="v2.9.0"></a>
## [v2.9.0] - 2021-12-03

<a name="v2.8.0"></a>
## [v2.8.0] - 2021-12-03

<a name="v2.7.2"></a>
## [v2.7.2] - 2021-12-03

<a name="v2.7.1"></a>
## [v2.7.1] - 2021-12-03

<a name="v2.7.0"></a>
## [v2.7.0] - 2021-12-03

[Unreleased]: https://github.com/epam/edp-sonar-operator/compare/v2.13.0...HEAD
[v2.13.0]: https://github.com/epam/edp-sonar-operator/compare/v2.12.0...v2.13.0
[v2.12.0]: https://github.com/epam/edp-sonar-operator/compare/v2.11.0...v2.12.0
[v2.11.0]: https://github.com/epam/edp-sonar-operator/compare/v2.10.3...v2.11.0
[v2.10.3]: https://github.com/epam/edp-sonar-operator/compare/v2.10.2...v2.10.3
[v2.10.2]: https://github.com/epam/edp-sonar-operator/compare/v2.10.1...v2.10.2
[v2.10.1]: https://github.com/epam/edp-sonar-operator/compare/v2.10.0...v2.10.1
[v2.10.0]: https://github.com/epam/edp-sonar-operator/compare/v2.9.0...v2.10.0
[v2.9.0]: https://github.com/epam/edp-sonar-operator/compare/v2.8.0...v2.9.0
[v2.8.0]: https://github.com/epam/edp-sonar-operator/compare/v2.7.2...v2.8.0
[v2.7.2]: https://github.com/epam/edp-sonar-operator/compare/v2.7.1...v2.7.2
[v2.7.1]: https://github.com/epam/edp-sonar-operator/compare/v2.7.0...v2.7.1
[v2.7.0]: https://github.com/epam/edp-sonar-operator/compare/v2.3.0-98...v2.7.0
