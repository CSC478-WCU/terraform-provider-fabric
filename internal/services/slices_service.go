package services

import (
	"context"

	"github.com/csc478-wcu/terraform-provider-fabric/internal/clients/orchestrator"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/utils"
)

type SlicesService interface {
	Create(ctx context.Context, name, leaseRFC3339, graphXML string, sshKeys []string) (id, state string, slivers int, leaseFinal string, err error)
	Get(ctx context.Context, id string) (name, state string, err error)
	Delete(ctx context.Context, id string) error
}

type slicesService struct{ orc orchestrator.Client }

func NewSlicesService(orc orchestrator.Client) SlicesService { return &slicesService{orc: orc} }

func (s *slicesService) Create(ctx context.Context, name, leaseRFC3339, xml string, keys []string) (string, string, int, string, error) {
	lease, err := utils.NormalizeLease(leaseRFC3339)
	if err != nil {
		return "", "", 0, "", err
	}
	id, state, slivers, err := s.orc.CreateSlice(ctx, name, lease, xml, keys)
	return id, state, slivers, lease, err
}

func (s *slicesService) Get(ctx context.Context, id string) (string, string, error) {
	return s.orc.GetSlice(ctx, id)
}

func (s *slicesService) Delete(ctx context.Context, id string) error {
	if err := s.orc.DeleteSlice(ctx, id); err != nil {
		// Ignore “already gone” so Terraform destroy is idempotent.
		if _, ok := err.(orchestrator.NotFoundError); ok {
			return nil
		}
		return err
	}
	return nil
}
