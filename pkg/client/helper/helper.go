package helper

import (
	"os"
)

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func CheckPluginInstalled(pluginsList []string, plugin string) bool {
	for _, value := range pluginsList {
		if value == plugin {
			return true
		}
	}
	return false
}
