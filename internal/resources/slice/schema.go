package slice

import (
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Schema() schema.Schema {
	return schema.Schema{
		MarkdownDescription: "Manages a FABRIC slice.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Slice identifier.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Slice name.",
				Required:            true,
			},
			"lease_end_time": schema.StringAttribute{
				MarkdownDescription: "Lease end time (RFC3339). Defaults to now+24h.",
				Optional:            true,
				Computed:            true,
			},
			"ssh_keys": schema.ListAttribute{
				MarkdownDescription: "SSH public keys.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Current slice state.",
				Computed:            true,
			},
			"sliver_count": schema.Int64Attribute{
				MarkdownDescription: "Number of slivers in the slice.",
				Computed:            true,
			},
			"topology": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"nodes": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name":          schema.StringAttribute{Required: true},
								"site":          schema.StringAttribute{Required: true},
								"type":          schema.StringAttribute{Optional: true, Computed: true},
								"image_ref":     schema.StringAttribute{Optional: true, Computed: true},
								"instance_type": schema.StringAttribute{Optional: true, Computed: true},
								"cores":         schema.Int64Attribute{Optional: true, Computed: true},
								"ram":           schema.Int64Attribute{Optional: true, Computed: true},
								"disk":          schema.Int64Attribute{Optional: true, Computed: true},
							},
						},
					},
					"links": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"name":   schema.StringAttribute{Required: true},
								"source": schema.StringAttribute{Required: true},
								"target": schema.StringAttribute{Required: true},
							},
						},
					},
				},
			},
		},
	}
}
