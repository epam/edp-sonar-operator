module github.com/epmd-edp/sonar-operator/v2

go 1.12

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0

replace github.com/openshift/api => github.com/openshift/api v0.0.0-20180801171038-322a19404e37

require (
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/epmd-edp/edp-component-operator v0.0.1-2
	github.com/epmd-edp/jenkins-operator/v2 v2.3.0-130.0.20200416062406-16c330e09a19
	github.com/epmd-edp/keycloak-operator v1.0.31-alpha-56
	github.com/openshift/api v3.9.0+incompatible
	github.com/openshift/client-go v3.9.0+incompatible
	github.com/operator-framework/operator-sdk v0.0.0-20190530173525-d6f9cdf2f52e
	github.com/pkg/errors v0.8.1
	github.com/spf13/pflag v1.0.3
	github.com/totherme/unstructured v0.0.0-20170821094912-3faf2d56d8b8
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	gopkg.in/resty.v1 v1.12.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/client-go v0.0.0-20190228174230-b40b2a5939e4
	sigs.k8s.io/controller-runtime v0.1.12
)
