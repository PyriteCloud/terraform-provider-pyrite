package provider

import deploymentsv1 "github.com/PyriteCloud/client-go/lib/gen/pyrite/v1/services/v1/deployments/v1"

// buildHealthCheckDtos converts Terraform health check models into protobuf DTOs.
func buildHealthCheckDtos(
	checks []DeploymentHealthCheckModel,
) []*deploymentsv1.DeploymentHealthCheckDto {
	result := make([]*deploymentsv1.DeploymentHealthCheckDto, 0, len(checks))

	for _, h := range checks {
		port := int32(h.Port.ValueInt64())
		protocol := h.Protocol.ValueString()
		path := h.Path.ValueString()

		initialDelay := int32(h.InitialDelay.ValueInt64())
		interval := int32(h.Interval.ValueInt64())
		timeout := int32(h.Timeout.ValueInt64())
		maxFailures := int32(h.MaxFailures.ValueInt64())

		result = append(result, &deploymentsv1.DeploymentHealthCheckDto{
			Port:         port,
			Protocol:     protocol,
			Path:         &path,
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
		teamVolumeId := v.TeamVolumeId.ValueString()

		result = append(result, &deploymentsv1.DeploymentVolumeDto{
			MountPath:    v.MountPath.ValueString(),
			TeamVolumeId: &teamVolumeId,
		})
	}

	return result
}

// buildFileDtos converts Terraform file models into protobuf file DTOs.
func buildFileDtos(files []DeploymentFileModel) []*deploymentsv1.DeploymentFileDto {
	result := make([]*deploymentsv1.DeploymentFileDto, 0, len(files))

	for _, f := range files {
		path := f.Path.ValueString()
		content := f.Content.ValueString()
		permissions := f.Permissions.ValueString()

		result = append(result, &deploymentsv1.DeploymentFileDto{
			MountPath:   path,
			Content:     content,
			Permissions: &permissions,
		})
	}

	return result
}
