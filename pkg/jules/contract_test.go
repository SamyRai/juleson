package jules

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContractFixtures(t *testing.T) {
	t.Run("session", func(t *testing.T) {
		var session Session
		readFixture(t, "session.json", &session)

		assert.Equal(t, "sessions/1234567", session.Name)
		assert.Equal(t, SessionStateAwaitingPlanApproval, session.State)
		assert.True(t, session.State.NeedsUserAction())
		assert.Equal(t, AutomationModeAutoCreatePR, session.AutomationMode)
		assert.Equal(t, testTime("2024-01-15T10:30:00Z"), session.CreateTime)
		require.NotNil(t, session.SourceContext)
		assert.Equal(t, "sources/github/myorg/myrepo", session.SourceContext.Source)
		require.Len(t, session.Outputs, 1)
		require.NotNil(t, session.Outputs[0].PullRequest)
	})

	t.Run("repoless session", func(t *testing.T) {
		var session Session
		readFixture(t, "repoless_session.json", &session)

		assert.Nil(t, session.SourceContext)
		assert.Equal(t, SessionStateCompleted, session.State)
		assert.True(t, session.State.IsTerminal())
		assert.True(t, session.State.IsSuccessful())
	})

	t.Run("source", func(t *testing.T) {
		var source Source
		readFixture(t, "source.json", &source)

		assert.Equal(t, "sources/github/myorg/myrepo", source.Name)
		require.NotNil(t, source.GithubRepo)
		assert.True(t, source.GithubRepo.IsPrivate)
		require.NotNil(t, source.GithubRepo.DefaultBranch)
		assert.Equal(t, "main", source.GithubRepo.DefaultBranch.DisplayName)
	})

	t.Run("activity artifacts", func(t *testing.T) {
		var activity Activity
		readFixture(t, "activity.json", &activity)

		assert.Equal(t, ActivityOriginatorAgent, activity.Originator)
		assert.Equal(t, testTime("2024-01-15T11:00:00Z"), activity.CreateTime)
		require.NotNil(t, activity.PlanGenerated)
		assert.Equal(t, testTime("2024-01-15T10:31:00Z"), activity.PlanGenerated.Plan.CreateTime)
		require.Len(t, activity.Artifacts, 3)

		patchContent, err := ArtifactContent(activity.Artifacts[0])
		require.NoError(t, err)
		assert.Contains(t, string(patchContent), "diff --git")

		bashContent, err := ArtifactContent(activity.Artifacts[1])
		require.NoError(t, err)
		assert.Contains(t, string(bashContent), "go test ./...")

		mediaContent, err := ArtifactContent(activity.Artifacts[2])
		require.NoError(t, err)
		assert.Equal(t, []byte("hello"), mediaContent)
	})
}

func TestTimestampRoundTrip(t *testing.T) {
	created := testTime("2026-01-17T00:03:53.13724Z")
	session := Session{
		Name:       "sessions/1",
		ID:         "1",
		State:      SessionStateQueued,
		CreateTime: created,
		UpdateTime: created.Add(time.Minute),
	}

	data, err := json.Marshal(session)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"createTime":"2026-01-17T00:03:53.13724Z"`)

	var decoded Session
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, session.CreateTime, decoded.CreateTime)
	assert.Equal(t, session.UpdateTime, decoded.UpdateTime)
}

func readFixture(t *testing.T, name string, target any) {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, target))
}
