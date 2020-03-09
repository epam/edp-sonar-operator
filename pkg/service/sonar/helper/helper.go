package helper

import (
	"bytes"
	"fmt"
	sonarClientHelper "github.com/epmd-edp/sonar-operator/v2/pkg/client/helper"
	"github.com/epmd-edp/sonar-operator/v2/pkg/helper"
	"github.com/epmd-edp/sonar-operator/v2/pkg/service/sonar/spec"
	"github.com/operator-framework/operator-sdk/pkg/k8sutil"
	"github.com/pkg/errors"
	"text/template"
)

const (
	jenkinsPluginConfigFileName    = "config-sonar-plugin.tmpl"
	defaultTemplatesAbsolutePath   = defaultConfigsAbsolutePath + "/" + defaultTemplatesDirectory
	defaultTemplatesDirectory      = "templates"
	defaultConfigsAbsolutePath     = defaultConfigFilesAbsolutePath + localConfigsRelativePath
	defaultConfigFilesAbsolutePath = "/usr/local/"
	localConfigsRelativePath       = "configs"
)

type JenkinsPluginData struct {
	ServerName string
	ServerPort int
	ServerPath string
	SecretName string
}

func InitNewJenkinsPluginInfo(defaultPort bool) JenkinsPluginData {
	if defaultPort {
		return JenkinsPluginData{ServerPort: spec.Port}
	}
	return JenkinsPluginData{}
}

func ParseDefaultTemplate(data JenkinsPluginData) (bytes.Buffer, error) {
	executableFilePath := helper.GetExecutableFilePath()
	templatesDirectoryPath := defaultTemplatesAbsolutePath
	if _, err := k8sutil.GetOperatorNamespace(); err != nil && err == k8sutil.ErrNoNamespace {
		templatesDirectoryPath = fmt.Sprintf("%v/../%v/%v", executableFilePath, localConfigsRelativePath, defaultTemplatesDirectory)
	}

	var jenkinsScriptContext bytes.Buffer
	templateAbsolutePath := fmt.Sprintf("%v/%v", templatesDirectoryPath, jenkinsPluginConfigFileName)
	if !sonarClientHelper.FileExists(templateAbsolutePath) {
		errMsg := fmt.Sprintf("Template file not found in path specificed! Path: %s", templateAbsolutePath)
		return bytes.Buffer{}, errors.New(errMsg)
	}
	t := template.Must(template.New(jenkinsPluginConfigFileName).ParseFiles(templateAbsolutePath))

	err := t.Execute(&jenkinsScriptContext, data)
	if err != nil {
		return jenkinsScriptContext, errors.Wrapf(err, "Couldn't parse template %v", jenkinsPluginConfigFileName)
	}

	return jenkinsScriptContext, nil
}
