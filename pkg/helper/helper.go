package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	watchNamespaceEnvVar   = "WATCH_NAMESPACE"
	debugModeEnvVar        = "DEBUG_MODE"
	inClusterNamespacePath = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"
	platformType           = "PLATFORM_TYPE"
)

var log = ctrl.Log.WithName("helper_sonar")

func GetExecutableFilePath() string {
	executableFilePath, err := os.Executable()
	if err != nil {
		log.Error(err, "Couldn't get executable path")
	}

	return filepath.Dir(executableFilePath)
}

// GenerateLabels returns map with labels for k8s objects.
func GenerateLabels(name string) map[string]string {
	return map[string]string{
		"app":                          name,
		"app.edp.epam.com/secret-type": "sonar",
	}
}

func GetPlatformTypeEnv() string {
	return os.Getenv(platformType)
}

// GetWatchNamespace returns the namespace the operator should be watching for changes.
func GetWatchNamespace() (string, error) {
	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		return "", fmt.Errorf("%s must be set", watchNamespaceEnvVar)
	}

	return ns, nil
}

// GetDebugMode returns the debug mode value.
func GetDebugMode() (bool, error) {
	mode, found := os.LookupEnv(debugModeEnvVar)
	if !found {
		return false, nil
	}

	b, err := strconv.ParseBool(mode)
	if err != nil {
		return false, err
	}

	return b, nil
}

// RunningInCluster Check whether the operator is running in cluster or locally.
func RunningInCluster() bool {
	_, err := os.Stat(inClusterNamespacePath)
	return !os.IsNotExist(err)
}

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func CheckPluginInstalled(pluginsList []string, plugin string) bool {
	for i := range pluginsList {
		if pluginsList[i] == plugin {
			return true
		}
	}

	return false
}
