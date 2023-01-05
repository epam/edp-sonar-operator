package helper

import (
	"bytes"
	"fmt"
	"text/template"

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

	if !helper.FileExists(templateAbsolutePath) {
		return bytes.Buffer{}, fmt.Errorf("template file not found in path specificed path - %s", templateAbsolutePath)
	}

	t := template.Must(template.New(jenkinsPluginConfigFileName).ParseFiles(templateAbsolutePath))

	if err := t.Execute(&jenkinsScriptContext, data); err != nil {
		return jenkinsScriptContext, fmt.Errorf("failed to parse template %s: %w", jenkinsPluginConfigFileName, err)
	}

	return jenkinsScriptContext, nil
}
