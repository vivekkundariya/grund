package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vivekkundariya/grund/internal/ui"
	"gopkg.in/yaml.v3"
)

var (
	addDatabase   string
	addMigrations string
	addQueue      string
	addTopic      string
	addBucket     string
	addDLQ        bool
)

var addCmd = &cobra.Command{
	Use:   "add <resource>",
	Short: "Add a resource to grund.yaml",
	Long: `Add infrastructure or service dependencies to an existing grund.yaml file.

Resources:
  postgres    Add PostgreSQL database
  mongodb     Add MongoDB database
  redis       Add Redis cache
  sqs         Add SQS queue (requires --queue)
  sns         Add SNS topic (requires --topic)
  s3          Add S3 bucket (requires --bucket)
  service     Add service dependency (requires service name as argument)

Examples:
  grund add postgres --database myapp_db
  grund add redis
  grund add sqs --queue order-events --dlq
  grund add sns --topic notifications
  grund add s3 --bucket uploads
  grund add service user-service`,
	Args: cobra.MinimumNArgs(1),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addDatabase, "database", "", "Database name (for postgres/mongodb)")
	addCmd.Flags().StringVar(&addMigrations, "migrations", "", "Migrations path (for postgres)")
	addCmd.Flags().StringVar(&addQueue, "queue", "", "Queue name (for sqs)")
	addCmd.Flags().StringVar(&addTopic, "topic", "", "Topic name (for sns)")
	addCmd.Flags().StringVar(&addBucket, "bucket", "", "Bucket name (for s3)")
	addCmd.Flags().BoolVar(&addDLQ, "dlq", true, "Create dead-letter queue (for sqs)")

	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	resource := strings.ToLower(args[0])

	// Find grund.yaml
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	configPath := filepath.Join(wd, "grund.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("grund.yaml not found. Run 'grund init' first")
	}

	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read grund.yaml: %w", err)
	}

	var config map[string]any
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse grund.yaml: %w", err)
	}

	// Get or create requires section
	requires, ok := config["requires"].(map[string]any)
	if !ok {
		requires = make(map[string]any)
		config["requires"] = requires
	}

	// Handle service dependency
	if resource == "service" {
		if len(args) < 2 {
			return fmt.Errorf("service name required: grund add service <name>")
		}
		serviceName := args[1]
		return addServiceDependency(config, requires, serviceName, configPath)
	}

	// Handle infrastructure
	infrastructure, ok := requires["infrastructure"].(map[string]any)
	if !ok {
		infrastructure = make(map[string]any)
		requires["infrastructure"] = infrastructure
	}

	switch resource {
	case "postgres":
		return addPostgres(config, infrastructure, configPath)
	case "mongodb":
		return addMongoDB(config, infrastructure, configPath)
	case "redis":
		return addRedis(config, infrastructure, configPath)
	case "sqs":
		return addSQS(config, infrastructure, configPath)
	case "sns":
		return addSNS(config, infrastructure, configPath)
	case "s3":
		return addS3(config, infrastructure, configPath)
	default:
		return fmt.Errorf("unknown resource: %s\nValid resources: postgres, mongodb, redis, sqs, sns, s3, service", resource)
	}
}

func addPostgres(config map[string]any, infra map[string]any, configPath string) error {
	if _, exists := infra["postgres"]; exists {
		return fmt.Errorf("postgres already configured in grund.yaml")
	}

	dbName := addDatabase
	if dbName == "" {
		// Get service name for default
		if svc, ok := config["service"].(map[string]any); ok {
			if name, ok := svc["name"].(string); ok {
				dbName = name + "_db"
			}
		}
		if dbName == "" {
			dbName = "app_db"
		}
	}

	pg := map[string]any{"database": dbName}
	if addMigrations != "" {
		pg["migrations"] = addMigrations
	}
	infra["postgres"] = pg

	// Add env_ref
	addEnvRef(config, "DATABASE_URL", "postgres://postgres:postgres@${postgres.host}:${postgres.port}/${self.postgres.database}")

	if err := writeConfig(config, configPath); err != nil {
		return err
	}

	ui.Successf("Added PostgreSQL (database: %s)", dbName)
	return nil
}

func addMongoDB(config map[string]any, infra map[string]any, configPath string) error {
	if _, exists := infra["mongodb"]; exists {
		return fmt.Errorf("mongodb already configured in grund.yaml")
	}

	dbName := addDatabase
	if dbName == "" {
		if svc, ok := config["service"].(map[string]any); ok {
			if name, ok := svc["name"].(string); ok {
				dbName = name + "_db"
			}
		}
		if dbName == "" {
			dbName = "app_db"
		}
	}

	infra["mongodb"] = map[string]any{"database": dbName}

	addEnvRef(config, "MONGO_URL", "mongodb://${mongodb.host}:${mongodb.port}/${self.mongodb.database}")

	if err := writeConfig(config, configPath); err != nil {
		return err
	}

	ui.Successf("Added MongoDB (database: %s)", dbName)
	return nil
}

func addRedis(config map[string]any, infra map[string]any, configPath string) error {
	if _, exists := infra["redis"]; exists {
		return fmt.Errorf("redis already configured in grund.yaml")
	}

	infra["redis"] = true

	addEnvRef(config, "REDIS_URL", "redis://${redis.host}:${redis.port}")

	if err := writeConfig(config, configPath); err != nil {
		return err
	}

	ui.Successf("Added Redis")
	return nil
}

func addSQS(config map[string]any, infra map[string]any, configPath string) error {
	if addQueue == "" {
		return fmt.Errorf("--queue flag required for sqs")
	}

	sqs, ok := infra["sqs"].(map[string]any)
	if !ok {
		sqs = map[string]any{"queues": []any{}}
		infra["sqs"] = sqs
	}

	queues, ok := sqs["queues"].([]any)
	if !ok {
		queues = []any{}
	}

	// Check if queue already exists
	for _, q := range queues {
		if qMap, ok := q.(map[string]any); ok {
			if qMap["name"] == addQueue {
				return fmt.Errorf("queue %s already configured", addQueue)
			}
		}
	}

	newQueue := map[string]any{"name": addQueue}
	if addDLQ {
		newQueue["dlq"] = true
	}
	queues = append(queues, newQueue)
	sqs["queues"] = queues

	if err := writeConfig(config, configPath); err != nil {
		return err
	}

	ui.Successf("Added SQS queue: %s", addQueue)
	return nil
}

func addSNS(config map[string]any, infra map[string]any, configPath string) error {
	if addTopic == "" {
		return fmt.Errorf("--topic flag required for sns")
	}

	sns, ok := infra["sns"].(map[string]any)
	if !ok {
		sns = map[string]any{"topics": []any{}}
		infra["sns"] = sns
	}

	topics, ok := sns["topics"].([]any)
	if !ok {
		topics = []any{}
	}

	// Check if topic already exists
	for _, t := range topics {
		if tMap, ok := t.(map[string]any); ok {
			if tMap["name"] == addTopic {
				return fmt.Errorf("topic %s already configured", addTopic)
			}
		}
	}

	topics = append(topics, map[string]any{"name": addTopic})
	sns["topics"] = topics

	if err := writeConfig(config, configPath); err != nil {
		return err
	}

	ui.Successf("Added SNS topic: %s", addTopic)
	return nil
}

func addS3(config map[string]any, infra map[string]any, configPath string) error {
	if addBucket == "" {
		return fmt.Errorf("--bucket flag required for s3")
	}

	s3, ok := infra["s3"].(map[string]any)
	if !ok {
		s3 = map[string]any{"buckets": []any{}}
		infra["s3"] = s3
	}

	buckets, ok := s3["buckets"].([]any)
	if !ok {
		buckets = []any{}
	}

	// Check if bucket already exists
	for _, b := range buckets {
		if bMap, ok := b.(map[string]any); ok {
			if bMap["name"] == addBucket {
				return fmt.Errorf("bucket %s already configured", addBucket)
			}
		}
	}

	buckets = append(buckets, map[string]any{"name": addBucket})
	s3["buckets"] = buckets

	if err := writeConfig(config, configPath); err != nil {
		return err
	}

	ui.Successf("Added S3 bucket: %s", addBucket)
	return nil
}

func addServiceDependency(config map[string]any, requires map[string]any, serviceName string, configPath string) error {
	services, ok := requires["services"].([]any)
	if !ok {
		services = []any{}
	}

	// Check if already exists
	for _, s := range services {
		if str, ok := s.(string); ok && str == serviceName {
			return fmt.Errorf("service %s already in dependencies", serviceName)
		}
	}

	services = append(services, serviceName)
	requires["services"] = services

	// Add env_ref for the service
	envKey := strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_")) + "_URL"
	envVal := fmt.Sprintf("http://${%s.host}:${%s.port}", serviceName, serviceName)
	addEnvRef(config, envKey, envVal)

	if err := writeConfig(config, configPath); err != nil {
		return err
	}

	ui.Successf("Added service dependency: %s", serviceName)
	return nil
}

func addEnvRef(config map[string]any, key, value string) {
	envRefs, ok := config["env_refs"].(map[string]any)
	if !ok {
		envRefs = make(map[string]any)
		config["env_refs"] = envRefs
	}
	if _, exists := envRefs[key]; !exists {
		envRefs[key] = value
	}
}

func writeConfig(config map[string]any, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write grund.yaml: %w", err)
	}

	return nil
}
