package topology

import "encoding/xml"

type GraphML struct {
	XMLName xml.Name `xml:"graphml"`
	Xmlns   string   `xml:"xmlns,attr"`
	Keys    []Key    `xml:"key"`
	Graph   Graph    `xml:"graph"`
}

type Key struct {
	ID       string `xml:"id,attr"`
	For      string `xml:"for,attr"`
	AttrName string `xml:"attr.name,attr"`
	AttrType string `xml:"attr.type,attr"`
}

type Graph struct {
	Edgedefault string `xml:"edgedefault,attr"`
	Nodes       []Node `xml:"node"`
	Edges       []Edge `xml:"edge"`
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
