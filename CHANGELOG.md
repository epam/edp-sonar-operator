<a name="unreleased"></a>
## [Unreleased]

### Features

- implement SonarPermissionTemplate custom resource [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)
- implement SonarGroup custom resource [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)

### Bug Fixes

- restry retry count [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)
- Changelog links [EPMDEDP-7847](https://jiraeu.epam.com/browse/EPMDEDP-7847)

### Code Refactoring

- Address golangci-lint issues [EPMDEDP-7945](https://jiraeu.epam.com/browse/EPMDEDP-7945)
- Refactor User and Token management [EPMDEDP-8006](https://jiraeu.epam.com/browse/EPMDEDP-8006)

### Testing

- Align CI flow for GH Actions [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)
- Add fake kubeconfig to fix unit tests [EPMDEDP-7391](https://jiraeu.epam.com/browse/EPMDEDP-7391)

### Routine

- Update sonar exclusion list [EPMDEDP-7390](https://jiraeu.epam.com/browse/EPMDEDP-7390)


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

[Unreleased]: https://github.com/epam/edp-sonar-operator/compare/v2.10.0...HEAD
[v2.10.0]: https://github.com/epam/edp-sonar-operator/compare/v2.9.0...v2.10.0
[v2.9.0]: https://github.com/epam/edp-sonar-operator/compare/v2.8.0...v2.9.0
[v2.8.0]: https://github.com/epam/edp-sonar-operator/compare/v2.7.2...v2.8.0
[v2.7.2]: https://github.com/epam/edp-sonar-operator/compare/v2.7.1...v2.7.2
[v2.7.1]: https://github.com/epam/edp-sonar-operator/compare/v2.7.0...v2.7.1
[v2.7.0]: https://github.com/epam/edp-sonar-operator/compare/v2.3.0-98...v2.7.0
