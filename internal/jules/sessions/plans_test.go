package sessions

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/SamyRai/go-jules"
)

func TestExtractPlanSummariesIncludesFullSteps(t *testing.T) {
	created := time.Date(2026, 5, 27, 10, 0, 0, 0, time.UTC)
	plans := ExtractPlanSummaries([]jules.Activity{
		{
			ID:         "activity-plan",
			Name:       "sessions/session-1/activities/activity-plan",
			CreateTime: created,
			PlanGenerated: &jules.PlanGenerated{Plan: jules.Plan{
				ID:         "plan-1",
				CreateTime: created,
				Steps: []jules.Step{
					{ID: "step-1", Index: 1, Title: "Inspect", Description: "Read relevant files"},
					{ID: "step-2", Index: 2, Title: "Patch", Description: "Make scoped changes"},
				},
			}},
		},
	})

	if len(plans) != 1 {
		t.Fatalf("len(plans) = %d, want 1", len(plans))
	}
	plan := plans[0]
	if plan.ActivityID != "activity-plan" || plan.ActivityName == "" || plan.PlanID != "plan-1" {
		t.Fatalf("unexpected plan identity: %+v", plan)
	}
	if len(plan.Steps) != 2 || plan.Steps[1].Description != "Make scoped changes" {
		t.Fatalf("steps not fully extracted: %+v", plan.Steps)
	}
}

func TestLatestPlanSummarySelectsNewest(t *testing.T) {
	oldTime := time.Date(2026, 5, 27, 9, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(time.Hour)
	plans := ExtractPlanSummaries([]jules.Activity{
		{ID: "old", CreateTime: oldTime, PlanGenerated: &jules.PlanGenerated{Plan: jules.Plan{ID: "old-plan", CreateTime: oldTime}}},
		{ID: "new", CreateTime: newTime, PlanGenerated: &jules.PlanGenerated{Plan: jules.Plan{ID: "new-plan", CreateTime: newTime}}},
	})

	latest := LatestPlanSummary(plans)
	if latest == nil || latest.PlanID != "new-plan" {
		t.Fatalf("latest = %+v, want new-plan", latest)
	}
}

func TestExtractPlanSummariesEmpty(t *testing.T) {
	plans := ExtractPlanSummaries([]jules.Activity{{ID: "activity-1"}})
	if len(plans) != 0 {
		t.Fatalf("len(plans) = %d, want 0", len(plans))
	}
	if latest := LatestPlanSummary(plans); latest != nil {
		t.Fatalf("latest = %+v, want nil", latest)
	}
}

func TestExtractPlanSummariesMarksApprovedPlans(t *testing.T) {
	plans := ExtractPlanSummaries([]jules.Activity{
		{ID: "activity-plan", PlanGenerated: &jules.PlanGenerated{Plan: jules.Plan{ID: "plan-1"}}},
		{ID: "activity-approval", PlanApproved: &jules.PlanApproved{PlanID: "plan-1"}},
	})

	if len(plans) != 1 {
		t.Fatalf("len(plans) = %d, want 1", len(plans))
	}
	if !plans[0].Approved || plans[0].ApprovalActivityID != "activity-approval" {
		t.Fatalf("approval not detected: %+v", plans[0])
	}
}

func TestPlanSummaryJSONFieldsAreStable(t *testing.T) {
	encoded, err := json.Marshal(PlanSummary{
		ActivityID: "activity-1",
		PlanID:     "plan-1",
		Steps: []PlanStepSummary{
			{ID: "step-1", Index: 1, Title: "Inspect", Description: "Read files"},
		},
	})
	if err != nil {
		t.Fatalf("json marshal: %v", err)
	}
	got := string(encoded)
	for _, want := range []string{`"activity_id"`, `"plan_id"`, `"steps"`, `"description"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("json missing %s: %s", want, got)
		}
	}
}
