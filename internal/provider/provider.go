package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	openapi "github.com/csc478-wcu/fabric-orchestrator-go-client"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/client"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/fabric_client_wrapper"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ provider.Provider = &FabricProvider{}

// FabricProvider defines the provider implementation.
type FabricProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// FabricProviderModel describes the provider data model.
type FabricProviderModel struct {
	Token    types.String `tfsdk:"token"`
	Endpoint types.String `tfsdk:"endpoint"`
	SSHKey   types.String `tfsdk:"ssh_key"`
}

func (p *FabricProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fabric"
	resp.Version = p.version
}

func (p *FabricProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Fabric provider allows you to manage FABRIC testbed resources using Terraform.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "The FABRIC API token. Can also be set via the FABRIC_TOKEN environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The FABRIC Orchestrator API endpoint. Defaults to https://orchestrator.fabric-testbed.net",
				Optional:            true,
			},
			"ssh_key": schema.StringAttribute{
				MarkdownDescription: "Default SSH public key for instances. Can also be set via FABRIC_SSH_KEY environment variable.",
				Optional:            true,
			},
		},
	}
}

func (p *FabricProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data FabricProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Default configuration values
	token := os.Getenv("FABRIC_TOKEN")
	endpoint := "https://orchestrator.fabric-testbed.net"
	sshKey := os.Getenv("FABRIC_SSH_KEY")

	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	if !data.Endpoint.IsNull() {
		endpoint = data.Endpoint.ValueString()
	}

	if !data.SSHKey.IsNull() {
		sshKey = data.SSHKey.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing API Token",
			"The provider cannot create the FABRIC API client as there is a missing or empty value for the token. "+
				"Set the token value in the configuration or use the FABRIC_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create API client
	cfg := openapi.NewConfiguration()
	cfg.Servers = openapi.ServerConfigurations{
		{
			URL: endpoint,
		},
	}

	apiClient := fabric_client_wrapper.NewFabricClientWrapper(cfg)
	fabricClient := client.NewFabricClient(apiClient, token)

	// Create provider client data
	providerData := &FabricProviderData{
		Client:        fabricClient,
		Endpoint:      endpoint,
		DefaultSSHKey: sshKey,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *FabricProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSliceResource,
	}
}

func (p *FabricProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewResourcesDataSource,
		NewSitesDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &FabricProvider{
			version: version,
		}
	}
}

// FabricProviderData contains the configured client and settings
type FabricProviderData struct {
	Client        client.FabricClient
	Endpoint      string
	DefaultSSHKey string
}
