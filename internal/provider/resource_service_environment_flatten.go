package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	commonv1 "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1/common/v1"
	deploymentsv1 "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1/deployments/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	t "github.com/julien040/go-ternary"
)

func flattenService(
	ctx context.Context,
	data *ServiceEnvironmentResourceModel,
	serviceEnvironment *commonv1.ServiceEnvironment,
) {
	data.Id = types.StringValue(serviceEnvironment.Id)
	data.Name = types.StringValue(serviceEnvironment.Name)
	data.ProjectId = types.StringValue(serviceEnvironment.Meta.Project.Id)
	data.Type = types.StringValue(serviceEnvironment.Meta.Service.Type)
	data.Environment = types.StringValue(serviceEnvironment.Name)

	flattenDockerConfig(
		ctx,
		data.DockerConfig,
		serviceEnvironment.GetDockerDeployment(),
	)

	flattenPostgresConfig(
	    data.PostgresConfig,
	    serviceEnvironment.GetPostgresDeployment(),
	)
}

func flattenDockerConfig(
	ctx context.Context,
	dockerConfigData *DockerConfigModel,
	dockerConfig *deploymentsv1.DockerDeployment,
) {
	dockerConfigData.Image = &DeploymentImageModel{
		Ref: t.If(
			dockerConfig.Image != nil,
			types.StringValue(*dockerConfig.Image),
			types.StringNull(),
		),
		RegistryID: t.If(
			dockerConfig.RegistryId != nil,
			types.StringValue(*dockerConfig.RegistryId),
			types.StringNull(),
		),
	}

	dockerConfigData.Git = &DeploymentGitModel{
		Url: t.If(
			dockerConfig.GitRepoUrl != nil,
			types.StringValue(*dockerConfig.GitRepoUrl),
			types.StringNull(),
		),
		Branch: t.If(
			dockerConfig.GitBranch != nil,
			types.StringValue(*dockerConfig.GitBranch),
			types.StringNull(),
		),
		Sha: t.If(
			dockerConfig.GitSha != nil,
			types.StringValue(*dockerConfig.GitSha),
			types.StringNull(),
		),

		Build: &DeploymentBuildModel{
			Builder: types.StringValue(
				dockerConfig.BuildConfig.Builder,
			),
			Context: types.StringValue(
				dockerConfig.BuildConfig.Context,
			),
			DockerfilePath: types.StringValue(
				dockerConfig.BuildConfig.DockerfilePath,
			),
		},

		WithBuild: t.If(
			dockerConfig.WithBuild != nil,
			types.BoolValue(*dockerConfig.WithBuild),
			types.BoolNull(),
		),
	}

	dockerConfigData.Runtime =
		types.StringValue(dockerConfig.Runtime)

	dockerConfigData.SourceType =
		types.StringValue(dockerConfig.SourceType)

	dockerConfigData.Plan =
		types.StringValue(dockerConfig.Plan)

	dockerConfigData.IsPrivate =
		types.BoolValue(dockerConfig.IsPrivate)

	dockerConfigData.IsPrivileged =
		types.BoolValue(dockerConfig.IsPrivileged)

	dockerConfigData.WithProjectEnv = t.If(
		dockerConfig.WithProjectEnv != nil,
		types.BoolValue(*dockerConfig.WithProjectEnv),
		types.BoolNull(),
	)

	dockerConfigData.Args = t.If(
		dockerConfig.Args != nil,
		types.StringValue(*dockerConfig.Args),
		types.StringNull(),
	)

	dockerConfigData.Command = t.If(
		dockerConfig.Command != nil,
		types.StringValue(*dockerConfig.Command),
		types.StringNull(),
	)

	flattenEnv(
		ctx,
		dockerConfig.Env,
		&dockerConfigData.Env,
	)

	flattenHealthChecks(
		dockerConfig.HealthChecks,
		&dockerConfigData.HealthChecksList,
	)

	flattenRegions(
		dockerConfig.Regions,
		&dockerConfigData.RegionsList,
	)

	flattenVolumes(
		dockerConfig.Volumes,
		&dockerConfigData.VolumesList,
	)

	flattenPorts(
		dockerConfig.Ports,
		&dockerConfigData.PortsList,
	)
}

func flattenHealthChecks(
	healthChecks []*deploymentsv1.DeploymentHealthCheck,
	target *[]DeploymentHealthCheckModel,
) {
	result := make([]DeploymentHealthCheckModel, 0, len(healthChecks))

	for _, healthCheck := range healthChecks {
		if healthCheck == nil {
			continue
		}

		result = append(result, DeploymentHealthCheckModel{
			Path: t.If(
				healthCheck.Path != nil,
				types.StringValue(*healthCheck.Path),
				types.StringNull(),
			),
			Port: types.Int64Value(
				int64(healthCheck.Port),
			),
			Interval: types.Int64Value(
				int64(healthCheck.Interval),
			),
		})
	}

	*target = result
}

func flattenRegions(
	regions []*deploymentsv1.DeploymentRegion,
	target *[]DeploymentRegionModel,
) {
	result := make([]DeploymentRegionModel, 0, len(regions))

	for _, region := range regions {
		if region == nil {
			continue
		}

		result = append(result, DeploymentRegionModel{
			Region: types.StringValue(region.Region),
			MinReplicas: types.Int64Value(
				int64(region.MinReplicas),
			),
			MaxReplicas: types.Int64Value(
				int64(region.MaxReplicas),
			),
		})
	}

	*target = result
}

func flattenVolumes(
	volumes []*deploymentsv1.DeploymentVolume,
	target *[]DeploymentVolumeModel,
) {
	result := make([]DeploymentVolumeModel, 0, len(volumes))

	for _, volume := range volumes {
		if volume == nil {
			continue
		}

		result = append(result, DeploymentVolumeModel{
			MountPath: types.StringValue(volume.MountPath),

			TeamVolumeId: t.If(
				volume.TeamVolumeId != nil,
				types.StringValue(*volume.TeamVolumeId),
				types.StringNull(),
			),
		})
	}

	*target = result
}

func flattenPorts(
    ports []*deploymentsv1.DeploymentPort,
    target *[]DeploymentPortModel,
) {
    result := make([]DeploymentPortModel, 0, len(ports))

    for _, port := range ports {
        if port == nil {
            continue
        }

        result = append(result, DeploymentPortModel{
            Port: types.Int64Value(int64(port.Port)),
			Protocol: types.StringValue(port.Protocol),
			Path: types.StringValue(port.Path),
        })
    }

    *target = result
}

func flattenEnv(
	ctx context.Context,
	env *string,
	target *types.Map,
) error {
	if env == nil {
		*target = types.MapNull(types.StringType)
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*env)
	if err != nil {
		return fmt.Errorf("failed to decode env base64: %w", err)
	}

	var values map[string]string

	if err := json.Unmarshal(decoded, &values); err != nil {
		return fmt.Errorf("failed to decode env JSON: %w", err)
	}

	result, diagnostics := types.MapValueFrom(
		ctx,
		types.StringType,
		values,
	)

	if diagnostics.HasError() {
		return fmt.Errorf(
			"failed to create Terraform env map: %s",
			diagnostics.Errors()[0].Summary(),
		)
	}

	*target = result

	return nil
}

func flattenPostgresConfig(postgresConfigData *PostgresConfigModel, postgresConfig *deploymentsv1.PostgresDeployment) {
	// Implementation for flattening Postgres config

	postgresConfigData.Version = types.StringValue(postgresConfig.Version)
	postgresConfigData.Plan = types.StringValue(postgresConfig.Plan)
	postgresConfigData.Region = types.StringValue(postgresConfig.Region)
	postgresConfigData.Size = types.Int64Value(int64(postgresConfig.Size))
}