package helper

import (
	"os"
	"path/filepath"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const platformType string = "PLATFORM_TYPE"

var log = logf.Log.WithName("helper_sonar")

func GetExecutableFilePath() string {
	executableFilePath, err := os.Executable()
	if err != nil {
		log.Error(err, "Couldn't get executable path")
	}
	return filepath.Dir(executableFilePath)
}

// GenerateLabels returns map with labels for k8s objects
func GenerateLabels(name string) map[string]string {
	return map[string]string{
		"app": name,
	}
}

func GetPlatformTypeEnv() string {
	return os.Getenv(platformType)
}
