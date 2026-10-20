package rotation

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRotateMediaKeyRevokesAfterLogFailure(t *testing.T) {
	revoked := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/account/keys/create":
			w.Write([]byte(`{"ok":true,"data":{"key_id":"temporary-123"}}`))
		case "/v1/account/keys/rotate/temporary-123":
			w.Write([]byte(`{"ok":true,"data":{}}`))
		case "/v1/logs/search":
			w.Write([]byte(`{"ok":false,"error":{"code":"FAIL","message":"search failed"}}`))
		case "/v1/account/keys/revoke/temporary-123":
			revoked = true
			w.Write([]byte(`{"ok":true,"data":{}}`))
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()
	_, err := RotateMediaKey(context.Background(), &Client{baseURL: server.URL, key: "test", http: server.Client()}, PipelineLoad{})
	if err == nil || !strings.Contains(err.Error(), "search failed") || !revoked {
		t.Fatalf("error = %v, revoked = %v", err, revoked)
	}
}

func TestPlanForKeepsOverlapForCreatorDelivery(t *testing.T) {
	cases := []struct {
		name  string
		load  PipelineLoad
		grace int
	}{
		{name: "creator delivery is active", load: PipelineLoad{AssetIngestions: 2, ProcessingJobs: 4, CreatorDeliveries: 1}, grace: 24},
		{name: "pipeline is idle", load: PipelineLoad{}, grace: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan := PlanFor(tc.load)
			if plan.GraceHours != tc.grace {
				t.Fatalf("grace hours = %d, want %d", plan.GraceHours, tc.grace)
			}
		})
	}
}
