package topology

import "fmt"

type NodeConfig struct {
	Name         string
	Site         string
	Type         string
	ImageRef     string
	InstanceType string
	Cores        int64
	RAM          int64
	Disk         int64
}

type LinkConfig struct {
	Name   string
	Source string
	Target string
}

type TopologyConfig struct {
	GraphID string
	Nodes   []NodeConfig
	Links   []LinkConfig
}

func CreateCustomTopology(config TopologyConfig) GraphML {
	keys := []Key{
		{ID: "Site", For: "node", AttrName: "Site", AttrType: "string"},
		{ID: "ImageRef", For: "node", AttrName: "ImageRef", AttrType: "string"},
		{ID: "Type", For: "node", AttrName: "Type", AttrType: "string"},
		{ID: "CapacityHints", For: "node", AttrName: "CapacityHints", AttrType: "string"},
		{ID: "Capacities", For: "node", AttrName: "Capacities", AttrType: "string"},
		{ID: "NodeID", For: "node", AttrName: "NodeID", AttrType: "string"},
		{ID: "GraphID", For: "node", AttrName: "GraphID", AttrType: "string"},
		{ID: "Name", For: "node", AttrName: "Name", AttrType: "string"},
		{ID: "Class", For: "node", AttrName: "Class", AttrType: "string"},
		{ID: "id", For: "node", AttrName: "id", AttrType: "string"},
		{ID: "Class", For: "edge", AttrName: "Class", AttrType: "string"},
		{ID: "Name", For: "edge", AttrName: "Name", AttrType: "string"},
	}

	var nodes []Node
	for i, n := range config.Nodes {
		capacityHints := fmt.Sprintf(`{"instance_type":"%s"}`, n.InstanceType)
		capacities := fmt.Sprintf(`{"core":%d,"ram":%d,"disk":%d}`, n.Cores, n.RAM, n.Disk)
		nodes = append(nodes, Node{
			ID: n.Name,
			Data: []Data{
				{Key: "Site", Value: n.Site},
				{Key: "ImageRef", Value: n.ImageRef},
				{Key: "Type", Value: n.Type},
				{Key: "CapacityHints", Value: capacityHints},
				{Key: "Capacities", Value: capacities},
				{Key: "NodeID", Value: n.Name},
				{Key: "GraphID", Value: config.GraphID},
				{Key: "Name", Value: n.Name},
				{Key: "Class", Value: "NetworkNode"},
				{Key: "id", Value: fmt.Sprintf("%d", i+1)},
			},
		})
	}

	var edges []Edge
	for _, e := range config.Links {
		edges = append(edges, Edge{
			Source: e.Source,
			Target: e.Target,
			Data: []Data{
				{Key: "Class", Value: "Link"},
				{Key: "Name", Value: e.Name},
			},
		})
	}

	return GraphML{
		Xmlns: "http://graphml.graphdrawing.org/xmlns",
		Keys:  keys,
		Graph: Graph{Edgedefault: "directed", Nodes: nodes, Edges: edges},
	}
}
