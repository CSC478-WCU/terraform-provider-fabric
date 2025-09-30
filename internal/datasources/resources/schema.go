package resources

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Schema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Lists available resources (from FABRIC API).",
		Attributes: map[string]schema.Attribute{
			"id":    schema.StringAttribute{Computed: true},
			"level": schema.StringAttribute{Optional: true},
			"includes": schema.ListAttribute{
				ElementType: types.StringType, Optional: true,
			},
			"excludes": schema.ListAttribute{
				ElementType: types.StringType, Optional: true,
			},
			"resources": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"model": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}
