package services

import (
	"context"

	"github.com/csc478-wcu/terraform-provider-fabric/internal/clients/orchestrator"
)

type ResourcesService interface {
	List(ctx context.Context, level *int32, includes, excludes []string) ([]string, error)
}

type resourcesService struct{ orc orchestrator.Client }

func NewResourcesService(orc orchestrator.Client) ResourcesService {
	return &resourcesService{orc: orc}
}

func (s *resourcesService) List(ctx context.Context, level *int32, includes, excludes []string) ([]string, error) {
	return s.orc.ListResources(ctx, level, includes, excludes)
}
