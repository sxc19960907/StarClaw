package daemon

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCouncilRunLifecycle(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var emptyList struct {
		Runs []CouncilRun `json:"runs"`
	}
	getJSON(t, ts.URL+"/council", http.StatusOK, &emptyList)
	if len(emptyList.Runs) != 0 {
		t.Fatalf("expected no council runs, got %d", len(emptyList.Runs))
	}

	var created CouncilRun
	postJSON(t, ts.URL+"/council", `{"goal":"Plan Astria Agent Council","agent":"default"}`, http.StatusCreated, &created)
	if created.ID == "" {
		t.Fatal("expected council run id")
	}
	if created.Status != "completed" {
		t.Fatalf("status = %q, want completed", created.Status)
	}
	if created.Goal != "Plan Astria Agent Council" {
		t.Fatalf("goal = %q", created.Goal)
	}
	if len(created.Roles) < 2 {
		t.Fatalf("expected at least two role contributions, got %d", len(created.Roles))
	}
	if created.Synthesis == "" || !strings.Contains(created.Synthesis, "Council synthesis") {
		t.Fatalf("missing synthesis: %#v", created.Synthesis)
	}
	for _, role := range created.Roles {
		if role.Status != "completed" {
			t.Fatalf("role %s status = %q", role.Role, role.Status)
		}
		if role.Summary == "" || role.Notes == "" {
			t.Fatalf("role %s missing contribution: %#v", role.Role, role)
		}
	}

	var list struct {
		Runs []CouncilRun `json:"runs"`
	}
	getJSON(t, ts.URL+"/council", http.StatusOK, &list)
	if len(list.Runs) != 1 || list.Runs[0].ID != created.ID {
		t.Fatalf("unexpected council list: %#v", list.Runs)
	}

	var detail CouncilRun
	getJSON(t, ts.URL+"/council/"+created.ID, http.StatusOK, &detail)
	if detail.ID != created.ID || detail.Synthesis != created.Synthesis {
		t.Fatalf("unexpected council detail: %#v", detail)
	}
	getJSON(t, ts.URL+"/council/missing", http.StatusNotFound, &map[string]string{})
}

func TestCreateCouncilRunRequiresGoal(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	postJSON(t, ts.URL+"/council", `{"goal":"   "}`, http.StatusBadRequest, &map[string]string{})
}

func TestCouncilRunHandoffStartsRun(t *testing.T) {
	s := newTestServer(t, newTestServerDeps(t))
	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	var created CouncilRun
	postJSON(t, ts.URL+"/council", `{"goal":"Ship the next Astria slice"}`, http.StatusCreated, &created)

	var handoff struct {
		RunID   string           `json:"run_id"`
		Run     RunAgentResponse `json:"run"`
		Council CouncilRun       `json:"council"`
	}
	postJSON(t, ts.URL+"/council/"+created.ID+"/run", `{}`, http.StatusOK, &handoff)
	if handoff.RunID == "" || handoff.Council.ID != created.ID {
		t.Fatalf("unexpected handoff response: %+v", handoff)
	}
	if handoff.Run.SessionID == "" || len(handoff.Run.Messages) == 0 {
		t.Fatalf("expected run response after handoff: %+v", handoff.Run)
	}

	var record RunRecord
	getJSON(t, ts.URL+"/runs/"+handoff.RunID, http.StatusOK, &record)
	if record.Channel != "council_handoff" || record.Request.Source != "council:"+created.ID {
		t.Fatalf("unexpected run source/channel: %+v", record)
	}
	if record.Request.Sender != "agent-council" || !strings.Contains(record.Prompt, "Ship the next Astria slice") {
		t.Fatalf("unexpected handoff prompt metadata: %+v", record)
	}
}
