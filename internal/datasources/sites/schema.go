package sites

import "github.com/hashicorp/terraform-plugin-framework/datasource/schema"

func Schema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Lists known FABRIC sites (static for now).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true},
			"sites": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":        schema.StringAttribute{Computed: true},
						"code":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"location":    schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}
