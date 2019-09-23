package spec

const (
	Port                              = 9000
	Image                             = "sonarqube"
	DbImage                           = "postgres:9.6"
	DBPort                            = 5432
	LivenessProbeDelay                = 180
	ReadinessProbeDelay               = 180
	DbLivenessProbeDelay              = 180
	DbReadinessProbeDelay             = 180
	MemoryRequest                     = "500Mi"
	JenkinsPluginConfigPostfix string = "jenkins-plugin-config"
)
