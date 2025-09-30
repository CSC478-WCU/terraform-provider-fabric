package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	openapi "github.com/csc478-wcu/fabric-orchestrator-go-client"
)

const defaultGraphFormat = "GRAPHML" // allowed: GRAPHML, JSON_NODELINK, CYTOSCAPE, NONE

type NotFoundError struct{ msg string }

func (e NotFoundError) Error() string { return e.msg }

type Config struct {
	Endpoint string
	Token    string
}

type Client interface {
	CreateSlice(ctx context.Context, name, leaseEnd, model string, sshKeys []string) (sliceID, state string, slivers int, err error)
	GetSlice(ctx context.Context, sliceID string) (name, state string, err error)
	DeleteSlice(ctx context.Context, sliceID string) error
	ListResources(ctx context.Context, level *int32, includes, excludes []string) ([]string, error)
}

type client struct {
	api   *openapi.APIClient
	token string
}

func New(cfg Config) Client {
	conf := openapi.NewConfiguration()
	if cfg.Endpoint != "" {
		conf.Servers = openapi.ServerConfigurations{{URL: cfg.Endpoint}}
	}
	// ensure our content-type fix transport is used
	conf.HTTPClient = withContentTypeFix(conf.HTTPClient)

	return &client{
		api:   openapi.NewAPIClient(conf),
		token: cfg.Token,
	}
}

func (c *client) CreateSlice(ctx context.Context, name, leaseEnd, model string, sshKeys []string) (string, string, int, error) {
	apiCtx := context.WithValue(ctx, openapi.ContextAccessToken, c.token)

	body := openapi.SlicesPost{}
	body.SetGraphModel(model)
	body.SetSshKeys(sshKeys)

	res, httpResp, err := c.api.SlicesAPI.
		SlicesCreatesPost(apiCtx).
		Name(name).
		LeaseEndTime(leaseEnd).
		SlicesPost(body).
		Execute()

	// Happy path: the SDK decoded a known type
	if err == nil {
		data := res.GetData()
		if len(data) == 0 {
			return "", "", 0, errors.New("create slice: empty response data")
		}
		first := data[0]
		return first.GetSliceId(), first.GetState(), len(data), nil
	}

	// Fallback path: Some FABRIC deployments return a 200 with an "untyped" JSON payload
	// the generated client doesn't recognize. If we got an HTTP response, inspect it.
	if httpResp != nil {
		raw, _ := io.ReadAll(httpResp.Body)
		_ = httpResp.Body.Close()

		// If it's not 200, just bubble up the error with raw for debugging.
		if httpResp.StatusCode != 200 {
			return "", "", 0, fmt.Errorf("create slice: %w raw=%s", err, string(raw))
		}

		// Try to parse the untyped JSON:
		// {
		//   "data": [ { "slice_id": "...", "state": "...", ... }, ... ],
		//   "size": 2,
		//   "status": 200,
		//   "type": "slivers"
		// }
		type sliverItem struct {
			SliceID string `json:"slice_id"`
			State   string `json:"state"`
		}
		var fallback struct {
			Data   []sliverItem `json:"data"`
			Size   int          `json:"size"`
			Status int          `json:"status"`
			Type   string       `json:"type"`
		}

		if json.Unmarshal(raw, &fallback) == nil && len(fallback.Data) > 0 {
			first := fallback.Data[0]
			// Some responses have per-sliver state; return that, and sliver count as size/data len.
			return first.SliceID, first.State, len(fallback.Data), nil
		}

		// JSON parse failed — surface original error and raw body
		return "", "", 0, fmt.Errorf("create slice: %w raw=%s", err, string(raw))
	}

	// No httpResp available - return the SDK error
	return "", "", 0, fmt.Errorf("create slice: %w", err)
}

func (c *client) GetSlice(ctx context.Context, sliceID string) (string, string, error) {
	apiCtx := context.WithValue(ctx, openapi.ContextAccessToken, c.token)

	res, httpResp, err := c.api.SlicesAPI.
		SlicesSliceIdGet(apiCtx, sliceID).
		GraphFormat(defaultGraphFormat). // "GRAPHML"
		Execute()

	// Happy path: SDK decoded a known type
	if err == nil {
		data := res.GetData()
		if len(data) == 0 {
			return "", "", NotFoundError{msg: fmt.Sprintf("slice %s not found (empty data)", sliceID)}
		}
		s := data[0]
		return s.GetName(), s.GetState(), nil
	}

	// Fallback path: SDK errored but we have an HTTP response body
	if httpResp != nil {
		raw, _ := io.ReadAll(httpResp.Body)
		_ = httpResp.Body.Close()

		// Not found?
		if httpResp.StatusCode == 404 {
			return "", "", NotFoundError{msg: fmt.Sprintf("slice %s not found: %s", sliceID, string(raw))}
		}

		// Some deployments return a 200 with a different shape the SDK can't decode.
		if httpResp.StatusCode == 200 {
			var fb struct {
				Data []struct {
					Name  string `json:"name"`
					State string `json:"state"`
				} `json:"data"`
			}
			if json.Unmarshal(raw, &fb) == nil && len(fb.Data) > 0 {
				return fb.Data[0].Name, fb.Data[0].State, nil
			}
			// If the shape changes again, surface the raw so we can tweak quickly.
			return "", "", fmt.Errorf("get slice: unrecognized 200 response shape: %s", string(raw))
		}

		// Other status codes: return original error plus raw for debugging
		return "", "", fmt.Errorf("get slice: %w raw=%s", err, string(raw))
	}

	// No httpResp available: return the SDK error
	return "", "", fmt.Errorf("get slice: %w", err)
}

func (c *client) DeleteSlice(ctx context.Context, sliceID string) error {
	apiCtx := context.WithValue(ctx, openapi.ContextAccessToken, c.token)

	_, httpResp, err := c.api.SlicesAPI.
		SlicesDeleteSliceIdDelete(apiCtx, sliceID).
		Execute()

	// Happy path: SDK decoded fine.
	if err == nil {
		return nil
	}

	// Fallback: inspect raw HTTP response.
	if httpResp != nil {
		raw, _ := io.ReadAll(httpResp.Body)
		_ = httpResp.Body.Close()

		switch httpResp.StatusCode {
		case 200, 202, 204:
			// Some deployments return an untyped 200/202/204; treat as success.
			return nil
		case 404:
			// Not found — bubble a typed error so the service layer can ignore it.
			return NotFoundError{msg: fmt.Sprintf("slice %s not found: %s", sliceID, string(raw))}
		default:
			return fmt.Errorf("delete slice: %w raw=%s", err, string(raw))
		}
	}

	// No httpResp available: return the SDK error.
	return fmt.Errorf("delete slice: %w", err)
}

func (c *client) ListResources(ctx context.Context, level *int32, includes, excludes []string) ([]string, error) {
	apiCtx := context.WithValue(ctx, openapi.ContextAccessToken, c.token)
	call := c.api.ResourcesAPI.ResourcesGet(apiCtx)
	if level != nil {
		call = call.Level(*level)
	}
	for _, inc := range includes {
		call = call.Includes(inc)
	}
	for _, exc := range excludes {
		call = call.Excludes(exc)
	}
	res, _, err := call.Execute()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(res.GetData()))
	for _, d := range res.GetData() {
		out = append(out, d.GetModel())
	}
	return out, nil
}
