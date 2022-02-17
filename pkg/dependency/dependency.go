package dependency

// This package is heavily inspired from
// https://github.com/dnaeon/go-dependency-graph-algorithm
// E.g. let's say these dependencies are defined
// C -> A
// D -> A
// A -> B
// B -> X
// then this package allows to resolve this dependency to this chain(order).
// X
// B
// A
// C
// D
// Read more about it and the motivation behind it at
// http://dnaeon.github.io/dependency-graph-resolution-algorithm-in-go/

import (
	"bytes"
	"errors"
	"fmt"

	mapset "github.com/deckarep/golang-set"
)

// Node represents a single node in the graph with it's dependencies
type Node struct {
	// Name of the node
	Name string

	// Dependencies of the node
	Deps []string
}

// NewNode creates a new node
func NewNode(name string, deps ...string) *Node {
	n := &Node{
		Name: name,
		Deps: deps,
	}

	return n
}

type Graph []*Node

// Displays the dependency graph
func Display(graph Graph) {
	fmt.Printf("%s", graph.String())
}

func (g Graph) String() string {
	var buf bytes.Buffer
	for _, node := range g {
		for _, dep := range node.Deps {
			fmt.Fprintf(&buf, "* %s -> %s\n", node.Name, dep)
		}
	}
	return buf.String()
}

func Has(graph Graph) bool {
	for _, node := range graph {
		if len(node.Deps) > 0 {
			return true
		}
	}
	return false
}

// Resolves the dependency graph
func Resolve(graph Graph) (Graph, error) {
	// A map containing the node names and the actual node object
	nodeNames := make(map[string]*Node)

	// A map containing the nodes and their dependencies
	nodeDependencies := make(map[string]mapset.Set)

	// Populate the maps
	for _, node := range graph {
		nodeNames[node.Name] = node

		dependencySet := mapset.NewSet()
		for _, dep := range node.Deps {
			dependencySet.Add(dep)
		}
		nodeDependencies[node.Name] = dependencySet
	}

	// Iteratively find and remove nodes from the graph which have no dependencies.
	// If at some point there are still nodes in the graph and we cannot find
	// nodes without dependencies, that means we have a circular dependency
	var resolved Graph
	for len(nodeDependencies) != 0 {
		// Get all nodes from the graph which have no dependencies
		readySet := mapset.NewSet()
		for name, deps := range nodeDependencies {
			if deps.Cardinality() == 0 {
				readySet.Add(name)
			}
		}

		// If there aren't any ready nodes, then we have a cicular dependency
		if readySet.Cardinality() == 0 {
			var g Graph
			for name := range nodeDependencies {
				g = append(g, nodeNames[name])
			}

			return g, errors.New("Circular dependency found")
		}

		// Remove the ready nodes and add them to the resolved graph
		for name := range readySet.Iter() {
			delete(nodeDependencies, name.(string))
			resolved = append(resolved, nodeNames[name.(string)])
		}

		// Also make sure to remove the ready nodes from the
		// remaining node dependencies as well
		for name, deps := range nodeDependencies {
			diff := deps.Difference(readySet)
			nodeDependencies[name] = diff
		}
	}

	return resolved, nil
}
