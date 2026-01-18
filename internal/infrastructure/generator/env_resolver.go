package generator

import (
	"fmt"
	"strings"

	"github.com/yourorg/grund/internal/application/ports"
)

// EnvironmentResolverImpl implements EnvironmentResolver
type EnvironmentResolverImpl struct{}

// NewEnvironmentResolver creates a new environment resolver
func NewEnvironmentResolver() ports.EnvironmentResolver {
	return &EnvironmentResolverImpl{}
}

// Resolve resolves environment variable references
func (r *EnvironmentResolverImpl) Resolve(envRefs map[string]string, context ports.EnvironmentContext) (map[string]string, error) {
	resolved := make(map[string]string)

	for key, value := range envRefs {
		resolvedValue, err := r.resolveValue(value, context)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", key, err)
		}
		resolved[key] = resolvedValue
	}

	return resolved, nil
}

func (r *EnvironmentResolverImpl) resolveValue(value string, context ports.EnvironmentContext) (string, error) {
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
