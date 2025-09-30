package mappers

import (
	"fmt"

	"github.com/csc478-wcu/terraform-provider-fabric/internal/resources/slice"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/topology"
)

func FromPlanToGraphML(p slice.Plan, graphID string) topology.GraphML {
	nodes := make([]topology.NodeConfig, 0, len(p.Topology.Nodes))
	for i, n := range p.Topology.Nodes {
		nodeType := n.Type
		if nodeType == "" {
			nodeType = "VM"
		}
		image := n.ImageRef
		if image == "" {
			image = "default_rocky_8,qcow2"
		}
		inst := n.InstanceType
		if inst == "" {
			inst = "fabric.c2.m2.d10"
		}

		cores, ram, disk := n.Cores, n.RAM, n.Disk
		if cores == 0 {
			cores = 2
		}
		if ram == 0 {
			ram = 2
		}
		if disk == 0 {
			disk = 10
		}

		name := n.Name
		if name == "" {
			name = fmt.Sprintf("node%d", i+1)
		}

		nodes = append(nodes, topology.NodeConfig{
			Name:         name,
			Site:         n.Site,
			Type:         nodeType,
			ImageRef:     image,
			InstanceType: inst,
			Cores:        cores,
			RAM:          ram,
			Disk:         disk,
		})
	}

	links := make([]topology.LinkConfig, 0, len(p.Topology.Links))
	for _, l := range p.Topology.Links {
		links = append(links, topology.LinkConfig{
			Name:   l.Name,
			Source: l.Source,
			Target: l.Target,
		})
	}

	return topology.CreateCustomTopology(topology.TopologyConfig{
		GraphID: graphID,
		Nodes:   nodes,
		Links:   links,
	})
}
