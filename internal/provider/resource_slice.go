package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	fabricclient "github.com/csc478-wcu/terraform-provider-fabric/internal/client"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/topology"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &SliceResource{}
	_ resource.ResourceWithImportState = &SliceResource{}
)

func NewSliceResource() resource.Resource {
	return &SliceResource{}
}

// SliceResource defines the resource implementation.
type SliceResource struct {
	providerData *FabricProviderData
}

// SliceResourceModel describes the resource data model.
type SliceResourceModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	LeaseEndTime types.String `tfsdk:"lease_end_time"`
	SSHKeys      types.List   `tfsdk:"ssh_keys"`
	Topology     types.Object `tfsdk:"topology"`
	State        types.String `tfsdk:"state"`
	SliverCount  types.Int64  `tfsdk:"sliver_count"`
}

// TopologyModel describes the topology configuration
type TopologyModel struct {
	Nodes []NodeModel `tfsdk:"nodes"`
	Links []LinkModel `tfsdk:"links"`
}

// NodeModel describes a node in the topology
type NodeModel struct {
	Name         types.String `tfsdk:"name"`
	Site         types.String `tfsdk:"site"`
	Type         types.String `tfsdk:"type"`
	ImageRef     types.String `tfsdk:"image_ref"`
	InstanceType types.String `tfsdk:"instance_type"`
	Cores        types.Int64  `tfsdk:"cores"`
	RAM          types.Int64  `tfsdk:"ram"`
	Disk         types.Int64  `tfsdk:"disk"`
}

// LinkModel describes a link between nodes
type LinkModel struct {
	Name   types.String `tfsdk:"name"`
	Source types.String `tfsdk:"source"`
	Target types.String `tfsdk:"target"`
}

func (r *SliceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_slice"
}

func (r *SliceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a FABRIC slice.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Slice identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the slice",
				Required:            true,
			},
			"lease_end_time": schema.StringAttribute{
				MarkdownDescription: "End time for the slice lease (RFC3339 format). If not specified, defaults to 24 hours from creation.",
				Optional:            true,
				Computed:            true,
			},
			"ssh_keys": schema.ListAttribute{
				MarkdownDescription: "SSH public keys for accessing slice resources",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"topology": schema.SingleNestedAttribute{
				MarkdownDescription: "Network topology configuration",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"nodes": schema.ListNestedAttribute{
						MarkdownDescription: "List of nodes in the topology",
						Required:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Name of the node",
									Required:            true,
								},
								"site": schema.StringAttribute{
									MarkdownDescription: "FABRIC site for the node (e.g., CLEM, NCSA, etc.)",
									Required:            true,
								},
								"type": schema.StringAttribute{
									MarkdownDescription: "Type of node (VM, Container, etc.)",
									Optional:            true,
									Computed:            true,
								},
								"image_ref": schema.StringAttribute{
									MarkdownDescription: "Image reference (e.g., default_rocky_8,qcow2)",
									Optional:            true,
									Computed:            true,
								},
								"instance_type": schema.StringAttribute{
									MarkdownDescription: "Instance type (e.g., fabric.c2.m2.d10)",
									Optional:            true,
									Computed:            true,
								},
								"cores": schema.Int64Attribute{
									MarkdownDescription: "Number of CPU cores",
									Optional:            true,
									Computed:            true,
								},
								"ram": schema.Int64Attribute{
									MarkdownDescription: "RAM in GB",
									Optional:            true,
									Computed:            true,
								},
								"disk": schema.Int64Attribute{
									MarkdownDescription: "Disk size in GB",
									Optional:            true,
									Computed:            true,
								},
							},
						},
					},
					"links": schema.ListNestedAttribute{
						MarkdownDescription: "List of links between nodes",
						Optional:            true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Name of the link",
									Required:            true,
								},
								"source": schema.StringAttribute{
									MarkdownDescription: "Source node name",
									Required:            true,
								},
								"target": schema.StringAttribute{
									MarkdownDescription: "Target node name",
									Required:            true,
								},
							},
						},
					},
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Current state of the slice",
				Computed:            true,
			},
			"sliver_count": schema.Int64Attribute{
				MarkdownDescription: "Number of slivers in the slice",
				Computed:            true,
			},
		},
	}
}

func (r *SliceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*FabricProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *FabricProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.providerData = providerData
}

func (r *SliceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SliceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Generate GraphML from topology configuration
	var topologyData TopologyModel
	resp.Diagnostics.Append(data.Topology.As(ctx, &topologyData, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	graphID := uuid.New().String()
	graphML := r.buildGraphMLFromTopology(topologyData, graphID)

	xmlStr, err := topology.GenerateFabricGraphML(graphML)
	if err != nil {
		resp.Diagnostics.AddError("Failed to generate GraphML", err.Error())
		return
	}

	// Prepare SSH keys
	var sshKeys []string
	if !data.SSHKeys.IsNull() {
		resp.Diagnostics.Append(data.SSHKeys.ElementsAs(ctx, &sshKeys, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Add default SSH key if none provided
	if len(sshKeys) == 0 && r.providerData.DefaultSSHKey != "" {
		sshKeys = append(sshKeys, r.providerData.DefaultSSHKey)
	}

	// Set default lease end time if not provided
	leaseEndTime := data.LeaseEndTime.ValueString()
	if leaseEndTime == "" {
		leaseEndTime = time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	}

	createResult, err := r.providerData.Client.CreateSlice(ctx, fabricclient.SliceCreateRequest{
		Name:         data.Name.ValueString(),
		LeaseEndTime: leaseEndTime,
		GraphModel:   xmlStr,
		SSHKeys:      sshKeys,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create slice", err.Error())
		return
	}

	data.ID = types.StringValue(createResult.ID)
	data.LeaseEndTime = types.StringValue(leaseEndTime)
	data.State = types.StringValue(createResult.State)
	data.SliverCount = types.Int64Value(int64(createResult.SliverCount))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SliceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SliceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sliceDetails, err := r.providerData.Client.GetSlice(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read slice", err.Error())
		return
	}

	data.State = types.StringValue(sliceDetails.State)
	data.Name = types.StringValue(sliceDetails.Name)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SliceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// For now, slices are immutable except for renewal
	resp.Diagnostics.AddError(
		"Update not supported",
		"Slice updates are not currently supported. To modify a slice, destroy and recreate it.",
	)
}

func (r *SliceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SliceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.providerData.Client.DeleteSlice(ctx, data.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Failed to delete slice", err.Error())
		return
	}
}

func (r *SliceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// buildGraphMLFromTopology converts TopologyModel to topology.GraphML
func (r *SliceResource) buildGraphMLFromTopology(topologyData TopologyModel, graphID string) topology.GraphML {
	// Define standard keys
	keys := []topology.Key{
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

	// Convert nodes
	var nodes []topology.Node
	for i, nodeModel := range topologyData.Nodes {
		// Set defaults
		nodeType := "VM"
		if !nodeModel.Type.IsNull() {
			nodeType = nodeModel.Type.ValueString()
		}

		imageRef := "default_rocky_8,qcow2"
		if !nodeModel.ImageRef.IsNull() {
			imageRef = nodeModel.ImageRef.ValueString()
		}

		instanceType := "fabric.c2.m2.d10"
		if !nodeModel.InstanceType.IsNull() {
			instanceType = nodeModel.InstanceType.ValueString()
		}

		cores := int64(2)
		if !nodeModel.Cores.IsNull() {
			cores = nodeModel.Cores.ValueInt64()
		}

		ram := int64(2)
		if !nodeModel.RAM.IsNull() {
			ram = nodeModel.RAM.ValueInt64()
		}

		disk := int64(10)
		if !nodeModel.Disk.IsNull() {
			disk = nodeModel.Disk.ValueInt64()
		}

		capacityHints := fmt.Sprintf(`{"instance_type":"%s"}`, instanceType)
		capacities := fmt.Sprintf(`{"core":%d,"ram":%d,"disk":%d}`, cores, ram, disk)

		node := topology.Node{
			ID: nodeModel.Name.ValueString(),
			Data: []topology.Data{
				{Key: "Site", Value: nodeModel.Site.ValueString()},
				{Key: "ImageRef", Value: imageRef},
				{Key: "Type", Value: nodeType},
				{Key: "CapacityHints", Value: capacityHints},
				{Key: "Capacities", Value: capacities},
				{Key: "NodeID", Value: nodeModel.Name.ValueString()},
				{Key: "GraphID", Value: graphID},
				{Key: "Name", Value: nodeModel.Name.ValueString()},
				{Key: "Class", Value: "NetworkNode"},
				{Key: "id", Value: fmt.Sprintf("%d", i+1)},
			},
		}
		nodes = append(nodes, node)
	}

	// Convert links
	var edges []topology.Edge
	for _, linkModel := range topologyData.Links {
		edge := topology.Edge{
			Source: linkModel.Source.ValueString(),
			Target: linkModel.Target.ValueString(),
			Data: []topology.Data{
				{Key: "Class", Value: "Link"},
				{Key: "Name", Value: linkModel.Name.ValueString()},
			},
		}
		edges = append(edges, edge)
	}

	graph := topology.Graph{
		Edgedefault: "directed",
		Nodes:       nodes,
		Edges:       edges,
	}

	return topology.GraphML{
		Xmlns: "http://graphml.graphdrawing.org/xmlns",
		Keys:  keys,
		Graph: graph,
	}
}
