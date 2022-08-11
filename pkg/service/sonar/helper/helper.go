package helper

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/pkg/errors"

	sonarClientHelper "github.com/epam/edp-sonar-operator/v2/pkg/client/helper"
	"github.com/epam/edp-sonar-operator/v2/pkg/helper"
	"github.com/epam/edp-sonar-operator/v2/pkg/service/sonar/spec"
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
	if !helper.RunningInCluster() {
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
