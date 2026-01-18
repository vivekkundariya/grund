package resolver

import (
	"fmt"
)

// TopologicalSort performs a topological sort on the dependency graph
// Returns services in order they should be started (dependencies first)
func TopologicalSort(graph *DependencyGraph, startServices []string) ([]string, error) {
	// Get all dependencies
	allServices := make(map[string]bool)
	for _, svc := range startServices {
		deps, err := graph.GetAllDependencies(svc)
		if err != nil {
			return nil, err
		}
		for _, dep := range deps {
			allServices[dep] = true
		}
		allServices[svc] = true
	}

	// Convert to list
	servicesList := make([]string, 0, len(allServices))
	for svc := range allServices {
		servicesList = append(servicesList, svc)
	}

	// Calculate in-degrees
	inDegree := make(map[string]int)
	for _, svc := range servicesList {
		inDegree[svc] = 0
	}

	for _, svc := range servicesList {
		node, err := graph.GetNode(svc)
		if err != nil {
			continue
		}
		for _, dep := range node.Dependencies {
			if _, ok := allServices[dep]; ok {
				inDegree[svc]++
			}
		}
	}

	// Kahn's algorithm
	var queue []string
	for _, svc := range servicesList {
		if inDegree[svc] == 0 {
			queue = append(queue, svc)
		}
	}

	var result []string
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		node, err := graph.GetNode(current)
		if err != nil {
			continue
		}

		for _, dependent := range node.Dependents {
			if _, ok := allServices[dependent]; ok {
				inDegree[dependent]--
				if inDegree[dependent] == 0 {
					queue = append(queue, dependent)
				}
			}
		}
	}

	// Check for cycles
	if len(result) != len(servicesList) {
		return nil, fmt.Errorf("circular dependency detected")
	}

	return result, nil
}
