// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/http"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/julien040/go-ternary"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	PYRITE_API_BASE_URL   = "https://api-grpc.pyrite.cloud"
	PYRITE_TOKEN_ENV_NAME = "PYRITE_TOKEN"
)

var uuidPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

// Ensure PyriteProvider satisfies various provider interfaces.
var _ provider.Provider = &PyriteProvider{}
var _ provider.ProviderWithFunctions = &PyriteProvider{}

// PyriteProvider defines the provider implementation.
type PyriteProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PyriteProviderModel describes the provider data model.
type PyriteProviderModel struct {
	TeamId types.String `tfsdk:"team_id"`
	Token  types.String `tfsdk:"token"`
}

func (p *PyriteProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pyrite"
	resp.Version = p.version
}

func (p *PyriteProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team ID for authentication",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidPattern, "Team ID must be a valid UUID"),
				},
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "API token for authentication",
				Required:            true,
				Sensitive:           true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(uuidPattern, "Token must be a valid UUID"),
				},
			},
		},
	}
}

func (p *PyriteProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PyriteProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := ternary.If(data.Token.IsNull(), os.Getenv(PYRITE_TOKEN_ENV_NAME), data.Token.ValueString())

	if token == "" {
		resp.Diagnostics.AddError("Token is empty", "Token must be provided")
		return
	}

	// Example client configuration for data sources and resources
	client := http.Client{
		Transport: &AuthTransport{
			token:   token,
			wrapped: http.DefaultTransport,
		},
	}

	resp.DataSourceData = &client
	resp.ResourceData = &client
}

func (p *PyriteProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
	}
}

func (p *PyriteProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExampleDataSource,
	}
}

func (p *PyriteProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PyriteProvider{
			version: version,
		}
	}
}
