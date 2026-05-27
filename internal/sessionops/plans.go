package sessionops

import (
	"sort"
	"time"

	"github.com/SamyRai/go-jules"
)

type PlanStepSummary struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Index       int    `json:"index,omitempty"`
}

type PlanSummary struct {
	ActivityID         string            `json:"activity_id"`
	ActivityName       string            `json:"activity_name,omitempty"`
	PlanID             string            `json:"plan_id"`
	ApprovalActivityID string            `json:"approval_activity_id,omitempty"`
	Steps              []PlanStepSummary `json:"steps"`
	ActivityCreateTime time.Time         `json:"activity_create_time,omitempty"`
	PlanCreateTime     time.Time         `json:"plan_create_time,omitempty"`
	Approved           bool              `json:"approved"`
}

func ExtractPlanSummaries(activities []jules.Activity) []PlanSummary {
	approvedPlans := approvedPlanActivities(activities)
	var plans []PlanSummary
	for i := range activities {
		activity := &activities[i]
		if activity.PlanGenerated == nil {
			continue
		}
		plan := activity.PlanGenerated.Plan
		summary := PlanSummary{
			ActivityID:         activity.ID,
			ActivityName:       activity.Name,
			ActivityCreateTime: activity.CreateTime,
			PlanID:             plan.ID,
			PlanCreateTime:     plan.CreateTime,
			Steps:              make([]PlanStepSummary, 0, len(plan.Steps)),
		}
		if approvalActivityID, ok := approvedPlans[plan.ID]; ok {
			summary.Approved = true
			summary.ApprovalActivityID = approvalActivityID
		}
		for _, step := range plan.Steps {
			summary.Steps = append(summary.Steps, PlanStepSummary{
				ID:          step.ID,
				Index:       step.Index,
				Title:       step.Title,
				Description: step.Description,
			})
		}
		plans = append(plans, summary)
	}
	sortPlanSummaries(plans)
	return plans
}

func LatestPlanSummary(plans []PlanSummary) *PlanSummary {
	if len(plans) == 0 {
		return nil
	}
	sorted := append([]PlanSummary(nil), plans...)
	sortPlanSummaries(sorted)
	latest := sorted[len(sorted)-1]
	return &latest
}

func approvedPlanActivities(activities []jules.Activity) map[string]string {
	approved := make(map[string]string)
	for i := range activities {
		activity := &activities[i]
		if activity.PlanApproved == nil || activity.PlanApproved.PlanID == "" {
			continue
		}
		approved[activity.PlanApproved.PlanID] = activity.ID
	}
	return approved
}

func sortPlanSummaries(plans []PlanSummary) {
	sort.SliceStable(plans, func(i, j int) bool {
		left := plans[i]
		right := plans[j]
		leftTime := left.PlanCreateTime
		if leftTime.IsZero() {
			leftTime = left.ActivityCreateTime
		}
		rightTime := right.PlanCreateTime
		if rightTime.IsZero() {
			rightTime = right.ActivityCreateTime
		}
		if leftTime.Equal(rightTime) {
			return left.ActivityID < right.ActivityID
		}
		return leftTime.Before(rightTime)
	})
}
