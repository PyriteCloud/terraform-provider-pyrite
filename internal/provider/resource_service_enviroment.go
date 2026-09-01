package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	connect "connectrpc.com/connect"

	servicesv1connect "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1/servicesv1connect"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ServiceEnvironmentResource{}
var _ resource.ResourceWithImportState = &ServiceEnvironmentResource{}
var _ resource.ResourceWithConfigValidators = &ServiceEnvironmentResource{}

func NewServiceEnvironmentResource() resource.Resource {
	return &ServiceEnvironmentResource{}
}

func (r *ServiceEnvironmentResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		deploymentConfigValidator{},
		sourceValidator{},
		buildValidator{},
		portsValidator{},
		regionValidator{},
	}
}

// ServiceEnvironmentResource defines the resource implementation.
type ServiceEnvironmentResource struct {
	client servicesv1connect.ServicesServiceClient
}

func (r *ServiceEnvironmentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_environment"
}

func (r *ServiceEnvironmentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Pyrite service environment",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service Environment ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				MarkdownDescription: "Project ID",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Service name",
				Required:            true,
			},
			"environment": schema.StringAttribute{
				MarkdownDescription: "Environment",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Service type",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"web",
						"pod",
						"worker",
						"postgres",
					),
				},
			},
			"docker_config": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"image": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"ref": schema.StringAttribute{
								Required: true,
							},
							"registry_id": schema.StringAttribute{
								Optional: true,
							},
						},
					},
					"git": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"url": schema.StringAttribute{
								Required: true,
							},
							"branch": schema.StringAttribute{
								Optional: true,
							},
							"sha": schema.StringAttribute{
								Optional: true,
							},
							"with_build": schema.BoolAttribute{
								Optional: true,
							},
							"build": schema.SingleNestedAttribute{
								Optional: true,
								Attributes: map[string]schema.Attribute{
									"builder": schema.StringAttribute{
										Required: true,
									},
									"context": schema.StringAttribute{
										Required: true,
									},
									"dockerfile_path": schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
					"source_type": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"image",
								"git",
							),
						},
					},
					"runtime": schema.StringAttribute{
						Required: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								"docker",
							),
						},
					},
					"is_private": schema.BoolAttribute{
						Optional: true,
					},
					"is_privileged": schema.BoolAttribute{
						Optional: true,
					},
					"with_project_env": schema.BoolAttribute{
						Optional: true,
					},
					"args": schema.StringAttribute{
						Optional: true,
					},
					"command": schema.StringAttribute{
						Optional: true,
					},
					"env": schema.StringAttribute{
						Optional: true,
					},
					"files": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"path": schema.StringAttribute{
									Required: true,
								},
								"content": schema.StringAttribute{
									Required: true,
								},
								"permissions": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
					"health_checks": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"port": schema.Int64Attribute{
									Required: true,
								},
								"protocol": schema.StringAttribute{
									Required: true,
									Validators: []validator.String{
										stringvalidator.OneOf(
											"http1",
											"http2",
											"tcp",
											"udp",
										),
									},
								},
								"path": schema.StringAttribute{
									Optional: true,
								},
								"initial_delay": schema.Int64Attribute{
									Optional: true,
								},
								"interval": schema.Int64Attribute{
									Optional: true,
								},
								"timeout": schema.Int64Attribute{
									Optional: true,
								},
								"max_failures": schema.Int64Attribute{
									Optional: true,
								},
							},
						},
					},
					"plan": schema.StringAttribute{
						Required: true,
					},
					"ports": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"port": schema.Int64Attribute{
									Required: true,
								},
								"protocol": schema.StringAttribute{
									Optional: true,
								},
								"path": schema.StringAttribute{
									Optional: true,
								},
							},
						},
					},
					"regions": schema.ListNestedAttribute{
						Required: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"region": schema.StringAttribute{
									Required: true,
								},
								"min_replicas": schema.Int64Attribute{
									Optional: true,
								},
								"max_replicas": schema.Int64Attribute{
									Optional: true,
								},
							},
						},
					},
					"volumes": schema.ListNestedAttribute{
						Optional: true,
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"mount_path": schema.StringAttribute{
									Required: true,
								},
								"team_volume_id": schema.StringAttribute{
									Required: true,
								},
							},
						},
					},
				},
			},

			"postgres_config": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"version": schema.StringAttribute{
						Required: true,
					},
					"plan": schema.StringAttribute{
						Required: true,
					},
					"region": schema.StringAttribute{
						Required: true,
					},
					"size": schema.Int64Attribute{
						Required: true,
					},
					"password": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
				},
			},
		},
	}
}

func (r *ServiceEnvironmentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*http.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	serviceServiceClient := servicesv1connect.NewServicesServiceClient(client, PYRITE_API_BASE_URL, connect.WithGRPC())

	r.client = serviceServiceClient
}
