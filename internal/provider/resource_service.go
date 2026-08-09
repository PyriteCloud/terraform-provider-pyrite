package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	connect "connectrpc.com/connect"
	servicesv1 "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1"
	servicesv1connect "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1/servicesv1connect"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ServiceResource{}
var _ resource.ResourceWithImportState = &ServiceResource{}

func NewServiceResource() resource.Resource {
	return &ServiceResource{}
}

// ServiceResource defines the resource implementation.
type ServiceResource struct {
	client servicesv1connect.ServicesServiceClient
}

// ServiceResourceModel describes the resource data model.
type ServiceResourceModel struct {
	Id        types.String `tfsdk:"id"`
	Name      types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
	Type      types.String `tfsdk:"type"`
}

func (r *ServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service"
}

func (r *ServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Pyrite service",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Service ID",
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
			"type": schema.StringAttribute{
				MarkdownDescription: "Service type",
				Required:            true,
			},
		},
	}
}

func (r *ServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := connect.NewRequest(&servicesv1.CreateServiceDto{
		ProjectId: data.ProjectId.ValueString(),
		Name:      data.Name.ValueString(),
		Type:      data.Type.ValueString(),
	})

	serviceRes, err := r.client.CreateService(ctx, request)

	fmt.Println(serviceRes, err)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create service, got error: %s", err))
		return
	}

	data.Id = types.StringValue(serviceRes.Msg.Id)
	data.Name = types.StringValue(serviceRes.Msg.Name)

	tflog.Trace(ctx, "created a service", map[string]any{
		"service_id":   serviceRes.Msg.Id,
		"service_name": serviceRes.Msg.Name,
		"project_id":   serviceRes.Msg.ProjectId,
		"service_type": serviceRes.Msg.Type,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := connect.NewRequest(&servicesv1.ServiceById{
		Id: data.Id.ValueString(),
	})

	serviceRes, err := r.client.FindOneService(ctx, request)

	fmt.Println(serviceRes, err)

	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "service no longer exists", map[string]any{
				"service_id": data.Id.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read service, got error: %s", err))
		return
	}

	data.Name = types.StringValue(serviceRes.Msg.Name)

	tflog.Trace(ctx, "read a service", map[string]any{
		"service_id":   serviceRes.Msg.Id,
		"service_name": serviceRes.Msg.Name,
		"project_id":   serviceRes.Msg.ProjectId,
		"service_type": serviceRes.Msg.Type,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Service cannot be updated",
		"Pyrite services are immutable. Any configuration change requires resource replacement.",
	)
}

func (r *ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := connect.NewRequest(&servicesv1.ServiceById{
		Id: data.Id.ValueString(),
	})

	_, err := r.client.DeleteService(ctx, request)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete service, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted a service", map[string]any{
		"service_id": data.Id.ValueString(),
	})
}

func (r *ServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
