package helper

const (
	UrlCutset = "!\"#$%&'()*+,-./@:;<=>[\\]^_`{|}~"
)

func GenerateLabels(name string) map[string]string {
	return map[string]string{
		"app":                          name,
		"app.edp.epam.com/secret-type": "sonar",
	}
}
