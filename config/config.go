package config

import "time"

type Config struct {
	User               string
	Name               string
	Role               string
	ExecutorServerPort int
	MesosEndpoint      string
	ExecutorBinary     string
	Timeout            time.Duration
	FailoverTimeout    time.Duration
	Checkpoint         bool
	Hostname           string
}

func NewConfig() Config {
	return Config{
		User:               env("FRAMEWORK_USER", "root"),
		Name:               env("FRAMEWORK_NAME", "example"),
		Role:               env("FRAMEWORK_ROLE", "*"),
		MesosEndpoint:      env("MESOS_ENDPOINT", "127.0.0.1:5050"),
		ExecutorBinary:     env("EXEC_BINARY", "./executor"),
		ExecutorServerPort: envInt("EXEC_SERVER_PORT", "5656"),
		FailoverTimeout:    envDuration("SCHEDULER_FAILOVER_TIMEOUT", "3m"),
		Timeout:            envDuration("MESOS_CONNECT_TIMEOUT", "20s"),
		Hostname:           env("HOST", ""),
		Checkpoint:         true,
	}
}
