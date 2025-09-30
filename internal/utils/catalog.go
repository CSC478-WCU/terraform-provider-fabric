package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

// ------------------ Catalog Structures ------------------

type Catalog struct {
	Sites        []string           `json:"sites"`
	ResourceTypes []string          `json:"resource_types"`
	Flavors      []Flavor           `json:"flavors"`
	Topology     []TopologyRelation `json:"topology"`
}

type Flavor struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type TopologyRelation struct {
	Source   string `json:"source"`
	Target   string `json:"target"`
	Relation string `json:"relation"`
}

// ------------------ Conversion Logic ------------------

// ConvertResourcesToCatalog parses resources.json and creates a static catalog.json
func ConvertResourcesToCatalog(resourcesFile, catalogFile string) error {
	// Step 1: read resources.json
	data, err := os.ReadFile(resourcesFile)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", resourcesFile, err)
	}

	// Step 2: outer JSON structure
	var outer map[string]any
	if err := json.Unmarshal(data, &outer); err != nil {
		return fmt.Errorf("failed to parse outer JSON: %w", err)
	}

	// Step 3: extract model string
	dataArr, ok := outer["data"].([]any)
	if !ok || len(dataArr) == 0 {
		return fmt.Errorf("no data array found in resources.json")
	}

	first, ok := dataArr[0].(map[string]any)
	if !ok {
		return fmt.Errorf("invalid data[0] format")
	}

	modelStr, ok := first["model"].(string)
	if !ok || modelStr == "" {
		return fmt.Errorf("model string missing or empty")
	}

	// Step 4: parse inner model
	var model struct {
		Nodes []map[string]any `json:"nodes"`
		Links []map[string]any `json:"links"`
	}
	if err := json.Unmarshal([]byte(modelStr), &model); err != nil {
		return fmt.Errorf("failed to parse model JSON: %w", err)
	}

	// Step 5: collect catalog info
	siteSet := make(map[string]bool)
	typeSet := make(map[string]bool)
	flavorSet := make(map[string]bool)
	var topology []TopologyRelation

	for _, node := range model.Nodes {
		if name, ok := node["Name"].(string); ok && name != "" {
			siteSet[name] = true
		}
		if t, ok := node["Type"].(string); ok && t != "" {
			typeSet[t] = true
		}
		if hints, ok := node["CapacityHints"].(string); ok && hints != "" {
			var hintMap map[string]any
			if err := json.Unmarshal([]byte(hints), &hintMap); err == nil {
				if inst, ok := hintMap["instance_type"].(string); ok {
					flavorSet[inst] = true
				}
			}
		}
	}

	for _, link := range model.Links {
		topology = append(topology, TopologyRelation{
			Source:   fmt.Sprint(link["source"]),
			Target:   fmt.Sprint(link["target"]),
			Relation: "connects",
		})
	}

	// Step 6: build catalog object
	var catalog Catalog
	for s := range siteSet {
		catalog.Sites = append(catalog.Sites, s)
	}
	for t := range typeSet {
		catalog.ResourceTypes = append(catalog.ResourceTypes, t)
	}
	for f := range flavorSet {
		catalog.Flavors = append(catalog.Flavors, Flavor{Name: f})
	}
	catalog.Topology = topology

	// Step 7: write catalog.json
	out, _ := json.MarshalIndent(catalog, "", "  ")
	if err := os.WriteFile(catalogFile, out, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", catalogFile, err)
	}

	return nil
}
