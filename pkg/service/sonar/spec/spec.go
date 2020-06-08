package spec

const (
	Port                              = 9000
	DBPort                            = 5432
	LivenessProbeDelay                = 180
	ReadinessProbeDelay               = 60
	DbLivenessProbeDelay              = 180
	DbReadinessProbeDelay             = 60
	MemoryRequest                     = "500Mi"
	JenkinsPluginConfigPostfix string = "jenkins-plugin-config"
)
