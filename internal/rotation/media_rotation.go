package rotation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type PipelineLoad struct {
	AssetIngestions   int
	ProcessingJobs    int
	CreatorDeliveries int
}

type RotationPlan struct {
	GraceHours int
	Reason     string
}

type RotationResult struct {
	TemporaryKeyID string
	Plan           RotationPlan
	LogSnapshot    json.RawMessage
}

func PlanFor(load PipelineLoad) RotationPlan {
	if load.AssetIngestions+load.ProcessingJobs+load.CreatorDeliveries > 0 {
		return RotationPlan{GraceHours: 24, Reason: "active media pipeline stages retain the old key during deployment"}
	}
	return RotationPlan{GraceHours: 1, Reason: "no active media pipeline stages"}
}

func RotateMediaKey(ctx context.Context, client *Client, load PipelineLoad) (result RotationResult, err error) {
	plan := PlanFor(load)
	idempotencyKey := fmt.Sprintf("media-rotation-%d", time.Now().UnixNano())
	created, err := client.Do(ctx, http.MethodPost, "/v1/account/keys/create", map[string]any{
		"name":            "media-rotation-temporary",
		"idempotency_key": idempotencyKey,
	})
	if err != nil {
		return RotationResult{}, err
	}

	var key struct {
		ID string `json:"key_id"`
	}
	if err := json.Unmarshal(created, &key); err != nil || key.ID == "" {
		return RotationResult{}, fmt.Errorf("created key response did not contain an id")
	}
	defer func() {
		// Cleanup must still run if the caller's context was canceled.
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if _, revokeErr := client.Do(cleanupCtx, http.MethodDelete, "/v1/account/keys/revoke/"+key.ID, nil); revokeErr != nil {
			err = errors.Join(err, fmt.Errorf("revoke temporary key %s: %w", key.ID, revokeErr))
		}
	}()

	_, err = client.Do(ctx, http.MethodPost, "/v1/account/keys/rotate/"+key.ID, map[string]any{
		"grace_hours":     plan.GraceHours,
		"idempotency_key": idempotencyKey,
	})
	if err != nil {
		return RotationResult{}, err
	}

	logs, err := client.Do(ctx, http.MethodGet, "/v1/logs/search", nil)
	if err != nil {
		return RotationResult{}, err
	}
	return RotationResult{TemporaryKeyID: key.ID, Plan: plan, LogSnapshot: logs}, nil
}

func (r RotationResult) Summary() string {
	return fmt.Sprintf("temporary key %s rotated with %d-hour overlap; deployment log snapshot: %s", r.TemporaryKeyID, r.Plan.GraceHours, strings.TrimSpace(string(r.LogSnapshot)))
}
