package ports

import (
	"github.com/vivekkundariya/grund/internal/domain/infrastructure"
	"github.com/vivekkundariya/grund/internal/domain/service"
)

// ComposeGenerator defines the interface for Docker Compose generation
type ComposeGenerator interface {
	Generate(services []*service.Service, infra infrastructure.InfrastructureRequirements) (string, error)
}

// EnvironmentResolver defines the interface for environment variable resolution
type EnvironmentResolver interface {
	Resolve(envRefs map[string]string, context EnvironmentContext) (map[string]string, error)
}

// EnvironmentContext provides context for resolving environment variables
type EnvironmentContext struct {
	// Infrastructure contexts (postgres, redis, mongodb)
	Infrastructure map[string]InfrastructureContext
	
	// Service contexts (other services this service depends on)
	Services map[string]ServiceContext
	
	// Self context (the service being configured)
	Self ServiceContext
	
	// AWS/LocalStack resources
	SQS map[string]QueueContext
	SNS map[string]TopicContext
	S3  map[string]BucketContext
	
	// LocalStack endpoint
	LocalStack LocalStackContext
}

// InfrastructureContext provides infrastructure connection details
type InfrastructureContext struct {
	Host     string
	Port     int
	Database string // For postgres/mongodb
	Password string
	Username string
}

// ServiceContext provides service connection details
type ServiceContext struct {
	Host   string
	Port   int
	Config map[string]interface{}
}

// QueueContext provides SQS queue details
type QueueContext struct {
	Name string
	URL  string
	ARN  string
	DLQ  string // Dead-letter queue URL if configured
}

// TopicContext provides SNS topic details
type TopicContext struct {
	Name string
	ARN  string
}

// BucketContext provides S3 bucket details
type BucketContext struct {
	Name string
	URL  string
}

// LocalStackContext provides LocalStack connection details
type LocalStackContext struct {
	Endpoint        string
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	AccountID       string
}

// NewDefaultEnvironmentContext creates a default environment context
// with standard LocalStack values
func NewDefaultEnvironmentContext() EnvironmentContext {
	return EnvironmentContext{
		Infrastructure: make(map[string]InfrastructureContext),
		Services:       make(map[string]ServiceContext),
		SQS:            make(map[string]QueueContext),
		SNS:            make(map[string]TopicContext),
		S3:             make(map[string]BucketContext),
		LocalStack: LocalStackContext{
			Endpoint:        "http://localstack:4566",
			Region:          "us-east-1",
			AccessKeyID:     "test",
			SecretAccessKey: "test",
			AccountID:       "000000000000",
		},
	}
}
