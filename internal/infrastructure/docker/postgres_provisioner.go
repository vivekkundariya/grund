package docker

import (
	"context"
	"fmt"

	"github.com/yourorg/grund/internal/application/ports"
	"github.com/yourorg/grund/internal/domain/infrastructure"
)

// PostgresProvisioner implements infrastructure provisioning for PostgreSQL
type PostgresProvisioner struct{}

// NewPostgresProvisioner creates a new Postgres provisioner
func NewPostgresProvisioner() ports.InfrastructureProvisioner {
	return &PostgresProvisioner{}
}

// ProvisionPostgres provisions PostgreSQL
func (p *PostgresProvisioner) ProvisionPostgres(ctx context.Context, config *infrastructure.PostgresConfig) error {
	// PostgreSQL is started via docker-compose, so provisioning here
	// mainly involves running migrations and seeding if needed
	// TODO: Implement migration and seeding logic
	return nil
}

// ProvisionMongoDB not applicable
func (p *PostgresProvisioner) ProvisionMongoDB(ctx context.Context, config *infrastructure.MongoDBConfig) error {
	return fmt.Errorf("mongodb provisioning not supported by postgres provisioner")
}

// ProvisionRedis not applicable
func (p *PostgresProvisioner) ProvisionRedis(ctx context.Context, config *infrastructure.RedisConfig) error {
	return fmt.Errorf("redis provisioning not supported by postgres provisioner")
}

// ProvisionLocalStack not applicable
func (p *PostgresProvisioner) ProvisionLocalStack(ctx context.Context, req infrastructure.InfrastructureRequirements) error {
	return fmt.Errorf("localstack provisioning not supported by postgres provisioner")
}
