// Copyright (c) 2025-2026 HP Development Company, L.P.
// SPDX-License-Identifier: MIT

package pipeline

import (
	"fmt"
	"strings"
)

// Node is a single step's position in the resolved execution graph, holding
// upstream (deps) and downstream (dependents) edges plus a back-pointer to
// the step definition.
type Node struct {
	Step       *Step
	Deps       []*Node // steps this one waits on
	Dependents []*Node // steps waiting on this one
}

// DAG is the fully-resolved execution graph for a pipeline.
type DAG struct {
	Nodes  []*Node          // topologically sorted; roots first
	byName map[string]*Node // lookup by step name
}

// BuildDAG derives the execution graph from a pipeline. Edges come from:
//   - inputFrom on a step
//   - explicit dependsOn entries
//   - ${{ steps.<name>.xx }} references anywhere in the step's values
//
// Returns an error if the graph contains a cycle. Callers should validate
// the pipeline before calling BuildDAG (Load and Parse do this automatically).
func BuildDAG(p *Pipeline) (*DAG, error) {
	nodes := make([]*Node, len(p.Steps))
	byName := make(map[string]*Node, len(p.Steps))
	for i, s := range p.Steps {
		n := &Node{Step: s}
		nodes[i] = n
		byName[s.Name] = n
	}

	for _, n := range nodes {
		edges := edgesFor(n.Step)
		for dep := range edges {
			parent := byName[dep]
			n.Deps = append(n.Deps, parent)
			parent.Dependents = append(parent.Dependents, n)
		}
	}

	sorted, err := topoSort(nodes)
	if err != nil {
		return nil, err
	}
	return &DAG{Nodes: sorted, byName: byName}, nil
}

// Node looks up a graph node by step name. Returns nil if not found.
func (d *DAG) Node(name string) *Node { return d.byName[name] }

// edgesFor returns the set of step names that s depends on, deduplicating
// across the three edge sources (inputFrom, dependsOn, ${{ steps.* }}).
func edgesFor(s *Step) map[string]struct{} {
	edges := make(map[string]struct{})
	if s.InputFrom != "" {
		edges[s.InputFrom] = struct{}{}
	}
	for _, dep := range s.DependsOn {
		edges[dep] = struct{}{}
	}
	for _, ref := range collectStepRefs(s) {
		edges[ref] = struct{}{}
	}
	return edges
}

// topoSort returns nodes in dependency-first order (roots at index 0).
// Uses Kahn's algorithm so cycle detection is a straightforward
// "did we visit every node" check at the end.
func topoSort(nodes []*Node) ([]*Node, error) {
	indegree := make(map[*Node]int, len(nodes))
	for _, n := range nodes {
		indegree[n] = len(n.Deps)
	}

	// Seed the queue with all roots in original declaration order to keep
	// output stable for equivalent inputs.
	queue := make([]*Node, 0, len(nodes))
	for _, n := range nodes {
		if indegree[n] == 0 {
			queue = append(queue, n)
		}
	}

	out := make([]*Node, 0, len(nodes))
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		out = append(out, n)
		for _, child := range n.Dependents {
			indegree[child]--
			if indegree[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	if len(out) != len(nodes) {
		return nil, fmt.Errorf("pipeline: dependency cycle involving steps [%s]",
			strings.Join(cycleMembers(nodes, indegree), ", "))
	}
	return out, nil
}

// cycleMembers returns the names of nodes still holding positive indegree
// after Kahn's algorithm i.e., every node participating in some cycle.
func cycleMembers(nodes []*Node, indegree map[*Node]int) []string {
	var names []string
	for _, n := range nodes {
		if indegree[n] > 0 {
			names = append(names, n.Step.Name)
		}
	}
	return names
}
