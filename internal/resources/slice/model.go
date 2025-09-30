package slice

import "github.com/hashicorp/terraform-plugin-framework/types"

// ---------- Framework-facing (plan/state) model ----------

type TFPlan struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	LeaseEndTime types.String `tfsdk:"lease_end_time"`
	SSHKeys      types.List   `tfsdk:"ssh_keys"`
	Topology     *TFTopology  `tfsdk:"topology"`
	State        types.String `tfsdk:"state"`
	SliverCount  types.Int64  `tfsdk:"sliver_count"`
}

type TFTopology struct {
	Nodes []TFNode `tfsdk:"nodes"`
	Links []TFLink `tfsdk:"links"`
}

type TFNode struct {
	Name         types.String `tfsdk:"name"`
	Site         types.String `tfsdk:"site"`
	Type         types.String `tfsdk:"type"`
	ImageRef     types.String `tfsdk:"image_ref"`
	InstanceType types.String `tfsdk:"instance_type"`
	Cores        types.Int64  `tfsdk:"cores"`
	RAM          types.Int64  `tfsdk:"ram"`
	Disk         types.Int64  `tfsdk:"disk"`
}

type TFLink struct {
	Name   types.String `tfsdk:"name"`
	Source types.String `tfsdk:"source"`
	Target types.String `tfsdk:"target"`
}

// ---------- Domain model ----------

type Plan struct {
	ID           string       `tfsdk:"id"`
	Name         string       `tfsdk:"name"`
	LeaseEndTime string       `tfsdk:"lease_end_time"`
	SSHKeys      []string     `tfsdk:"ssh_keys"`
	Topology     TopologyPlan `tfsdk:"topology"`
	State        string       `tfsdk:"state"`
	SliverCount  int64        `tfsdk:"sliver_count"`
}

type TopologyPlan struct {
	Nodes []NodePlan `tfsdk:"nodes"`
	Links []LinkPlan `tfsdk:"links"`
}

type NodePlan struct {
	Name         string `tfsdk:"name"`
	Site         string `tfsdk:"site"`
	Type         string `tfsdk:"type"`
	ImageRef     string `tfsdk:"image_ref"`
	InstanceType string `tfsdk:"instance_type"`
	Cores        int64  `tfsdk:"cores"`
	RAM          int64  `tfsdk:"ram"`
	Disk         int64  `tfsdk:"disk"`
}

type LinkPlan struct {
	Name   string `tfsdk:"name"`
	Source string `tfsdk:"source"`
	Target string `tfsdk:"target"`
}

// ---------- Converters ----------

func toString(v types.String) string {
	if v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

func toInt64(v types.Int64) int64 {
	if v.IsNull() || v.IsUnknown() {
		return 0
	}
	return v.ValueInt64()
}

func toStringSlice(v types.List) []string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	var out []string
	_ = v.ElementsAs(nil, &out, false) // ctx not required for plain conversion here
	return out
}

func FromTFPlan(tf TFPlan) Plan {
	var topo TopologyPlan
	if tf.Topology != nil {
		for _, n := range tf.Topology.Nodes {
			topo.Nodes = append(topo.Nodes, NodePlan{
				Name:         toString(n.Name),
				Site:         toString(n.Site),
				Type:         toString(n.Type),
				ImageRef:     toString(n.ImageRef),
				InstanceType: toString(n.InstanceType),
				Cores:        toInt64(n.Cores),
				RAM:          toInt64(n.RAM),
				Disk:         toInt64(n.Disk),
			})
		}
		for _, l := range tf.Topology.Links {
			topo.Links = append(topo.Links, LinkPlan{
				Name:   toString(l.Name),
				Source: toString(l.Source),
				Target: toString(l.Target),
			})
		}
	}

	return Plan{
		ID:           toString(tf.ID),
		Name:         toString(tf.Name),
		LeaseEndTime: toString(tf.LeaseEndTime),
		SSHKeys:      toStringSlice(tf.SSHKeys),
		Topology:     topo,
		State:        toString(tf.State),
		SliverCount:  toInt64(tf.SliverCount),
	}
}
