package generator

import (
	"fmt"
	"strings"

	"github.com/yourorg/grund/internal/config"
)

// ResolveEnvRefs resolves environment variable references
// Example: "${postgres.host}" -> "localhost"
func ResolveEnvRefs(envRefs map[string]string, context *EnvContext) (map[string]string, error) {
	resolved := make(map[string]string)

	for key, value := range envRefs {
		resolvedValue, err := resolveEnvValue(value, context)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", key, err)
		}
		resolved[key] = resolvedValue
	}

	return resolved, nil
}

// EnvContext provides context for resolving environment variable references
type EnvContext struct {
	Infrastructure map[string]InfraContext
	Services      map[string]ServiceContext
	Self          ServiceContext
}

type InfraContext struct {
	Host string
	Port int
}

type ServiceContext struct {
	Host string
	Port int
	// Service-specific config like database name
	Config map[string]interface{}
}

func resolveEnvValue(value string, context *EnvContext) (string, error) {
	// Simple placeholder implementation
	// TODO: Implement full variable resolution with ${...} syntax
	
	result := value
	
	// Replace infrastructure references
	for infraName, infraCtx := range context.Infrastructure {
		result = strings.ReplaceAll(result, fmt.Sprintf("${%s.host}", infraName), infraCtx.Host)
		result = strings.ReplaceAll(result, fmt.Sprintf("${%s.port}", infraName), fmt.Sprintf("%d", infraCtx.Port))
	}
	
	// Replace service references
	for svcName, svcCtx := range context.Services {
		result = strings.ReplaceAll(result, fmt.Sprintf("${%s.host}", svcName), svcCtx.Host)
		result = strings.ReplaceAll(result, fmt.Sprintf("${%s.port}", svcName), fmt.Sprintf("%d", svcCtx.Port))
	}
	
	// Replace self references
	result = strings.ReplaceAll(result, "${self.host}", context.Self.Host)
	result = strings.ReplaceAll(result, "${self.port}", fmt.Sprintf("%d", context.Self.Port))
	
	return result, nil
}
