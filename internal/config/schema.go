package config

import "time"

// ServiceConfig represents the grund.yaml in each service
type ServiceConfig struct {
	Version  string                  `yaml:"version"`
	Service  ServiceInfo             `yaml:"service"`
	Requires Requirements            `yaml:"requires"`
	Env      map[string]string       `yaml:"env"`
	EnvRefs  map[string]string       `yaml:"env_refs"`
	Secrets  map[string]SecretConfig `yaml:"secrets,omitempty"`
}

// SecretConfig defines a secret required by the service
type SecretConfig struct {
	Description string `yaml:"description"`
	Required    *bool  `yaml:"required,omitempty"` // nil defaults to true
}

// IsRequired returns true if the secret is required (default: true)
func (s SecretConfig) IsRequired() bool {
	if s.Required == nil {
		return true // default to required
	}
	return *s.Required
}

type ServiceInfo struct {
	Name   string       `yaml:"name"`
	Type   string       `yaml:"type"` // go, python, node
	Port   int          `yaml:"port"`
	Build  *BuildConfig `yaml:"build,omitempty"`
	Run    *RunConfig   `yaml:"run,omitempty"`
	Health HealthConfig `yaml:"health"`
}

type BuildConfig struct {
	Dockerfile string `yaml:"dockerfile"`
	Context    string `yaml:"context"`
}

type RunConfig struct {
	Command   string `yaml:"command"`
	HotReload bool   `yaml:"hot_reload"`
}

type HealthConfig struct {
	Endpoint string        `yaml:"endpoint"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`
	Retries  int           `yaml:"retries"`
}

type Requirements struct {
	Services       []string             `yaml:"services"`
	Infrastructure InfrastructureConfig `yaml:"infrastructure"`
}

type InfrastructureConfig struct {
	Postgres *PostgresConfig `yaml:"postgres,omitempty"`
	MongoDB  *MongoDBConfig  `yaml:"mongodb,omitempty"`
	Redis    interface{}     `yaml:"redis,omitempty"` // bool or RedisConfig
	SQS      *SQSConfig      `yaml:"sqs,omitempty"`
	SNS      *SNSConfig      `yaml:"sns,omitempty"`
	S3       *S3Config       `yaml:"s3,omitempty"`
}

type PostgresConfig struct {
	Database   string `yaml:"database"`
	Migrations string `yaml:"migrations,omitempty"`
	Seed       string `yaml:"seed,omitempty"`
}

type MongoDBConfig struct {
	Database string `yaml:"database"`
	Seed     string `yaml:"seed,omitempty"`
}

type RedisConfig struct {
	// Future: specific redis config if needed
}

type SQSConfig struct {
	Queues []QueueConfig `yaml:"queues"`
}

type QueueConfig struct {
	Name string `yaml:"name"`
	DLQ  bool   `yaml:"dlq,omitempty"`
}

type SNSConfig struct {
	Topics []TopicConfig `yaml:"topics"`
}

type TopicConfig struct {
	Name          string               `yaml:"name"`
	Subscriptions []SubscriptionConfig `yaml:"subscriptions,omitempty"`
}

type SubscriptionConfig struct {
	Protocol   string            `yaml:"protocol"`
	Endpoint   string            `yaml:"endpoint"`
	Attributes map[string]string `yaml:"attributes,omitempty"`
}

type S3Config struct {
	Buckets []BucketConfig `yaml:"buckets"`
}

type BucketConfig struct {
	Name string `yaml:"name"`
	Seed string `yaml:"seed,omitempty"`
}

// ServiceRegistry represents the services.yaml in the orchestration repo
type ServiceRegistry struct {
	Version      string                  `yaml:"version"`
	Services     map[string]ServiceEntry `yaml:"services"`
	PathDefaults PathDefaults            `yaml:"path_defaults"`
}

type ServiceEntry struct {
	Repo string `yaml:"repo"`
	Path string `yaml:"path"`
}

type PathDefaults struct {
	Base string `yaml:"base"`
}
