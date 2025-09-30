package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	fabricclient "github.com/csc478-wcu/terraform-provider-fabric/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ResourcesDataSource{}

func NewResourcesDataSource() datasource.DataSource {
	return &ResourcesDataSource{}
}

// ResourcesDataSource defines the data source implementation.
type ResourcesDataSource struct {
	providerData *FabricProviderData
}

// ResourcesDataSourceModel describes the data source data model.
type ResourcesDataSourceModel struct {
	ID        types.String    `tfsdk:"id"`
	Level     types.String    `tfsdk:"level"`
	Includes  types.List      `tfsdk:"includes"`
	Excludes  types.List      `tfsdk:"excludes"`
	Resources []ResourceModel `tfsdk:"resources"`
}

// ResourceModel describes a single resource
type ResourceModel struct {
	Model types.String `tfsdk:"model"`
}

func (d *ResourcesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resources"
}

func (d *ResourcesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves available FABRIC resources.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"level": schema.StringAttribute{
				MarkdownDescription: "Level of detail for resource information",
				Optional:            true,
			},
			"includes": schema.ListAttribute{
				MarkdownDescription: "Resource types to include",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"excludes": schema.ListAttribute{
				MarkdownDescription: "Resource types to exclude",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"resources": schema.ListNestedAttribute{
				MarkdownDescription: "List of available resources",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"model": schema.StringAttribute{
							MarkdownDescription: "Resource model data (GraphML)",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *ResourcesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*FabricProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *FabricProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.providerData = providerData
}

func (d *ResourcesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ResourcesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	options := fabricclient.ResourceListOptions{}

	if !data.Level.IsNull() {
		levelStr := data.Level.ValueString()
		if levelStr != "" {
			parsed, err := strconv.ParseInt(levelStr, 10, 32)
			if err != nil {
				resp.Diagnostics.AddError("Invalid level value", fmt.Sprintf("expected integer string, got %q", levelStr))
				return
			}
			levelVal := int32(parsed)
			options.Level = &levelVal
		}
	}

	if !data.Includes.IsNull() {
		var includes []string
		resp.Diagnostics.Append(data.Includes.ElementsAs(ctx, &includes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		options.Includes = includes
	}

	if !data.Excludes.IsNull() {
		var excludes []string
		resp.Diagnostics.Append(data.Excludes.ElementsAs(ctx, &excludes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		options.Excludes = excludes
	}

	resources, err := d.providerData.Client.ListResources(ctx, options)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch resources", err.Error())
		return
	}

	// Convert API response to model
	data.ID = types.StringValue("fabric-resources")
	data.Resources = []ResourceModel{}

	for _, resource := range resources {
		resourceModel := ResourceModel{
			Model: types.StringValue(resource.Model),
		}
		data.Resources = append(data.Resources, resourceModel)
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
