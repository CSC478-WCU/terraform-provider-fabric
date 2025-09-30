package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	openapi "github.com/csc478-wcu/fabric-orchestrator-go-client"
	"github.com/csc478-wcu/terraform-provider-fabric/internal/fabric_client_wrapper"
)

// FabricClient defines the high-level operations required by the provider components.
type FabricClient interface {
	CreateSlice(ctx context.Context, req SliceCreateRequest) (*SliceCreateResult, error)
	GetSlice(ctx context.Context, sliceID string) (*SliceDetails, error)
	DeleteSlice(ctx context.Context, sliceID string) error
	ListResources(ctx context.Context, opts ResourceListOptions) ([]ResourceModel, error)
}

type fabricClient struct {
	api   *fabric_client_wrapper.FabricClientWrapper
	token string
}

// SliceCreateRequest captures the parameters necessary to create a slice.
type SliceCreateRequest struct {
	Name         string
	LeaseEndTime string
	GraphModel   string
	SSHKeys      []string
}

// SliceCreateResult summarizes the slice right after creation.
type SliceCreateResult struct {
	ID          string
	State       string
	SliverCount int
}

// SliceDetails captures essential slice information for state refresh.
type SliceDetails struct {
	ID    string
	Name  string
	State string
}

// ResourceListOptions controls filtering options for resource discovery.
type ResourceListOptions struct {
	Level    *int32
	Includes []string
	Excludes []string
}

// ResourceModel represents a simplified view of an available resource.
type ResourceModel struct {
	Model string
}

// NewFabricClient constructs a new FabricClient instance.
func NewFabricClient(api *fabric_client_wrapper.FabricClientWrapper, token string) FabricClient {
	return &fabricClient{
		api:   api,
		token: token,
	}
}

// normalizeLeaseEndTime ensures the lease time is always in FABRIC's expected format.
// Accepts RFC3339 (e.g., 2025-09-30T12:00:00Z) or already-correct format.
func normalizeLeaseEndTime(input string) (string, error) {
	if input == "" {
		return "", nil
	}
	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, input); err == nil {
		return t.Format("2006-01-02 15:04:05 -0700"), nil
	}
	// Try parsing as FABRIC's format already
	if _, err := time.Parse("2006-01-02 15:04:05 -0700", input); err == nil {
		return input, nil
	}
	return "", fmt.Errorf("invalid lease_end_time format: %s", input)
}

// --- Slice Operations ---

func (c *fabricClient) CreateSlice(ctx context.Context, req SliceCreateRequest) (*SliceCreateResult, error) {
	apiCtx := context.WithValue(ctx, openapi.ContextAccessToken, c.token)

	// Normalize lease time
	leaseEnd, err := normalizeLeaseEndTime(req.LeaseEndTime)
	if err != nil {
		return nil, err
	}

	// Build request body (JSON)
	body := openapi.SlicesPost{}
	body.SetGraphModel(req.GraphModel)
	body.SetSshKeys(req.SSHKeys)

	call := c.api.SlicesAPI.
		SlicesCreatesPost(apiCtx).
		Name(req.Name).
		LeaseEndTime(leaseEnd).
		SlicesPost(body)

	res, httpResp, err := call.Execute()
	if err != nil {
		b, _ := io.ReadAll(httpResp.Body)
		return nil, fmt.Errorf("create slice failed: %w\nraw: %s", err, string(b))
	}

	data := res.GetData()
	if len(data) == 0 {
		return nil, errors.New("slice create returned no sliver information")
	}

	first := data[0]
	return &SliceCreateResult{
		ID:          first.GetSliceId(),
		State:       first.GetState(),
		SliverCount: len(data),
	}, nil
}

func (c *fabricClient) GetSlice(ctx context.Context, sliceID string) (*SliceDetails, error) {
	apiCtx := context.WithValue(ctx, openapi.ContextAccessToken, c.token)

	result, _, err := c.api.SlicesAPI.SlicesSliceIdGet(apiCtx, sliceID).Execute()
	if err != nil {
		return nil, err
	}

	data := result.GetData()
	if len(data) == 0 {
		return nil, fmt.Errorf("slice %s not found", sliceID)
	}

	slice := data[0]
	return &SliceDetails{
		ID:    slice.GetSliceId(),
		Name:  slice.GetName(),
		State: slice.GetState(),
	}, nil
}

func (c *fabricClient) DeleteSlice(ctx context.Context, sliceID string) error {
	apiCtx := context.WithValue(ctx, openapi.ContextAccessToken, c.token)
	_, _, err := c.api.SlicesAPI.SlicesDeleteSliceIdDelete(apiCtx, sliceID).Execute()
	return err
}

func (c *fabricClient) ListResources(ctx context.Context, opts ResourceListOptions) ([]ResourceModel, error) {
	apiCtx := context.WithValue(ctx, openapi.ContextAccessToken, c.token)

	call := c.api.ResourcesAPI.ResourcesGet(apiCtx)
	if opts.Level != nil {
		call = call.Level(*opts.Level)
	}
	for _, include := range opts.Includes {
		call = call.Includes(include)
	}
	for _, exclude := range opts.Excludes {
		call = call.Excludes(exclude)
	}

	result, _, err := call.Execute()
	if err != nil {
		return nil, err
	}

	data := result.GetData()
	resources := make([]ResourceModel, 0, len(data))
	for _, res := range data {
		resources = append(resources, ResourceModel{Model: res.GetModel()})
	}

	return resources, nil
}
