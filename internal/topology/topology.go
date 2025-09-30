package topology

import (
	"bytes"
	"encoding/xml"
	"fmt"

	"github.com/google/uuid"
)

type GraphML struct {
	XMLName xml.Name `xml:"graphml"`
	Xmlns   string   `xml:"xmlns,attr"`

	Keys  []Key `xml:"key"`
	Graph Graph `xml:"graph"`
}

type Key struct {
	ID       string `xml:"id,attr"`
	For      string `xml:"for,attr"`
	AttrName string `xml:"attr.name,attr"`
	AttrType string `xml:"attr.type,attr"`
}

type Graph struct {
	Edgedefault string  `xml:"edgedefault,attr"`
	Nodes       []Node  `xml:"node"`
	Edges       []Edge  `xml:"edge"`
}

type Data struct {
	Key   string `xml:"key,attr"`
	Value string `xml:",chardata"`
}

type Node struct {
	ID   string `xml:"id,attr"`
	Data []Data `xml:"data"`
}

type Edge struct {
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
	Data   []Data `xml:"data"`
}

// NodeConfig represents configuration for a topology node
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

// LinkConfig represents configuration for a topology link
type LinkConfig struct {
	Name   string
	Source string
	Target string
}

// TopologyConfig represents the complete topology configuration
type TopologyConfig struct {
	GraphID string
	Nodes   []NodeConfig
	Links   []LinkConfig
}

// CreateCustomTopology generates a topology from the provided configuration
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
	for i, nodeConfig := range config.Nodes {
		capacityHints := fmt.Sprintf(`{"instance_type":"%s"}`, nodeConfig.InstanceType)
		capacities := fmt.Sprintf(`{"core":%d,"ram":%d,"disk":%d}`, nodeConfig.Cores, nodeConfig.RAM, nodeConfig.Disk)

		node := Node{
			ID: nodeConfig.Name,
			Data: []Data{
				{Key: "Site", Value: nodeConfig.Site},
				{Key: "ImageRef", Value: nodeConfig.ImageRef},
				{Key: "Type", Value: nodeConfig.Type},
				{Key: "CapacityHints", Value: capacityHints},
				{Key: "Capacities", Value: capacities},
				{Key: "NodeID", Value: nodeConfig.Name},
				{Key: "GraphID", Value: config.GraphID},
				{Key: "Name", Value: nodeConfig.Name},
				{Key: "Class", Value: "NetworkNode"},
				{Key: "id", Value: fmt.Sprintf("%d", i+1)},
			},
		}
		nodes = append(nodes, node)
	}

	var edges []Edge
	for _, linkConfig := range config.Links {
		edge := Edge{
			Source: linkConfig.Source,
			Target: linkConfig.Target,
			Data: []Data{
				{Key: "Class", Value: "Link"},
				{Key: "Name", Value: linkConfig.Name},
			},
		}
		edges = append(edges, edge)
	}

	graph := Graph{
		Edgedefault: "directed",
		Nodes:       nodes,
		Edges:       edges,
	}

	return GraphML{
		Xmlns: "http://graphml.graphdrawing.org/xmlns",
		Keys:  keys,
		Graph: graph,
	}
}

// CreateLinearTopology creates a linear chain of nodes
func CreateLinearTopology(nodeCount int, site string, graphID string) GraphML {
	if graphID == "" {
		graphID = NewGraphID()
	}

	var nodes []NodeConfig
	var links []LinkConfig

	// Create nodes
	for i := 0; i < nodeCount; i++ {
		nodeName := fmt.Sprintf("node%d", i+1)
		nodes = append(nodes, NodeConfig{
			Name:         nodeName,
			Site:         site,
			Type:         "VM",
			ImageRef:     "default_rocky_8,qcow2",
			InstanceType: "fabric.c2.m2.d10",
			Cores:        2,
			RAM:          2,
			Disk:         10,
		})

		// Create links between consecutive nodes
		if i > 0 {
			linkName := fmt.Sprintf("link%d", i)
			links = append(links, LinkConfig{
				Name:   linkName,
				Source: fmt.Sprintf("node%d", i),
				Target: fmt.Sprintf("node%d", i+1),
			})
		}
	}

	config := TopologyConfig{
		GraphID: graphID,
		Nodes:   nodes,
		Links:   links,
	}

	return CreateCustomTopology(config)
}

// CreateStarTopology creates a star topology with a central hub
func CreateStarTopology(spokeCount int, site string, graphID string) GraphML {
	if graphID == "" {
		graphID = NewGraphID()
	}

	var nodes []NodeConfig
	var links []LinkConfig

	// Create central hub
	nodes = append(nodes, NodeConfig{
		Name:         "hub",
		Site:         site,
		Type:         "VM",
		ImageRef:     "default_rocky_8,qcow2",
		InstanceType: "fabric.c4.m4.d20", // Larger instance for hub
		Cores:        4,
		RAM:          4,
		Disk:         20,
	})

	// Create spokes
	for i := 0; i < spokeCount; i++ {
		spokeName := fmt.Sprintf("spoke%d", i+1)
		nodes = append(nodes, NodeConfig{
			Name:         spokeName,
			Site:         site,
			Type:         "VM",
			ImageRef:     "default_rocky_8,qcow2",
			InstanceType: "fabric.c2.m2.d10",
			Cores:        2,
			RAM:          2,
			Disk:         10,
		})

		// Create link from hub to spoke
		linkName := fmt.Sprintf("link_hub_spoke%d", i+1)
		links = append(links, LinkConfig{
			Name:   linkName,
			Source: "hub",
			Target: spokeName,
		})
	}

	config := TopologyConfig{
		GraphID: graphID,
		Nodes:   nodes,
		Links:   links,
	}

	return CreateCustomTopology(config)
}

// CreateMultiSiteTopology creates a topology spanning multiple sites
func CreateMultiSiteTopology(sites []string, nodesPerSite int, graphID string) GraphML {
	if graphID == "" {
		graphID = NewGraphID()
	}

	var nodes []NodeConfig
	var links []LinkConfig

	// Create nodes for each site
	for siteIndex, site := range sites {
		for nodeIndex := 0; nodeIndex < nodesPerSite; nodeIndex++ {
			nodeName := fmt.Sprintf("%s_node%d", site, nodeIndex+1)
			nodes = append(nodes, NodeConfig{
				Name:         nodeName,
				Site:         site,
				Type:         "VM",
				ImageRef:     "default_rocky_8,qcow2",
				InstanceType: "fabric.c2.m2.d10",
				Cores:        2,
				RAM:          2,
				Disk:         10,
			})

			// Create intra-site links (linear within site)
			if nodeIndex > 0 {
				linkName := fmt.Sprintf("%s_link%d", site, nodeIndex)
				links = append(links, LinkConfig{
					Name:   linkName,
					Source: fmt.Sprintf("%s_node%d", site, nodeIndex),
					Target: fmt.Sprintf("%s_node%d", site, nodeIndex+1),
				})
			}
		}

		// Create inter-site links (connect first node of each site to first node of next site)
		if siteIndex > 0 && nodesPerSite > 0 {
			linkName := fmt.Sprintf("inter_site_%s_%s", sites[siteIndex-1], site)
			links = append(links, LinkConfig{
				Name:   linkName,
				Source: fmt.Sprintf("%s_node1", sites[siteIndex-1]),
				Target: fmt.Sprintf("%s_node1", site),
			})
		}
	}

	config := TopologyConfig{
		GraphID: graphID,
		Nodes:   nodes,
		Links:   links,
	}

	return CreateCustomTopology(config)
}

func GenerateFabricGraphML(g GraphML) (string, error) {
	buf := &bytes.Buffer{}
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(buf)
	enc.Indent("", "  ")
	if err := enc.Encode(g); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func NewGraphID() string {
	return uuid.New().String()
}
