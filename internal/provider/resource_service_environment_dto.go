package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	deploymentsv1 "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1/deployments/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	t "github.com/julien040/go-ternary"
)

// buildHealthCheckDtos converts Terraform health check models into protobuf DTOs.
func buildHealthCheckDtos(
	checks []DeploymentHealthCheckModel,
) []*deploymentsv1.DeploymentHealthCheckDto {
	result := make([]*deploymentsv1.DeploymentHealthCheckDto, 0, len(checks))

	for _, h := range checks {
		port := int32(h.Port.ValueInt64())
		protocol := h.Protocol.ValueString()

		path := t.If(
			!h.Path.IsNull() && !h.Path.IsUnknown(),
			stringPtr(h.Path.ValueString()),
			(*string)(nil),
		)

		initialDelay := int32(h.InitialDelay.ValueInt64())
		interval := int32(h.Interval.ValueInt64())
		timeout := int32(h.Timeout.ValueInt64())
		maxFailures := int32(h.MaxFailures.ValueInt64())

		result = append(result, &deploymentsv1.DeploymentHealthCheckDto{
			Port:         port,
			Protocol:     protocol,
			Path:         path,
			InitialDelay: initialDelay,
			Interval:     interval,
			Timeout:      timeout,
			MaxFailures:  maxFailures,
		})
	}

	return result
}

// buildPortDtos converts Terraform port models into protobuf DTOs.
func buildPortDtos(ports []DeploymentPortModel) []*deploymentsv1.DeploymentPortDto {
	result := make([]*deploymentsv1.DeploymentPortDto, 0, len(ports))

	for _, p := range ports {
		result = append(result, &deploymentsv1.DeploymentPortDto{
			Port:     int32(p.Port.ValueInt64()),
			Protocol: p.Protocol.ValueString(),
			Path:     p.Path.ValueString(),
		})
	}

	return result
}

// buildRegionDtos converts Terraform region models into protobuf DTOs.
func buildRegionDtos(regions []DeploymentRegionModel) []*deploymentsv1.DeploymentRegionDto {
	result := make([]*deploymentsv1.DeploymentRegionDto, 0, len(regions))

	for _, r := range regions {
		result = append(result, &deploymentsv1.DeploymentRegionDto{
			Region:      r.Region.ValueString(),
			MinReplicas: int32(r.MinReplicas.ValueInt64()),
			MaxReplicas: int32(r.MaxReplicas.ValueInt64()),
		})
	}

	return result
}

// buildVolumeDtos converts Terraform volume models into protobuf DTOs.
func buildVolumeDtos(volumes []DeploymentVolumeModel) []*deploymentsv1.DeploymentVolumeDto {
	result := make([]*deploymentsv1.DeploymentVolumeDto, 0, len(volumes))

	for _, v := range volumes {
		teamVolumeId := t.If(
			!v.TeamVolumeId.IsNull() && !v.TeamVolumeId.IsUnknown(),
			stringPtr(v.TeamVolumeId.ValueString()),
			(*string)(nil),
		)

		result = append(result, &deploymentsv1.DeploymentVolumeDto{
			MountPath:    v.MountPath.ValueString(),
			TeamVolumeId: teamVolumeId,
		})
	}

	return result
}

// buildFileDtos converts Terraform file models into protobuf file DTOs.
func buildFileDtos(files []DeploymentFileModel) []*deploymentsv1.DeploymentFileDto {
	result := make([]*deploymentsv1.DeploymentFileDto, 0, len(files))

	for _, f := range files {
		path := f.MountPath.ValueString()
		content := base64.StdEncoding.EncodeToString(
			[]byte(f.Content.ValueString()),
		)
		permissions := t.If(
			f.Permissions.IsNull() || f.Permissions.IsUnknown(),
			(*string)(nil),
			stringPtr(f.Permissions.ValueString()),
		)

		result = append(result, &deploymentsv1.DeploymentFileDto{
			MountPath:   path,
			Content:     content,
			Permissions: permissions,
		})
	}

	return result
}

func buildEnvValue(ctx context.Context, env types.Map) (*string, error) {
	if env.IsNull() || env.IsUnknown() {
		return nil, nil
	}

	var values map[string]string

	diags := env.ElementsAs(ctx, &values, false)

	if diags.HasError() {
		return nil, fmt.Errorf(
			"failed to convert Terraform env map: %s",
			diags.Errors()[0].Summary(),
		)
	}

	data, err := json.Marshal(values)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal env: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(data)

	return &encoded, nil
}

func stringPtr(value string) *string {
	return &value
}