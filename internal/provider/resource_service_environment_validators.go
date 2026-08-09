package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// -----------------------------------------------------------------------------
// Shared helper
// -----------------------------------------------------------------------------

func loadServiceEnvConfig(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) *ServiceEnvironmentResourceModel {
	var data ServiceEnvironmentResourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return nil
	}

	return &data
}

// -----------------------------------------------------------------------------
// Deployment config validator
// -----------------------------------------------------------------------------

type deploymentConfigValidator struct{}

// Description, MarkdownDescription, ValidateResource...

func (v deploymentConfigValidator) Description(ctx context.Context) string {
	return "Exactly one deployment configuration must be provided."
}

func (v deploymentConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v deploymentConfigValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	data := loadServiceEnvConfig(ctx, req, resp)
	if data == nil {
		return
	}

	isDocker := data.DockerConfig != nil
	isPostgres := data.PostgresConfig != nil

	if isDocker == isPostgres {
		resp.Diagnostics.AddError(
			"Invalid deployment configuration",
			"Exactly one of docker_config or postgres_config must be provided.",
		)
	}
}

// -----------------------------------------------------------------------------
// Docker source validator
// -----------------------------------------------------------------------------

type sourceValidator struct{}

// Description, MarkdownDescription, ValidateResource...

func (v sourceValidator) Description(ctx context.Context) string {
	return "Validate Docker source configuration."
}

func (v sourceValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v sourceValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	data := loadServiceEnvConfig(ctx, req, resp)
	if data == nil || data.DockerConfig == nil {
		return
	}

	hasImage := data.DockerConfig.Image != nil &&
		!data.DockerConfig.Image.Ref.IsNull() &&
		!data.DockerConfig.Image.Ref.IsUnknown()

	hasGit := data.DockerConfig.Git != nil &&
		!data.DockerConfig.Git.Url.IsNull() &&
		!data.DockerConfig.Git.Url.IsUnknown()

	// Exactly one source must be configured
	if hasImage == hasGit {
		resp.Diagnostics.AddError(
			"Invalid Docker source",
			"Exactly one of image or git must be provided inside docker_config.",
		)
		return
	}

	// Skip validation if source_type is not known yet
	if data.DockerConfig.SourceType.IsNull() || data.DockerConfig.SourceType.IsUnknown() {
		return
	}

	switch data.DockerConfig.SourceType.ValueString() {
	case "image":
		if !hasImage {
			resp.Diagnostics.AddError(
				"Image block required",
				"docker_config.image must be set when source_type = image.",
			)
		}

	case "git":
		if !hasGit {
			resp.Diagnostics.AddError(
				"Git block required",
				"docker_config.git must be set when source_type = git.",
			)
		}
	}
}

// -----------------------------------------------------------------------------
// Git build validator
// -----------------------------------------------------------------------------

type buildValidator struct{}

// Description, MarkdownDescription, ValidateResource...

func (v buildValidator) Description(ctx context.Context) string {
	return "Validate Git build configuration."
}

func (v buildValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v buildValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	data := loadServiceEnvConfig(ctx, req, resp)
	if data == nil || data.DockerConfig == nil || data.DockerConfig.Git == nil {
		return
	}

	git := data.DockerConfig.Git

	// build block requires with_build = true
	if git.Build != nil && !git.WithBuild.ValueBool() {
		resp.Diagnostics.AddError(
			"Invalid build configuration",
			"git.build can only be used when git.with_build = true.",
		)
	}

	if git.Build != nil {
		if git.Build.Context.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Invalid build context",
				"git.build.context cannot be empty.",
			)
		}

		if git.Build.DockerfilePath.ValueString() == "" {
			resp.Diagnostics.AddError(
				"Invalid Dockerfile path",
				"git.build.dockerfile_path cannot be empty.",
			)
		}
	}
}

// -----------------------------------------------------------------------------
// Ports validator
// -----------------------------------------------------------------------------

type portsValidator struct{}

// Description, MarkdownDescription, ValidateResource...

func (v portsValidator) Description(ctx context.Context) string {
	return "Validate ports and protocols based on service type."
}

func (v portsValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v portsValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	data := loadServiceEnvConfig(ctx, req, resp)
	if data == nil || data.DockerConfig == nil {
		return
	}

	serviceType := data.Type.ValueString()
	ports := data.DockerConfig.PortsList

	// Worker services cannot expose ports
	if serviceType == "worker" && len(ports) > 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("docker_config").AtName("ports"),
			"Ports are not allowed for worker services",
			"Worker services cannot expose ports or protocols.",
		)
		return
	}

	seen := make(map[int64]struct{})

	for i, p := range ports {
		port := p.Port.ValueInt64()
		protocol := p.Protocol.ValueString()

		// Port range
		if port < 1 || port > 65535 {
			resp.Diagnostics.AddAttributeError(
				path.Root("docker_config").
					AtName("ports").
					AtListIndex(i).
					AtName("port"),
				"Invalid port number",
				fmt.Sprintf("Port %d must be between 1 and 65535.", port),
			)
		}

		// Duplicate ports
		if _, ok := seen[port]; ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("docker_config").
					AtName("ports").
					AtListIndex(i).
					AtName("port"),
				"Duplicate port",
				fmt.Sprintf("Port %d is defined more than once.", port),
			)
		}
		seen[port] = struct{}{}

		// Protocol rules
		switch serviceType {
		case "web":
			if protocol != "http1" && protocol != "http2" {
				resp.Diagnostics.AddAttributeError(
					path.Root("docker_config").
						AtName("ports").
						AtListIndex(i).
						AtName("protocol"),
					"Invalid protocol for web service",
					"Web services only support http1 or http2.",
				)
			}

		case "pod":
			if protocol != "tcp" {
				resp.Diagnostics.AddAttributeError(
					path.Root("docker_config").
						AtName("ports").
						AtListIndex(i).
						AtName("protocol"),
					"Invalid protocol for pod service",
					"Pod services only support tcp.",
				)
			}
		}
	}
}

// -----------------------------------------------------------------------------
// Regions validator
// -----------------------------------------------------------------------------

type regionValidator struct{}

// Description, MarkdownDescription, ValidateResource...

func (v regionValidator) Description(ctx context.Context) string {
	return "Validate deployment regions."
}

func (v regionValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v regionValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	data := loadServiceEnvConfig(ctx, req, resp)
	if data == nil || data.DockerConfig == nil {
		return
	}

	seen := make(map[string]struct{})

	for i, r := range data.DockerConfig.RegionsList {
		region := r.Region.ValueString()

		if _, ok := seen[region]; ok {
			resp.Diagnostics.AddAttributeError(
				path.Root("docker_config").
					AtName("regions").
					AtListIndex(i).
					AtName("region"),
				"Duplicate region",
				"Each region may only be specified once.",
			)
		}

		if r.MinReplicas.ValueInt64() > r.MaxReplicas.ValueInt64() {
			resp.Diagnostics.AddAttributeError(
				path.Root("docker_config").
					AtName("regions").
					AtListIndex(i),
				"Invalid replica configuration",
				"min_replicas cannot be greater than max_replicas.",
			)
		}

		seen[region] = struct{}{}
	}
}
