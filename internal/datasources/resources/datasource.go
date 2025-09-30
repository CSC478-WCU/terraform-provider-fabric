package resources

import (
	"context"
	"fmt"
	"strconv"

	"github.com/csc478-wcu/terraform-provider-fabric/internal/runtime"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

var _ datasource.DataSource = &DataSource{}

type DataSource struct{ deps *runtime.Deps }

func New() datasource.DataSource { return &DataSource{} }

func (d *DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_resources"
}

func (d *DataSource) Schema(_ context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = Schema()
}

func (d *DataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	deps, ok := req.ProviderData.(*runtime.Deps)
	if !ok {
		resp.Diagnostics.AddError("Internal error", fmt.Sprintf("unexpected provider deps type %T", req.ProviderData))
		return
	}
	d.deps = deps
}
func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var p Plan
	resp.Diagnostics.Append(req.Config.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var level *int32
	if p.Level != "" {
		parsed, err := strconv.ParseInt(p.Level, 10, 32)
		if err != nil {
			resp.Diagnostics.AddError("Invalid level", err.Error())
			return
		}
		v := int32(parsed)
		level = &v
	}

	models, err := d.deps.Resources.List(ctx, level, p.Includes, p.Excludes)
	if err != nil {
		resp.Diagnostics.AddError("Fetch resources failed", err.Error())
		return
	}

	items := make([]Item, 0, len(models))
	for _, m := range models {
		items = append(items, Item{Model: m})
	}

	p.ID = "fabric-resources"
	p.Resources = items
	resp.Diagnostics.Append(resp.State.Set(ctx, &p)...)
}
