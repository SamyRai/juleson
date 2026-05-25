package jules

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeSessionName converts a bare session ID into a session resource name.
func NormalizeSessionName(session string) string {
	session = strings.TrimSpace(session)
	if strings.HasPrefix(session, "sessions/") {
		return session
	}
	return "sessions/" + session
}

// NormalizeSourceName converts a bare source ID into a source resource name.
func NormalizeSourceName(source string) string {
	source = strings.TrimSpace(source)
	if strings.HasPrefix(source, "sources/") {
		return source
	}
	return "sources/" + source
}

func sessionPath(session string) (string, error) {
	name := NormalizeSessionName(session)
	id, ok := strings.CutPrefix(name, "sessions/")
	if !ok || id == "" || strings.Contains(id, "/") {
		return "", fmt.Errorf("invalid session resource name %q", session)
	}
	return "sessions/" + url.PathEscape(id), nil
}

func sourcePath(source string) (string, error) {
	name := NormalizeSourceName(source)
	id, ok := strings.CutPrefix(name, "sources/")
	if !ok || id == "" {
		return "", fmt.Errorf("invalid source resource name %q", source)
	}
	parts := strings.Split(id, "/")
	for _, part := range parts {
		if part == "" {
			return "", fmt.Errorf("invalid source resource name %q", source)
		}
	}
	return "sources/" + strings.Join(parts, "/"), nil
}

func activityPath(session, activity string) (string, error) {
	activity = strings.TrimSpace(activity)
	if strings.HasPrefix(activity, "sessions/") {
		parts := strings.Split(activity, "/")
		if len(parts) != 4 || parts[0] != "sessions" || parts[2] != "activities" || parts[1] == "" || parts[3] == "" {
			return "", fmt.Errorf("invalid activity resource name %q", activity)
		}
		return "sessions/" + url.PathEscape(parts[1]) + "/activities/" + url.PathEscape(parts[3]), nil
	}

	activityID := activity
	if id, ok := strings.CutPrefix(activity, "activities/"); ok {
		activityID = id
	}
	if activityID == "" || strings.Contains(activityID, "/") {
		return "", fmt.Errorf("invalid activity ID %q", activity)
	}

	sessionResourcePath, err := sessionPath(session)
	if err != nil {
		return "", err
	}
	return sessionResourcePath + "/activities/" + url.PathEscape(activityID), nil
}
