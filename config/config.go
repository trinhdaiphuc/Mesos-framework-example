package config

import "time"

type Config struct {
	User              string
	Name              string
	Role              string
	SchedulerEndpoint string
	ExecutorBinary    string
	Timeout           time.Duration
	FailoverTimeout   time.Duration
	Checkpoint        bool
	Hostname          string
}

func NewConfig() Config {
	return Config{
		User:              env("FRAMEWORK_USER", "root"),
		Name:              env("FRAMEWORK_NAME", "example"),
		Role:              env("FRAMEWORK_ROLE", "*"),
		SchedulerEndpoint: env("MESOS_ENDPOINT", "http://127.0.0.1:5050/api/v1/scheduler"),
		ExecutorBinary:    env("EXEC_BINARY", "./executor"),
		FailoverTimeout:   envDuration("SCHEDULER_FAILOVER_TIMEOUT", "1000h"),
		Timeout:           envDuration("MESOS_CONNECT_TIMEOUT", "20s"),
		Hostname:          env("HOST", ""),
		Checkpoint:        true,
	}
}

func ProtoString(s string) *string    { return &s }
func ProtoFloat64(f float64) *float64 { return &f }
func ProtoBool(b bool) *bool          { return &b }
