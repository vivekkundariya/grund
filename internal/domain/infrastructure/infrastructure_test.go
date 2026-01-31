package infrastructure

import (
	"testing"
)

func TestInfrastructureRequirements_Has_Postgres(t *testing.T) {
	tests := []struct {
		name     string
		infra    InfrastructureRequirements
		expected bool
	}{
		{
			name: "has postgres",
			infra: InfrastructureRequirements{
				Postgres: &PostgresConfig{Database: "testdb"},
			},
			expected: true,
		},
		{
			name:     "no postgres",
			infra:    InfrastructureRequirements{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.infra.Has("postgres")
			if result != tt.expected {
				t.Errorf("Has('postgres') = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestInfrastructureRequirements_Has_MongoDB(t *testing.T) {
	infra := InfrastructureRequirements{
		MongoDB: &MongoDBConfig{Database: "testdb"},
	}

	if !infra.Has("mongodb") {
		t.Error("Has('mongodb') = false, want true")
	}

	empty := InfrastructureRequirements{}
	if empty.Has("mongodb") {
		t.Error("Has('mongodb') = true for empty, want false")
	}
}

func TestInfrastructureRequirements_Has_Redis(t *testing.T) {
	infra := InfrastructureRequirements{
		Redis: &RedisConfig{},
	}

	if !infra.Has("redis") {
		t.Error("Has('redis') = false, want true")
	}

	empty := InfrastructureRequirements{}
	if empty.Has("redis") {
		t.Error("Has('redis') = true for empty, want false")
	}
}

func TestInfrastructureRequirements_Has_LocalStack(t *testing.T) {
	tests := []struct {
		name     string
		infra    InfrastructureRequirements
		expected bool
	}{
		{
			name: "has SQS",
			infra: InfrastructureRequirements{
				SQS: &SQSConfig{Queues: []QueueConfig{{Name: "test"}}},
			},
			expected: true,
		},
		{
			name: "has SNS",
			infra: InfrastructureRequirements{
				SNS: &SNSConfig{Topics: []TopicConfig{{Name: "test"}}},
			},
			expected: true,
		},
		{
			name: "has S3",
			infra: InfrastructureRequirements{
				S3: &S3Config{Buckets: []BucketConfig{{Name: "test"}}},
			},
			expected: true,
		},
		{
			name: "has all AWS services",
			infra: InfrastructureRequirements{
				SQS: &SQSConfig{Queues: []QueueConfig{{Name: "test"}}},
				SNS: &SNSConfig{Topics: []TopicConfig{{Name: "test"}}},
				S3:  &S3Config{Buckets: []BucketConfig{{Name: "test"}}},
			},
			expected: true,
		},
		{
			name:     "no AWS services",
			infra:    InfrastructureRequirements{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.infra.Has("localstack")
			if result != tt.expected {
				t.Errorf("Has('localstack') = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestInfrastructureRequirements_Has_Unknown(t *testing.T) {
	infra := InfrastructureRequirements{
		Postgres: &PostgresConfig{Database: "testdb"},
	}

	if infra.Has("unknown") {
		t.Error("Has('unknown') = true, want false")
	}
}

func TestAggregate_SingleService(t *testing.T) {
	req := InfrastructureRequirements{
		Postgres: &PostgresConfig{Database: "testdb"},
		Redis:    &RedisConfig{},
	}

	result := Aggregate(req)

	if result.Postgres == nil || result.Postgres.Database != "testdb" {
		t.Error("Aggregate() did not preserve Postgres config")
	}
	if result.Redis == nil {
		t.Error("Aggregate() did not preserve Redis config")
	}
}

func TestAggregate_MultipleServices(t *testing.T) {
	req1 := InfrastructureRequirements{
		Postgres: &PostgresConfig{Database: "db1"},
	}
	req2 := InfrastructureRequirements{
		MongoDB: &MongoDBConfig{Database: "db2"},
	}
	req3 := InfrastructureRequirements{
		Redis: &RedisConfig{},
	}

	result := Aggregate(req1, req2, req3)

	if result.Postgres == nil {
		t.Error("Aggregate() missing Postgres")
	}
	if result.MongoDB == nil {
		t.Error("Aggregate() missing MongoDB")
	}
	if result.Redis == nil {
		t.Error("Aggregate() missing Redis")
	}
}

func TestAggregate_SQSQueues(t *testing.T) {
	req1 := InfrastructureRequirements{
		SQS: &SQSConfig{
			Queues: []QueueConfig{
				{Name: "queue-1", DLQ: true},
			},
		},
	}
	req2 := InfrastructureRequirements{
		SQS: &SQSConfig{
			Queues: []QueueConfig{
				{Name: "queue-2", DLQ: false},
				{Name: "queue-3", DLQ: true},
			},
		},
	}

	result := Aggregate(req1, req2)

	if result.SQS == nil {
		t.Fatal("Aggregate() SQS is nil")
	}
	if len(result.SQS.Queues) != 3 {
		t.Errorf("Aggregate() SQS.Queues has %d queues, want 3", len(result.SQS.Queues))
	}

	// Verify all queues are present
	queueNames := make(map[string]bool)
	for _, q := range result.SQS.Queues {
		queueNames[q.Name] = true
	}
	for _, expected := range []string{"queue-1", "queue-2", "queue-3"} {
		if !queueNames[expected] {
			t.Errorf("Aggregate() missing queue %q", expected)
		}
	}
}

func TestAggregate_SNSTopics(t *testing.T) {
	req1 := InfrastructureRequirements{
		SNS: &SNSConfig{
			Topics: []TopicConfig{
				{Name: "topic-1"},
			},
		},
	}
	req2 := InfrastructureRequirements{
		SNS: &SNSConfig{
			Topics: []TopicConfig{
				{Name: "topic-2", Subscriptions: []SubscriptionConfig{{Protocol: "sqs", Endpoint: "${sqs.queue-1.arn}"}}},
			},
		},
	}

	result := Aggregate(req1, req2)

	if result.SNS == nil {
		t.Fatal("Aggregate() SNS is nil")
	}
	if len(result.SNS.Topics) != 2 {
		t.Errorf("Aggregate() SNS.Topics has %d topics, want 2", len(result.SNS.Topics))
	}
}

func TestAggregate_S3Buckets(t *testing.T) {
	req1 := InfrastructureRequirements{
		S3: &S3Config{
			Buckets: []BucketConfig{
				{Name: "bucket-1"},
			},
		},
	}
	req2 := InfrastructureRequirements{
		S3: &S3Config{
			Buckets: []BucketConfig{
				{Name: "bucket-2", Seed: "./fixtures"},
			},
		},
	}

	result := Aggregate(req1, req2)

	if result.S3 == nil {
		t.Fatal("Aggregate() S3 is nil")
	}
	if len(result.S3.Buckets) != 2 {
		t.Errorf("Aggregate() S3.Buckets has %d buckets, want 2", len(result.S3.Buckets))
	}
}

func TestAggregate_FirstPostgresWins(t *testing.T) {
	// When multiple services have postgres, the first one's config is used
	req1 := InfrastructureRequirements{
		Postgres: &PostgresConfig{Database: "db1", Migrations: "./migrations1"},
	}
	req2 := InfrastructureRequirements{
		Postgres: &PostgresConfig{Database: "db2", Migrations: "./migrations2"},
	}

	result := Aggregate(req1, req2)

	if result.Postgres.Database != "db1" {
		t.Errorf("Aggregate() Postgres.Database = %q, want 'db1' (first wins)", result.Postgres.Database)
	}
}

func TestAggregate_Empty(t *testing.T) {
	result := Aggregate()

	if result.Postgres != nil {
		t.Error("Aggregate() with no args should have nil Postgres")
	}
	if result.MongoDB != nil {
		t.Error("Aggregate() with no args should have nil MongoDB")
	}
	if result.Redis != nil {
		t.Error("Aggregate() with no args should have nil Redis")
	}
	if result.SQS != nil {
		t.Error("Aggregate() with no args should have nil SQS")
	}
	if result.SNS != nil {
		t.Error("Aggregate() with no args should have nil SNS")
	}
	if result.S3 != nil {
		t.Error("Aggregate() with no args should have nil S3")
	}
}

func TestInfrastructureType_Constants(t *testing.T) {
	if InfrastructureTypePostgres != "postgres" {
		t.Errorf("InfrastructureTypePostgres = %q, want 'postgres'", InfrastructureTypePostgres)
	}
	if InfrastructureTypeMongoDB != "mongodb" {
		t.Errorf("InfrastructureTypeMongoDB = %q, want 'mongodb'", InfrastructureTypeMongoDB)
	}
	if InfrastructureTypeRedis != "redis" {
		t.Errorf("InfrastructureTypeRedis = %q, want 'redis'", InfrastructureTypeRedis)
	}
	if InfrastructureTypeLocalStack != "localstack" {
		t.Errorf("InfrastructureTypeLocalStack = %q, want 'localstack'", InfrastructureTypeLocalStack)
	}
}

func TestAggregate_SQSQueues_Deduplication(t *testing.T) {
	// Two services requesting the same queue should result in only one queue
	req1 := InfrastructureRequirements{
		SQS: &SQSConfig{
			Queues: []QueueConfig{
				{Name: "shared-queue", DLQ: true},
				{Name: "service-a-queue"},
			},
		},
	}
	req2 := InfrastructureRequirements{
		SQS: &SQSConfig{
			Queues: []QueueConfig{
				{Name: "shared-queue", DLQ: false}, // Same queue, different config
				{Name: "service-b-queue"},
			},
		},
	}

	result := Aggregate(req1, req2)

	if result.SQS == nil {
		t.Fatal("Aggregate() SQS is nil")
	}

	// Should have 3 queues: shared-queue, service-a-queue, service-b-queue
	// shared-queue should NOT be duplicated
	if len(result.SQS.Queues) != 3 {
		t.Errorf("Aggregate() SQS.Queues has %d queues, want 3 (deduplicated)", len(result.SQS.Queues))
	}

	// Count occurrences of shared-queue
	sharedCount := 0
	for _, q := range result.SQS.Queues {
		if q.Name == "shared-queue" {
			sharedCount++
		}
	}
	if sharedCount != 1 {
		t.Errorf("shared-queue appears %d times, want 1 (deduplicated)", sharedCount)
	}
}

func TestAggregate_SNSTopics_Deduplication(t *testing.T) {
	req1 := InfrastructureRequirements{
		SNS: &SNSConfig{
			Topics: []TopicConfig{
				{Name: "shared-events"},
				{Name: "service-a-events"},
			},
		},
	}
	req2 := InfrastructureRequirements{
		SNS: &SNSConfig{
			Topics: []TopicConfig{
				{Name: "shared-events"}, // Duplicate
				{Name: "service-b-events"},
			},
		},
	}

	result := Aggregate(req1, req2)

	if result.SNS == nil {
		t.Fatal("Aggregate() SNS is nil")
	}
	if len(result.SNS.Topics) != 3 {
		t.Errorf("Aggregate() SNS.Topics has %d topics, want 3 (deduplicated)", len(result.SNS.Topics))
	}
}

func TestAggregate_S3Buckets_Deduplication(t *testing.T) {
	req1 := InfrastructureRequirements{
		S3: &S3Config{
			Buckets: []BucketConfig{
				{Name: "shared-bucket"},
				{Name: "service-a-bucket"},
			},
		},
	}
	req2 := InfrastructureRequirements{
		S3: &S3Config{
			Buckets: []BucketConfig{
				{Name: "shared-bucket"}, // Duplicate
				{Name: "service-b-bucket"},
			},
		},
	}

	result := Aggregate(req1, req2)

	if result.S3 == nil {
		t.Fatal("Aggregate() S3 is nil")
	}
	if len(result.S3.Buckets) != 3 {
		t.Errorf("Aggregate() S3.Buckets has %d buckets, want 3 (deduplicated)", len(result.S3.Buckets))
	}
}

func TestInfrastructureRequirementsTunnel(t *testing.T) {
	reqs := InfrastructureRequirements{
		Tunnel: &TunnelRequirement{
			Provider: "cloudflared",
			Targets: []TunnelTargetRequirement{
				{Name: "localstack", Host: "${localstack.host}", Port: "${localstack.port}"},
			},
		},
	}

	if reqs.Tunnel == nil {
		t.Fatal("expected tunnel requirement")
	}
	if reqs.Tunnel.Provider != "cloudflared" {
		t.Errorf("expected cloudflared, got %s", reqs.Tunnel.Provider)
	}
}

func TestAggregateTunnel(t *testing.T) {
	req1 := InfrastructureRequirements{
		Tunnel: &TunnelRequirement{
			Provider: "cloudflared",
			Targets: []TunnelTargetRequirement{
				{Name: "localstack", Host: "localhost", Port: "4566"},
			},
		},
	}
	req2 := InfrastructureRequirements{
		Tunnel: &TunnelRequirement{
			Provider: "cloudflared",
			Targets: []TunnelTargetRequirement{
				{Name: "api", Host: "localhost", Port: "8080"},
			},
		},
	}

	result := Aggregate(req1, req2)

	if result.Tunnel == nil {
		t.Fatal("expected aggregated tunnel")
	}
	if len(result.Tunnel.Targets) != 2 {
		t.Errorf("expected 2 targets, got %d", len(result.Tunnel.Targets))
	}
}
