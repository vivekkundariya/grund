package resolver

import (
	"fmt"
	"strings"
)

// DetectCircularDependency detects if there's a circular dependency in the graph
func DetectCircularDependency(graph *DependencyGraph, startService string) ([]string, error) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cycle []string

	var dfs func(string, []string) error
	dfs = func(name string, path []string) error {
		visited[name] = true
		recStack[name] = true
		path = append(path, name)

		node, err := graph.GetNode(name)
		if err != nil {
			return err
		}

		for _, dep := range node.Dependencies {
			if !visited[dep] {
				if err := dfs(dep, path); err != nil {
					return err
				}
			} else if recStack[dep] {
				// Found cycle
				cycleStart := -1
				for i, s := range path {
					if s == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					cycle = append(path[cycleStart:], dep)
					return fmt.Errorf("circular dependency detected: %s", strings.Join(cycle, " → "))
				}
			}
		}

		recStack[name] = false
		return nil
	}

	if err := dfs(startService, []string{}); err != nil {
		return cycle, err
	}

	return nil, nil
}
