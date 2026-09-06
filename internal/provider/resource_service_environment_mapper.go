package provider

import (
	"context"

	servicesv1 "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1"
	deploymentsv1 "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1/deployments/v1"
)

// buildServiceRequest creates the base UpsertServiceDto shared by both Docker and Postgres deployments.
func buildServiceRequest(data *ServiceEnvironmentResourceModel) *servicesv1.UpsertServiceDto {
	return &servicesv1.UpsertServiceDto{
		ProjectId:   data.ProjectId.ValueString(),
		Name:        data.Name.ValueString(),
		Type:        data.Type.ValueString(),
		Environment: data.Environment.ValueStringPointer(),
	}
}

// buildDockerSource converts the Terraform image/git configuration into a protobuf DockerDeploymentSourceDto.
// Validation has already been handled by ConfigValidators.
func buildDockerSource(cfg *DockerConfigModel) *deploymentsv1.DockerDeploymentSourceDto {
	if cfg.Image != nil {
		return &deploymentsv1.DockerDeploymentSourceDto{
			Image: &deploymentsv1.DockerDeploymentImageSourceDto{
				Ref: cfg.Image.Ref.ValueString(),
			},
		}
	}

	git := cfg.Git

	var buildCfg *deploymentsv1.DockerDeploymentBuildConfigDto
	if git.Build != nil {
		builder := git.Build.Builder.ValueString()
		contextPath := git.Build.Context.ValueStringPointer()
		dockerfilePath := git.Build.DockerfilePath.ValueStringPointer()

		buildCfg = &deploymentsv1.DockerDeploymentBuildConfigDto{
			Builder:        builder,
			Context:        contextPath,
			DockerfilePath: dockerfilePath,
		}
	}

	sha := git.Sha.ValueStringPointer()
	withBuild := git.WithBuild.ValueBoolPointer()

	return &deploymentsv1.DockerDeploymentSourceDto{
		Git: &deploymentsv1.DockerDeploymentGitSourceDto{
			RepoUrl:   git.Url.ValueString(),
			Branch:    git.Branch.ValueString(),
			Sha:       sha,
			WithBuild: withBuild,
			Build:     buildCfg,
		},
	}
}

// buildDockerPrimitiveValues converts Terraform primitive values into Go values
// ready for protobuf DTO construction.
func buildDockerPrimitiveValues(
	ctx context.Context,
	cfg *DockerConfigModel,
) (dockerPrimitiveValues, error) {
	env, err := buildEnvValue(ctx, cfg.Env)
	if err != nil {
		return dockerPrimitiveValues{}, err
	}

	values := dockerPrimitiveValues{
		Runtime:        cfg.Runtime.ValueString(),
		SourceType:     cfg.SourceType.ValueString(),
		Plan:           cfg.Plan.ValueString(),
		IsPrivate:      cfg.IsPrivate.ValueBoolPointer(),
		IsPrivileged:   cfg.IsPrivileged.ValueBoolPointer(),
		WithProjectEnv: cfg.WithProjectEnv.ValueBoolPointer(),
		Command:        cfg.Command.ValueStringPointer(),
		Args:           cfg.Args.ValueStringPointer(),
		Env:            env,
	}

	return values, nil
}

// buildDockerDeploymentConfig converts the Terraform Docker configuration
// into the protobuf DockerDeploymentDto used by the API.
func buildDockerDeploymentConfig(
	ctx context.Context,
	cfg *DockerConfigModel,
) (*deploymentsv1.DockerDeploymentDto, error) {
	values, err := buildDockerPrimitiveValues(ctx, cfg)
	if err != nil {
		return nil, err
	}

	source := buildDockerSource(cfg)

	return &deploymentsv1.DockerDeploymentDto{
		SourceType: values.SourceType,
		Source:     source,
		Runtime:    values.Runtime,
		Plan:       values.Plan,

		Command:        values.Command,
		Args:           values.Args,
		Env:            values.Env,
		WithProjectEnv: values.WithProjectEnv,
		IsPrivate:      values.IsPrivate,
		IsPrivileged:   values.IsPrivileged,

		PortsList: &deploymentsv1.DeploymentPortList{
			Ports: buildPortDtos(cfg.PortsList),
		},
		RegionsList: &deploymentsv1.DeploymentRegionList{
			Regions: buildRegionDtos(cfg.RegionsList),
		},
		FilesList: &deploymentsv1.DeploymentFileList{
			Files: buildFileDtos(cfg.FilesList),
		},
		VolumesList: &deploymentsv1.DeploymentVolumeList{
			Volumes: buildVolumeDtos(cfg.VolumesList),
		},
		HealthChecksList: &deploymentsv1.DeploymentHealthCheckList{
			HealthChecks: buildHealthCheckDtos(cfg.HealthChecksList),
		},
	}, nil
}

// buildPostgresDeploymentConfig converts the Terraform Postgres configuration
// into the protobuf PostgresDeploymentDto used by the API.
func buildPostgresDeploymentConfig(cfg *PostgresConfigModel) *deploymentsv1.PostgresDeploymentDto {
	return &deploymentsv1.PostgresDeploymentDto{
		Version:  cfg.Version.ValueString(),
		Plan:     cfg.Plan.ValueString(),
		Region:   cfg.Region.ValueString(),
		Size:     int32(cfg.Size.ValueInt64()),
		Password: cfg.Password.ValueString(),
	}
}
