module github.com/epam/edp-sonar-operator/v2

go 1.14

replace (
	git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
	github.com/openshift/api => github.com/openshift/api v0.0.0-20210416130433-86964261530c
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	k8s.io/api => k8s.io/api v0.20.7-rc.0
)

require (
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/epam/edp-component-operator v0.1.1-0.20210427065236-c7dce7f4ea2b
	github.com/epam/edp-jenkins-operator/v2 v2.3.0-130.0.20210519090957-e6bc8bcf1302
	github.com/epam/edp-keycloak-operator v1.3.0-alpha-81.0.20210427070516-9b6232f72684
	github.com/go-logr/logr v0.4.0
	github.com/openshift/client-go v3.9.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/totherme/unstructured v0.0.0-20170821094912-3faf2d56d8b8
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	gopkg.in/resty.v1 v1.12.0
	k8s.io/api v0.21.0-rc.0
	k8s.io/apimachinery v0.21.0-rc.0
	k8s.io/client-go v0.20.2
	sigs.k8s.io/controller-runtime v0.8.3
)
