package sites

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &DataSource{}

type DataSource struct{}

func New() datasource.DataSource { return &DataSource{} }

func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sites"
}

func (d *DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema()
}

func (d *DataSource) Configure(_ context.Context, _ datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
}

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var s State
	s.ID = "fabric-sites"
	s.Sites = []Site{
		{Name: "Clemson University", Code: "CLEM", Description: "Clemson University FABRIC site", Location: "Clemson, SC"},
		{Name: "National Center for Supercomputing Applications", Code: "NCSA", Description: "NCSA FABRIC site", Location: "Urbana, IL"},
		{Name: "University of Kentucky", Code: "UKY", Description: "University of Kentucky FABRIC site", Location: "Lexington, KY"},
		{Name: "University of Utah", Code: "UTAH", Description: "University of Utah FABRIC site", Location: "Salt Lake City, UT"},
		{Name: "RENCI", Code: "RENC", Description: "RENCI FABRIC site", Location: "Chapel Hill, NC"},
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &s)...)
}
