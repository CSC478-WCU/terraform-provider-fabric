package provider

import (
	"context"

	"github.com/csc478-wcu/terraform-provider-fabric/internal/clients/orchestrator"
	resourcesds "github.com/csc478-wcu/terraform-provider-fabric/internal/datasources/resources"
	sitesds "github.com/csc478-wcu/terraform-provider-fabric/internal/datasources/sites"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/resources/slice"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/runtime"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/services"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	pframework "github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ pframework.Provider = &FabricProvider{}

type FabricProvider struct {
	version string
}

type FabricProviderModel struct {
	Token    types.String `tfsdk:"token"`
	Endpoint types.String `tfsdk:"endpoint"`
	SSHKey   types.String `tfsdk:"ssh_key"`
}

func (p *FabricProvider) Metadata(_ context.Context, req pframework.MetadataRequest, resp *pframework.MetadataResponse) {
	resp.TypeName = "fabric"
	resp.Version = p.version
}

func (p *FabricProvider) Schema(_ context.Context, _ pframework.SchemaRequest, resp *pframework.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "FABRIC provider for managing testbed slices and resources.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "FABRIC API token (or FABRIC_TOKEN).",
				Optional:            true,
				Sensitive:           true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "FABRIC Orchestrator API endpoint.",
				Optional:            true,
			},
			"ssh_key": schema.StringAttribute{
				MarkdownDescription: "Default SSH public key (or FABRIC_SSH_KEY).",
				Optional:            true,
			},
		},
	}
}

func (p *FabricProvider) Configure(ctx context.Context, req pframework.ConfigureRequest, resp *pframework.ConfigureResponse) {
	var cfg FabricProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := getenvOr("FABRIC_TOKEN", "")
	if !cfg.Token.IsNull() {
		token = cfg.Token.ValueString()
	}
	if token == "" {
		resp.Diagnostics.AddError("Missing API Token",
			"Provide 'token' in configuration or set FABRIC_TOKEN.")
		return
	}

	endpoint := defaultEndpoint
	if !cfg.Endpoint.IsNull() && cfg.Endpoint.ValueString() != "" {
		endpoint = cfg.Endpoint.ValueString()
	}
	sshKey := getenvOr("FABRIC_SSH_KEY", "")
	if !cfg.SSHKey.IsNull() && cfg.SSHKey.ValueString() != "" {
		sshKey = cfg.SSHKey.ValueString()
	}

	orc := orchestrator.New(orchestrator.Config{
		Endpoint: endpoint,
		Token:    token,
	})
	slicesSvc := services.NewSlicesService(orc)
	resSvc := services.NewResourcesService(orc)

	deps := &runtime.Deps{
		Slices:        slicesSvc,
		Resources:     resSvc,
		DefaultSSHKey: sshKey,
		Endpoint:      endpoint,
	}

	resp.DataSourceData = deps
	resp.ResourceData = deps
}

func (p *FabricProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return slice.New() },
	}
}

func (p *FabricProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return resourcesds.New() },
		func() datasource.DataSource { return sitesds.New() },
	}
}

func New(v string) func() pframework.Provider {
	return func() pframework.Provider {
		return &FabricProvider{version: v}
	}
}
