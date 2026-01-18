package docker

import (
	"context"
	"fmt"

	"github.com/yourorg/grund/internal/application/ports"
	"github.com/yourorg/grund/internal/domain/infrastructure"
)

// CompositeInfrastructureProvisioner combines multiple provisioners
// This follows the Composite pattern
type CompositeInfrastructureProvisioner struct {
	provisioners []ports.InfrastructureProvisioner
}

// NewCompositeInfrastructureProvisioner creates a composite provisioner
func NewCompositeInfrastructureProvisioner(provisioners ...ports.InfrastructureProvisioner) ports.InfrastructureProvisioner {
	return &CompositeInfrastructureProvisioner{
		provisioners: provisioners,
	}
}

// ProvisionPostgres delegates to the first provisioner that supports it
func (c *CompositeInfrastructureProvisioner) ProvisionPostgres(ctx context.Context, config *infrastructure.PostgresConfig) error {
	for _, p := range c.provisioners {
		if err := p.ProvisionPostgres(ctx, config); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no provisioner supports postgres")
}

// ProvisionMongoDB delegates to the first provisioner that supports it
func (c *CompositeInfrastructureProvisioner) ProvisionMongoDB(ctx context.Context, config *infrastructure.MongoDBConfig) error {
	for _, p := range c.provisioners {
		if err := p.ProvisionMongoDB(ctx, config); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no provisioner supports mongodb")
}

// ProvisionRedis delegates to the first provisioner that supports it
func (c *CompositeInfrastructureProvisioner) ProvisionRedis(ctx context.Context, config *infrastructure.RedisConfig) error {
	for _, p := range c.provisioners {
		if err := p.ProvisionRedis(ctx, config); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no provisioner supports redis")
}

// ProvisionLocalStack delegates to LocalStack provisioner
func (c *CompositeInfrastructureProvisioner) ProvisionLocalStack(ctx context.Context, req infrastructure.InfrastructureRequirements) error {
	for _, p := range c.provisioners {
		if err := p.ProvisionLocalStack(ctx, req); err == nil {
			return nil
		}
	}
	return fmt.Errorf("no provisioner supports localstack")
}
