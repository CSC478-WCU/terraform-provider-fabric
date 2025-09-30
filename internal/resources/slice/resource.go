package slice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/csc478-wcu/terraform-provider-fabric/internal/runtime"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/topology"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	rframework "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ rframework.Resource = &Resource{}
var _ rframework.ResourceWithImportState = &Resource{}

type Resource struct {
	deps *runtime.Deps
}

func New() rframework.Resource { return &Resource{} }

func (r *Resource) Metadata(_ context.Context, req rframework.MetadataRequest, resp *rframework.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slice"
}

func (r *Resource) Schema(_ context.Context, _ rframework.SchemaRequest, resp *rframework.SchemaResponse) {
	resp.Schema = Schema()
}

func (r *Resource) Configure(_ context.Context, req rframework.ConfigureRequest, resp *rframework.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d, ok := req.ProviderData.(*runtime.Deps)
	if !ok {
		resp.Diagnostics.AddError("Internal error", fmt.Sprintf("unexpected provider deps type %T", req.ProviderData))
		return
	}
	r.deps = d
}

func (r *Resource) Create(ctx context.Context, req rframework.CreateRequest, resp *rframework.CreateResponse) {
	// 1) Read plan into TF types (handles unknown/null properly)
	var tf TFPlan
	resp.Diagnostics.Append(req.Plan.Get(ctx, &tf)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p := FromTFPlan(tf)

	// 2) Normalize domain plan (apply provider defaults so everything is concrete)
	pNorm := applyDefaultsToPlan(p)

	// 3) Build GraphML from the normalized plan
	graphID := uuid.New().String()
	graph := planToGraphML(pNorm, graphID)

	xmlStr, err := topology.Marshal(graph)
	if err != nil {
		resp.Diagnostics.AddError("GraphML generation failed", err.Error())
		return
	}

	// 4) Prepare SSH keys for the API request
	sshWasNull := tf.SSHKeys.IsNull() || tf.SSHKeys.IsUnknown()

	keys := pNorm.SSHKeys
	if len(keys) == 0 && r.deps.DefaultSSHKey != "" {
		keys = []string{r.deps.DefaultSSHKey}
	}
	if keys == nil { // API must see an array, not null
		keys = []string{}
	}

	// 5) Lease defaulting (RFC3339; service normalizes to FABRIC format)
	lease := pNorm.LeaseEndTime
	if lease == "" {
		lease = time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	}

	// 6) Create slice
	id, state, slivers, leaseFinal, err := r.deps.Slices.Create(ctx, pNorm.Name, lease, xmlStr, keys)
	if err != nil {
		resp.Diagnostics.AddError("Create slice failed", err.Error())
		return
	}

	// 7) Build TF topology value from the **normalized** plan (all concrete)
	tfTopo := &TFTopology{
		Nodes: make([]TFNode, 0, len(pNorm.Topology.Nodes)),
		Links: make([]TFLink, 0, len(pNorm.Topology.Links)),
	}
	for _, n := range pNorm.Topology.Nodes {
		tfTopo.Nodes = append(tfTopo.Nodes, TFNode{
			Name:         types.StringValue(n.Name),
			Site:         types.StringValue(n.Site),
			Type:         types.StringValue(n.Type),
			ImageRef:     types.StringValue(n.ImageRef),
			InstanceType: types.StringValue(n.InstanceType),
			Cores:        types.Int64Value(n.Cores),
			RAM:          types.Int64Value(n.RAM),
			Disk:         types.Int64Value(n.Disk),
		})
	}
	for _, l := range pNorm.Topology.Links {
		tfTopo.Links = append(tfTopo.Links, TFLink{
			Name:   types.StringValue(l.Name),
			Source: types.StringValue(l.Source),
			Target: types.StringValue(l.Target),
		})
	}

	// 8) Write state with concrete values
	tfState := TFPlan{
		ID:           types.StringValue(id),
		Name:         types.StringValue(pNorm.Name),
		LeaseEndTime: types.StringValue(leaseFinal),
		State:        types.StringValue(state),
		SliverCount:  types.Int64Value(int64(slivers)),
		Topology:     tfTopo, // normalized (no unknowns)
	}

	// Preserve null vs list semantics for ssh_keys
	if sshWasNull {
		tfState.SSHKeys = types.ListNull(types.StringType)
	} else {
		v, diag := types.ListValueFrom(ctx, types.StringType, keys)
		resp.Diagnostics.Append(diag...)
		if resp.Diagnostics.HasError() {
			return
		}
		tfState.SSHKeys = v
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &tfState)...)
}

// applyDefaultsToPlan returns a copy of p with all node fields concretized
func applyDefaultsToPlan(p Plan) Plan {
	const (
		defType               = "VM"
		defImage              = "default_rocky_8,qcow2"
		defInstanceType       = "fabric.c2.m2.d10"
		defCores        int64 = 2
		defRAM          int64 = 2
		defDisk         int64 = 10
	)
	out := p
	// Normalize nodes
	out.Topology.Nodes = make([]NodePlan, 0, len(p.Topology.Nodes))
	for i, n := range p.Topology.Nodes {
		nn := n
		if nn.Name == "" {
			nn.Name = fmt.Sprintf("node%d", i+1)
		}
		if nn.Type == "" {
			nn.Type = defType
		}
		if nn.ImageRef == "" {
			nn.ImageRef = defImage
		}
		if nn.InstanceType == "" {
			nn.InstanceType = defInstanceType
		}
		if nn.Cores == 0 {
			nn.Cores = defCores
		}
		if nn.RAM == 0 {
			nn.RAM = defRAM
		}
		if nn.Disk == 0 {
			nn.Disk = defDisk
		}
		out.Topology.Nodes = append(out.Topology.Nodes, nn)
	}
	// Links don't have defaults besides names already set in plan, just copy
	out.Topology.Links = append([]LinkPlan(nil), p.Topology.Links...)
	return out
}

func (r *Resource) Read(ctx context.Context, req rframework.ReadRequest, resp *rframework.ReadResponse) {
	// Read prior state using TF types (handles null/unknown safely)
	var tf TFPlan
	resp.Diagnostics.Append(req.State.Get(ctx, &tf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := toString(tf.ID)
	if id == "" {
		// nothing to refresh
		return
	}

	name, state, err := r.deps.Slices.Get(ctx, id)
	if err != nil {
		// if remote is gone, remove from state
		if strings.Contains(err.Error(), "not found") || strings.Contains(strings.ToLower(err.Error()), "no slices") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read slice failed", err.Error())
		return
	}

	// Update only the fields we learned; keep the rest (including any nulls) as-is
	tf.Name = types.StringValue(name)
	tf.State = types.StringValue(state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &tf)...)
}

func (r *Resource) Update(_ context.Context, _ rframework.UpdateRequest, resp *rframework.UpdateResponse) {
	resp.Diagnostics.AddError("Update not supported", "Modify the slice by destroying and re-creating it.")
}

func (r *Resource) Delete(ctx context.Context, req rframework.DeleteRequest, resp *rframework.DeleteResponse) {
	// Read prior state using TF types to avoid null → primitive conversion issues
	var tf TFPlan
	resp.Diagnostics.Append(req.State.Get(ctx, &tf)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := toString(tf.ID)
	if id == "" {
		return // nothing to delete
	}
	if err := r.deps.Slices.Delete(ctx, id); err != nil {
		resp.Diagnostics.AddError("Delete slice failed", err.Error())
	}
}

func (r *Resource) ImportState(ctx context.Context, req rframework.ImportStateRequest, resp *rframework.ImportStateResponse) {
	rframework.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
