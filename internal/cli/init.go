package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// initConfig holds the collected configuration during init
type initConfig struct {
	Name        string
	Type        string
	Port        int
	Dockerfile  string
	HealthPath  string
	Postgres    *postgresInit
	MongoDB     *mongoInit
	Redis       bool
	SQSQueues   []string
	SNSTopics   []string
	S3Buckets   []string
	Services    []string
}

type postgresInit struct {
	Database   string
	Migrations string
}

type mongoInit struct {
	Database string
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize grund in current service",
	Long: `Create a grund.yaml file in the current directory with an interactive setup.

This command walks you through configuring:
  - Service name, type, and port
  - Build configuration (Dockerfile)
  - Health check endpoint
  - Infrastructure requirements (Postgres, MongoDB, Redis, SQS, SNS, S3)
  - Service dependencies`,
	RunE: runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	configPath := filepath.Join(wd, "grund.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("grund.yaml already exists in %s", wd)
	}

	fmt.Println("Initializing Grund configuration...")
	fmt.Println("Press Enter to accept defaults shown in [brackets]")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	cfg := &initConfig{}

	// Service basics
	defaultName := filepath.Base(wd)
	cfg.Name = prompt(reader, "Service name", defaultName)
	cfg.Type = promptChoice(reader, "Service type", []string{"go", "python", "node"}, "go")
	cfg.Port = promptInt(reader, "Port", 8080)

	// Build config
	cfg.Dockerfile = prompt(reader, "Dockerfile path", "Dockerfile")
	cfg.HealthPath = prompt(reader, "Health check endpoint", "/health")

	// Infrastructure
	fmt.Println("\n--- Infrastructure Requirements ---")

	if promptYesNo(reader, "Need PostgreSQL?", false) {
		dbName := prompt(reader, "  Database name", cfg.Name+"_db")
		migrations := prompt(reader, "  Migrations path (leave empty if none)", "")
		cfg.Postgres = &postgresInit{Database: dbName, Migrations: migrations}
	}

	if promptYesNo(reader, "Need MongoDB?", false) {
		dbName := prompt(reader, "  Database name", cfg.Name+"_db")
		cfg.MongoDB = &mongoInit{Database: dbName}
	}

	cfg.Redis = promptYesNo(reader, "Need Redis?", false)

	if promptYesNo(reader, "Need SQS queues?", false) {
		queues := prompt(reader, "  Queue names (comma-separated)", "")
		if queues != "" {
			cfg.SQSQueues = splitAndTrim(queues)
		}
	}

	if promptYesNo(reader, "Need SNS topics?", false) {
		topics := prompt(reader, "  Topic names (comma-separated)", "")
		if topics != "" {
			cfg.SNSTopics = splitAndTrim(topics)
		}
	}

	if promptYesNo(reader, "Need S3 buckets?", false) {
		buckets := prompt(reader, "  Bucket names (comma-separated)", "")
		if buckets != "" {
			cfg.S3Buckets = splitAndTrim(buckets)
		}
	}

	// Service dependencies
	fmt.Println("\n--- Service Dependencies ---")
	deps := prompt(reader, "Dependent services (comma-separated, or empty)", "")
	if deps != "" {
		cfg.Services = splitAndTrim(deps)
	}

	// Generate YAML
	yamlContent := generateGrundYAML(cfg)

	// Preview
	fmt.Println("\n--- Generated grund.yaml ---")
	fmt.Println(yamlContent)
	fmt.Println("---")

	if !promptYesNo(reader, "Write this configuration?", true) {
		fmt.Println("Aborted.")
		return nil
	}

	// Write file
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		return fmt.Errorf("failed to write grund.yaml: %w", err)
	}

	fmt.Printf("\nCreated: %s\n", configPath)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Add this service to your services.yaml registry")
	fmt.Println("  2. Run: grund up " + cfg.Name)

	return nil
}

// prompt asks for input with a default value
func prompt(reader *bufio.Reader, question, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", question, defaultVal)
	} else {
		fmt.Printf("%s: ", question)
	}

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}
	return input
}

// promptInt asks for an integer with a default value
func promptInt(reader *bufio.Reader, question string, defaultVal int) int {
	fmt.Printf("%s [%d]: ", question, defaultVal)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}

	val, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("  Invalid number, using default: %d\n", defaultVal)
		return defaultVal
	}
	return val
}

// promptYesNo asks a yes/no question
func promptYesNo(reader *bufio.Reader, question string, defaultVal bool) bool {
	defaultStr := "y/N"
	if defaultVal {
		defaultStr = "Y/n"
	}

	fmt.Printf("%s [%s]: ", question, defaultStr)

	input, _ := reader.ReadString('\n')
	input = strings.ToLower(strings.TrimSpace(input))

	if input == "" {
		return defaultVal
	}

	return input == "y" || input == "yes"
}

// promptChoice asks user to pick from choices
func promptChoice(reader *bufio.Reader, question string, choices []string, defaultVal string) string {
	fmt.Printf("%s (%s) [%s]: ", question, strings.Join(choices, "/"), defaultVal)

	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		return defaultVal
	}

	// Validate choice
	if slices.Contains(choices, input) {
		return input
	}

	fmt.Printf("  Invalid choice, using default: %s\n", defaultVal)
	return defaultVal
}

// splitAndTrim splits a comma-separated string and trims whitespace
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// generateGrundYAML creates the YAML content from config
func generateGrundYAML(cfg *initConfig) string {
	// Build the config structure for YAML marshaling
	config := map[string]any{
		"version": "1",
		"service": map[string]any{
			"name": cfg.Name,
			"type": cfg.Type,
			"port": cfg.Port,
			"build": map[string]any{
				"dockerfile": cfg.Dockerfile,
				"context":    ".",
			},
			"health": map[string]any{
				"endpoint": cfg.HealthPath,
				"interval": "5s",
				"timeout":  "3s",
				"retries":  10,
			},
		},
	}

	// Build requires section
	requires := map[string]any{
		"services": cfg.Services,
	}

	infrastructure := map[string]any{}

	if cfg.Postgres != nil {
		pg := map[string]any{
			"database": cfg.Postgres.Database,
		}
		if cfg.Postgres.Migrations != "" {
			pg["migrations"] = cfg.Postgres.Migrations
		}
		infrastructure["postgres"] = pg
	}

	if cfg.MongoDB != nil {
		infrastructure["mongodb"] = map[string]any{
			"database": cfg.MongoDB.Database,
		}
	}

	if cfg.Redis {
		infrastructure["redis"] = true
	}

	if len(cfg.SQSQueues) > 0 {
		queues := make([]map[string]any, len(cfg.SQSQueues))
		for i, q := range cfg.SQSQueues {
			queues[i] = map[string]any{
				"name": q,
				"dlq":  true,
			}
		}
		infrastructure["sqs"] = map[string]any{
			"queues": queues,
		}
	}

	if len(cfg.SNSTopics) > 0 {
		topics := make([]map[string]any, len(cfg.SNSTopics))
		for i, t := range cfg.SNSTopics {
			topics[i] = map[string]any{
				"name": t,
			}
		}
		infrastructure["sns"] = map[string]any{
			"topics": topics,
		}
	}

	if len(cfg.S3Buckets) > 0 {
		buckets := make([]map[string]any, len(cfg.S3Buckets))
		for i, b := range cfg.S3Buckets {
			buckets[i] = map[string]any{
				"name": b,
			}
		}
		infrastructure["s3"] = map[string]any{
			"buckets": buckets,
		}
	}

	if len(infrastructure) > 0 {
		requires["infrastructure"] = infrastructure
	}

	config["requires"] = requires

	// Environment variables
	env := map[string]string{
		"APP_ENV":   "development",
		"LOG_LEVEL": "debug",
	}
	config["env"] = env

	// Environment references
	envRefs := map[string]string{}

	if cfg.Postgres != nil {
		envRefs["DATABASE_URL"] = "postgres://postgres:postgres@${postgres.host}:${postgres.port}/${self.postgres.database}"
	}

	if cfg.MongoDB != nil {
		envRefs["MONGO_URL"] = "mongodb://${mongodb.host}:${mongodb.port}/${self.mongodb.database}"
	}

	if cfg.Redis {
		envRefs["REDIS_URL"] = "redis://${redis.host}:${redis.port}"
	}

	// Add service dependency URLs
	for _, svc := range cfg.Services {
		envKey := strings.ToUpper(strings.ReplaceAll(svc, "-", "_")) + "_URL"
		envRefs[envKey] = fmt.Sprintf("http://${%s.host}:${%s.port}", svc, svc)
	}

	if len(envRefs) > 0 {
		config["env_refs"] = envRefs
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Sprintf("# Error generating YAML: %v", err)
	}

	return string(data)
}
