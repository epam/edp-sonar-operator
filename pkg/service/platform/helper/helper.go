package helper

func GenerateLabels(name string) map[string]string {
	return map[string]string{
		"app": name,
	}
}
