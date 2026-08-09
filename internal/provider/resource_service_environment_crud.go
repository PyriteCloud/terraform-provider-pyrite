package provider

import (
	"context"
	"fmt"

	connect "connectrpc.com/connect"
	servicesv1 "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func (r *ServiceEnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceEnvironmentResourceModel

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

func (r *ServiceEnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceEnvironmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.upsertServiceEnv(ctx, &data, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceEnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServiceEnvironmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	r.upsertServiceEnv(ctx, &data, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceEnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceEnvironmentResourceModel

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

func (r *ServiceEnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
