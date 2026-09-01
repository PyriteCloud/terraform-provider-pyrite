package provider

import "github.com/hashicorp/terraform-plugin-framework/types"

// ServiceEnvironmentResourceModel describes the resource data model.
type ServiceEnvironmentResourceModel struct {
	Id             types.String         `tfsdk:"id"`
	Name           types.String         `tfsdk:"name"`
	Environment    types.String         `tfsdk:"environment"`
	ProjectId      types.String         `tfsdk:"project_id"`
	Type           types.String         `tfsdk:"type"`
	DockerConfig   *DockerConfigModel   `tfsdk:"docker_config"`
	PostgresConfig *PostgresConfigModel `tfsdk:"postgres_config"`
}

type DockerConfigModel struct {
	Image *DeploymentImageModel `tfsdk:"image"`
	Git   *DeploymentGitModel   `tfsdk:"git"`

	PortsList        []DeploymentPortModel        `tfsdk:"ports"`
	RegionsList      []DeploymentRegionModel      `tfsdk:"regions"`
	FilesList        []DeploymentFileModel        `tfsdk:"files"`
	VolumesList      []DeploymentVolumeModel      `tfsdk:"volumes"`
	HealthChecksList []DeploymentHealthCheckModel `tfsdk:"health_checks"`

	Runtime        types.String `tfsdk:"runtime"`
	SourceType     types.String `tfsdk:"source_type"`
	Plan           types.String `tfsdk:"plan"`
	Command        types.String `tfsdk:"command"`
	Args           types.String `tfsdk:"args"`
	Env            types.String `tfsdk:"env"`
	IsPrivate      types.Bool   `tfsdk:"is_private"`
	IsPrivileged   types.Bool   `tfsdk:"is_privileged"`
	WithProjectEnv types.Bool   `tfsdk:"with_project_env"`
}

type PostgresConfigModel struct {
	Version  types.String `tfsdk:"version"`
	Plan     types.String `tfsdk:"plan"`
	Region   types.String `tfsdk:"region"`
	Size     types.Int64  `tfsdk:"size"`
	Password types.String `tfsdk:"password"`
}

type DeploymentPortModel struct {
	Port     types.Int64  `tfsdk:"port"`
	Protocol types.String `tfsdk:"protocol"`
	Path     types.String `tfsdk:"path"`
}

type DeploymentRegionModel struct {
	Region      types.String `tfsdk:"region"`
	MinReplicas types.Int64  `tfsdk:"min_replicas"`
	MaxReplicas types.Int64  `tfsdk:"max_replicas"`
}

type DeploymentHealthCheckModel struct {
	Port         types.Int64  `tfsdk:"port"`
	Protocol     types.String `tfsdk:"protocol"`
	Path         types.String `tfsdk:"path"`
	InitialDelay types.Int64  `tfsdk:"initial_delay"`
	Interval     types.Int64  `tfsdk:"interval"`
	Timeout      types.Int64  `tfsdk:"timeout"`
	MaxFailures  types.Int64  `tfsdk:"max_failures"`
}

type DeploymentVolumeModel struct {
	MountPath    types.String `tfsdk:"mount_path"`
	TeamVolumeId types.String `tfsdk:"team_volume_id"`
}

type DeploymentFileModel struct {
	Path        types.String `tfsdk:"path"`
	Content     types.String `tfsdk:"content"`
	Permissions types.String `tfsdk:"permissions"`
}

type DeploymentImageModel struct {
	Ref        types.String `tfsdk:"ref"`
	RegistryID types.String `tfsdk:"registry_id"`
}

type DeploymentGitModel struct {
	Url       types.String          `tfsdk:"url"`
	Branch    types.String          `tfsdk:"branch"`
	Sha       types.String          `tfsdk:"sha"`
	Build     *DeploymentBuildModel `tfsdk:"build"`
	WithBuild types.Bool            `tfsdk:"with_build"`
}

type DeploymentBuildModel struct {
	Builder        types.String `tfsdk:"builder"`
	Context        types.String `tfsdk:"context"`
	DockerfilePath types.String `tfsdk:"dockerfile_path"`
}

type dockerPrimitiveValues struct {
	Runtime        string
	SourceType     string
	Plan           string
	Command        *string
	Args           *string
	Env            *string
	IsPrivate      *bool
	IsPrivileged   *bool
	WithProjectEnv *bool
}
