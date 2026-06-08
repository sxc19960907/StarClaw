package daemon

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCloudLifecycleAPI(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var status CloudLifecycleStatus
	getJSON(t, ts.URL+"/cloud/lifecycle", http.StatusOK, &status)
	if status.Running || status.Enabled || status.Configured {
		t.Fatalf("initial status = %#v", status)
	}

	postJSON(t, ts.URL+"/cloud/lifecycle", `{"action":"start"}`, http.StatusOK, &status)
	if !status.Running || status.StartedAt == nil {
		t.Fatalf("start status = %#v", status)
	}
	postJSON(t, ts.URL+"/cloud/lifecycle", `{"action":"restart"}`, http.StatusOK, &status)
	if status.RestartCount != 1 {
		t.Fatalf("restart status = %#v", status)
	}
	postJSON(t, ts.URL+"/cloud/lifecycle", `{"action":"stop"}`, http.StatusOK, &status)
	waitCloudLifecycleStopped(t, s.cloudLifecycle)
	getJSON(t, ts.URL+"/cloud/lifecycle", http.StatusOK, &status)
	if status.Running || status.StoppedAt == nil {
		t.Fatalf("stop status = %#v", status)
	}
	postJSON(t, ts.URL+"/cloud/lifecycle", `{"action":"bad"}`, http.StatusBadRequest, &map[string]string{})
}
