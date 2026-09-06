package provider

import (
	"context"
	"fmt"

	connect "connectrpc.com/connect"
	servicesv1 "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r *ServiceEnvironmentResource) upsertServiceEnv(
	ctx context.Context,
	data *ServiceEnvironmentResourceModel,
	diags *diag.Diagnostics,
) {
	request := buildServiceRequest(data)

	if data.DockerConfig != nil {
		dockerConfig, err := buildDockerDeploymentConfig(
			ctx,
			data.DockerConfig,
		)
		if err != nil {
			diags.AddError(
				"Invalid Docker configuration",
				err.Error(),
			)
			return
		}

		request.DeploymentConfig = &servicesv1.UpsertServiceDto_DockerConfig{
			DockerConfig: dockerConfig,
		}
	} else {
		request.DeploymentConfig = &servicesv1.UpsertServiceDto_PostgresConfig{
			PostgresConfig: buildPostgresDeploymentConfig(
				data.PostgresConfig,
			),
		}
	}

	serviceEnvRes, err := r.client.UpsertService(
		ctx,
		connect.NewRequest(request),
	)
	if err != nil {
		diags.AddError(
			"Client Error",
			fmt.Sprintf(
				"Unable to upsert service environment: %v",
				err,
			),
		)
		return
	}

	data.Id = types.StringValue(
		serviceEnvRes.Msg.Service.Id,
	)
}
