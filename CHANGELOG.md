<a name="unreleased"></a>
## [Unreleased]

### Features

- implement SonarPermissionTemplate custom resource [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)
- implement SonarGroup custom resource [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Use sonar-developers role for OIDC integration [EPMDEDP-7506](https://jiraeu.epam.com/browse/EPMDEDP-7506)
- implment default permission template for sonar CR [EPMDEDP-7633](https://jiraeu.epam.com/browse/EPMDEDP-7633)

### Bug Fixes

- restry retry count [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)
- permission template id in sync groups [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)
- Ensure CI user has browse permissions [EPMDEDP-7506](https://jiraeu.epam.com/browse/EPMDEDP-7506)
- sonar secrets generation [EPMDEDP-7633](https://jiraeu.epam.com/browse/EPMDEDP-7633)
- Changelog links [EPMDEDP-7847](https://jiraeu.epam.com/browse/EPMDEDP-7847)
- Upgrade SonarQube to the LTS 8.9.4 version [EPMDEDP-8041](https://jiraeu.epam.com/browse/EPMDEDP-8041)
- Upgrade SonarQube to the LTS 8.9.6 version [EPMDEDP-8041](https://jiraeu.epam.com/browse/EPMDEDP-8041)
- save template ID in k8s, immediately after creation in sonar [EPMDEDP-8055](https://jiraeu.epam.com/browse/EPMDEDP-8055)

### Code Refactoring

- Decrease initial delay for sonar-db readinessProbe [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Remove initial delay for configuration phase [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Provision developers group in sonar controller [EPMDEDP-7506](https://jiraeu.epam.com/browse/EPMDEDP-7506)
- Address golangci-lint issues [EPMDEDP-7945](https://jiraeu.epam.com/browse/EPMDEDP-7945)
- Refactor User and Token management [EPMDEDP-8006](https://jiraeu.epam.com/browse/EPMDEDP-8006)

### Testing

- Align CI flow for GH Actions [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Add fake kubeconfig to fix unit tests [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Add tests [EPMDEDP-7994](https://jiraeu.epam.com/browse/EPMDEDP-7994)

### Routine

- Add mocks to sonar exlcusion list [EPMDEDP-7096](https://jiraeu.epam.com/browse/EPMDEDP-7096)
- Update sonar exclusion list [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)
- Update release CI pipelines [EPMDEDP-7847](https://jiraeu.epam.com/browse/EPMDEDP-7847)


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

[Unreleased]: https://github.com/epam/edp-sonar-operator/compare/v2.10.2...HEAD
[v2.10.2]: https://github.com/epam/edp-sonar-operator/compare/v2.10.1...v2.10.2
[v2.10.1]: https://github.com/epam/edp-sonar-operator/compare/v2.10.0...v2.10.1
[v2.10.0]: https://github.com/epam/edp-sonar-operator/compare/v2.9.0...v2.10.0
[v2.9.0]: https://github.com/epam/edp-sonar-operator/compare/v2.8.0...v2.9.0
[v2.8.0]: https://github.com/epam/edp-sonar-operator/compare/v2.7.2...v2.8.0
[v2.7.2]: https://github.com/epam/edp-sonar-operator/compare/v2.7.1...v2.7.2
[v2.7.1]: https://github.com/epam/edp-sonar-operator/compare/v2.7.0...v2.7.1
[v2.7.0]: https://github.com/epam/edp-sonar-operator/compare/v2.3.0-98...v2.7.0
