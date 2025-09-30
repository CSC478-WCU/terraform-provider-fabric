package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &SitesDataSource{}

func NewSitesDataSource() datasource.DataSource {
	return &SitesDataSource{}
}

// SitesDataSource defines the data source implementation.
type SitesDataSource struct {
	providerData *FabricProviderData
}

// SitesDataSourceModel describes the data source data model.
type SitesDataSourceModel struct {
	ID    types.String   `tfsdk:"id"`
	Sites []SiteModel    `tfsdk:"sites"`
}

// SiteModel describes a FABRIC site
type SiteModel struct {
	Name        types.String `tfsdk:"name"`
	Code        types.String `tfsdk:"code"`
	Description types.String `tfsdk:"description"`
	Location    types.String `tfsdk:"location"`
}

func (d *SitesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sites"
}

func (d *SitesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves available FABRIC sites.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Data source identifier",
				Computed:            true,
			},
			"sites": schema.ListNestedAttribute{
				MarkdownDescription: "List of available FABRIC sites",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Site name",
							Computed:            true,
						},
						"code": schema.StringAttribute{
							MarkdownDescription: "Site code (e.g., CLEM, NCSA)",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Site description",
							Computed:            true,
						},
						"location": schema.StringAttribute{
							MarkdownDescription: "Site location",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *SitesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SitesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// For now, provide a static list of known FABRIC sites
	// In a real implementation, you would fetch this from the API
	knownSites := []SiteModel{
		{
			Name:        types.StringValue("Clemson University"),
			Code:        types.StringValue("CLEM"),
			Description: types.StringValue("Clemson University FABRIC site"),
			Location:    types.StringValue("Clemson, SC"),
		},
		{
			Name:        types.StringValue("National Center for Supercomputing Applications"),
			Code:        types.StringValue("NCSA"),
			Description: types.StringValue("NCSA FABRIC site"),
			Location:    types.StringValue("Urbana, IL"),
		},
		{
			Name:        types.StringValue("University of Kentucky"),
			Code:        types.StringValue("UKY"),
			Description: types.StringValue("University of Kentucky FABRIC site"),
			Location:    types.StringValue("Lexington, KY"),
		},
		{
			Name:        types.StringValue("University of Utah"),
			Code:        types.StringValue("UTAH"),
			Description: types.StringValue("University of Utah FABRIC site"),
			Location:    types.StringValue("Salt Lake City, UT"),
		},
		{
			Name:        types.StringValue("RENCI"),
			Code:        types.StringValue("RENC"),
			Description: types.StringValue("RENCI FABRIC site"),
			Location:    types.StringValue("Chapel Hill, NC"),
		},
	}

	data.ID = types.StringValue("fabric-sites")
	data.Sites = knownSites

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}