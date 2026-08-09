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
	projectsv1 "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/projects/v1"
	projectsv1connect "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/projects/v1/projectsv1connect"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

// ProjectResource defines the resource implementation.
type ProjectResource struct {
	client projectsv1connect.ProjectServiceClient
}

// ProjectResourceModel describes the resource data model.
type ProjectResourceModel struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	TeamId types.String `tfsdk:"team_id"`
}

func (r *ProjectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Pyrite project",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Project ID",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Project name",
				Required:            true,
			},
			"team_id": schema.StringAttribute{
				MarkdownDescription: "Team ID",
				Required:            true,
			},
		},
	}
}

func (r *ProjectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	projectServiceClient := projectsv1connect.NewProjectServiceClient(client, PYRITE_API_BASE_URL, connect.WithGRPC())

	r.client = projectServiceClient
}

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := connect.NewRequest(&projectsv1.CreateProjectDto{
		TeamId: data.TeamId.ValueString(),
		Name:   data.Name.ValueString(),
	})

	projectRes, err := r.client.CreateProject(ctx, request)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create project, got error: %s", err))
		return
	}

	data.Id = types.StringValue(projectRes.Msg.Id)
	data.Name = types.StringValue(projectRes.Msg.Name)
	data.TeamId = types.StringValue(projectRes.Msg.TeamId)

	tflog.Trace(ctx, "created a project", map[string]any{
		"project_id":   projectRes.Msg.Id,
		"project_name": projectRes.Msg.Name,
		"team_id":      projectRes.Msg.TeamId,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := connect.NewRequest(&projectsv1.ProjectById{
		Id: data.Id.ValueString(),
	})

	projectRes, err := r.client.FindOneProject(ctx, request)

	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			tflog.Warn(ctx, "project no longer exists", map[string]any{
				"project_id": data.Id.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read project, got error: %s", err))
		return
	}

	data.Name = types.StringValue(projectRes.Msg.Name)
	data.TeamId = types.StringValue(projectRes.Msg.TeamId)

	tflog.Trace(ctx, "read a project", map[string]any{
		"project_id":   projectRes.Msg.Id,
		"project_name": projectRes.Msg.Name,
		"team_id":      projectRes.Msg.TeamId,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := connect.NewRequest(&projectsv1.UpdateProjectDto{
		Id:   data.Id.ValueString(),
		Name: data.Name.ValueStringPointer(),
	})

	projectRes, err := r.client.UpdateProject(ctx, request)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update project, got error: %s", err))
		return
	}

	data.Id = types.StringValue(projectRes.Msg.Id)
	data.Name = types.StringValue(projectRes.Msg.Name)
	data.TeamId = types.StringValue(projectRes.Msg.TeamId)

	tflog.Trace(ctx, "updated a project", map[string]any{
		"project_id":   projectRes.Msg.Id,
		"project_name": projectRes.Msg.Name,
		"team_id":      projectRes.Msg.TeamId,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProjectResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	request := connect.NewRequest(&projectsv1.ProjectById{
		Id: data.Id.ValueString(),
	})

	_, err := r.client.DeleteProject(ctx, request)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete project, got error: %s", err))
		return
	}

	tflog.Trace(ctx, "deleted a project", map[string]any{
		"project_id": data.Id.ValueString(),
	})
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
