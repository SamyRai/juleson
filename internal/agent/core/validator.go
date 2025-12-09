package core

import (
	"fmt"
	"strings"

	"github.com/SamyRai/juleson/internal/agent"
)

// ConstraintValidator validates that agent actions respect goal constraints
type ConstraintValidator struct {
	constraints []Constraint
}

// Constraint represents a limitation on what the agent can do
type Constraint struct {
	Type        ConstraintType
	Description string
	Validator   func(action agent.Action) error
}

// ConstraintType categorizes constraints
type ConstraintType string

const (
	ConstraintTypeAPI          ConstraintType = "API"          // No API changes
	ConstraintTypeArchitecture ConstraintType = "ARCHITECTURE" // Must follow pattern
	ConstraintTypeSecurity     ConstraintType = "SECURITY"     // Security requirements
	ConstraintTypePerformance  ConstraintType = "PERFORMANCE"  // Performance limits
	ConstraintTypeTesting      ConstraintType = "TESTING"      // Testing requirements
)

// NewConstraintValidator creates a validator for goal constraints
func NewConstraintValidator(goalConstraints []string) *ConstraintValidator {
	validator := &ConstraintValidator{
		constraints: make([]Constraint, 0),
	}

	// Parse textual constraints into structured constraints
	for _, constraintText := range goalConstraints {
		constraint := parseConstraint(constraintText)
		if constraint != nil {
			validator.constraints = append(validator.constraints, *constraint)
		}
	}

	return validator
}

// Validate checks if an action violates any constraints
func (cv *ConstraintValidator) Validate(action agent.Action) error {
	for _, constraint := range cv.constraints {
		if err := constraint.Validator(action); err != nil {
			return fmt.Errorf("constraint violation (%s): %w", constraint.Description, err)
		}
	}
	return nil
}

// ValidateChanges validates that changes respect constraints
func (cv *ConstraintValidator) ValidateChanges(changes []agent.Change) error {
	for _, change := range changes {
		// Check each constraint against the change
		for _, constraint := range cv.constraints {
			// Create an action from the change for validation
			action := agent.Action{
				Type: mapChangeToActionType(change.Type),
				Parameters: map[string]interface{}{
					"file_path": change.FilePath,
					"patch":     change.Patch,
				},
			}

			if err := constraint.Validator(action); err != nil {
				return fmt.Errorf("change to %s violates constraint: %w", change.FilePath, err)
			}
		}
	}
	return nil
}

// parseConstraint converts textual constraints into structured validation functions
func parseConstraint(text string) *Constraint {
	lower := strings.ToLower(text)

	// "Don't change public APIs"
	if strings.Contains(lower, "public api") || strings.Contains(lower, "api") {
		return &Constraint{
			Type:        ConstraintTypeAPI,
			Description: text,
			Validator: func(action agent.Action) error {
				// Check if action modifies public API
				filePathVal, ok := action.Parameters["file_path"]
				if !ok {
					return nil // No file path, skip validation
				}
				filePath, ok := filePathVal.(string)
				if !ok {
					return nil // Not a string, skip validation
				}

				patchVal, ok := action.Parameters["patch"]
				if !ok {
					return nil // No patch, skip validation
				}
				patch, ok := patchVal.(string)
				if !ok {
					return nil // Not a string, skip validation
				}

				// Simple heuristic: check for exported function/method changes
				if strings.Contains(patch, "func ") || strings.Contains(patch, "type ") {
					// Check if it's exported (starts with capital letter)
					if containsExportedSymbol(patch) {
						return fmt.Errorf("modifying public API: %s", filePath)
					}
				}
				return nil
			},
		}
	}

	// "Maintain backward compatibility"
	if strings.Contains(lower, "backward compat") {
		return &Constraint{
			Type:        ConstraintTypeAPI,
			Description: text,
			Validator: func(action agent.Action) error {
				patchVal, ok := action.Parameters["patch"]
				if !ok {
					return nil // No patch, skip validation
				}
				patch, ok := patchVal.(string)
				if !ok {
					return nil // Not a string, skip validation
				}

				// Check for breaking changes
				if strings.Contains(patch, "- func") || strings.Contains(patch, "- type") {
					return fmt.Errorf("removing functions/types breaks compatibility")
				}
				return nil
			},
		}
	}

	// "Must include tests"
	if strings.Contains(lower, "test") || strings.Contains(lower, "coverage") {
		return &Constraint{
			Type:        ConstraintTypeTesting,
			Description: text,
			Validator: func(action agent.Action) error {
				filePathVal, ok := action.Parameters["file_path"]
				if !ok {
					return nil // No file path, skip validation
				}
				filePath, ok := filePathVal.(string)
				if !ok {
					return nil // Not a string, skip validation
				}

				// If adding/modifying code, ensure tests are included
				if action.Type == agent.ActionTypeCreate || action.Type == agent.ActionTypeModify {
					if strings.HasSuffix(filePath, ".go") && !strings.HasSuffix(filePath, "_test.go") {
						// Should have corresponding test file
						// This is a simplified check
						return nil // Would need to check if test file exists
					}
				}
				return nil
			},
		}
	}

	// "No external dependencies"
	if strings.Contains(lower, "no") && strings.Contains(lower, "depend") {
		return &Constraint{
			Type:        ConstraintTypeArchitecture,
			Description: text,
			Validator: func(action agent.Action) error {
				patchVal, ok := action.Parameters["patch"]
				if !ok {
					return nil // No patch, skip validation
				}
				patch, ok := patchVal.(string)
				if !ok {
					return nil // Not a string, skip validation
				}

				// Check for import statements
				if strings.Contains(patch, "+ import") || strings.Contains(patch, "+import") {
					return fmt.Errorf("adding new dependencies is not allowed")
				}
				return nil
			},
		}
	}

	// Generic constraint - just track it
	return &Constraint{
		Type:        "CUSTOM",
		Description: text,
		Validator: func(action agent.Action) error {
			// No automatic validation for custom constraints
			// Will need human review
			return nil
		},
	}
}

// containsExportedSymbol checks if patch contains exported (public) symbols
func containsExportedSymbol(patch string) bool {
	lines := strings.Split(patch, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "+") {
			line = strings.TrimPrefix(strings.TrimSpace(line), "+")

			// Check for exported function
			if strings.HasPrefix(line, "func ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					funcName := parts[1]
					// Remove receiver if present
					if strings.HasPrefix(funcName, "(") {
						if len(parts) >= 4 {
							funcName = parts[3]
						}
					}
					// Remove parameters
					if idx := strings.Index(funcName, "("); idx > 0 {
						funcName = funcName[:idx]
					}
					// Check if exported (starts with uppercase)
					if len(funcName) > 0 && funcName[0] >= 'A' && funcName[0] <= 'Z' {
						return true
					}
				}
			}

			// Check for exported type
			if strings.HasPrefix(line, "type ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					typeName := parts[1]
					if len(typeName) > 0 && typeName[0] >= 'A' && typeName[0] <= 'Z' {
						return true
					}
				}
			}
		}
	}
	return false
}

// mapChangeToActionType maps ChangeType to ActionType
func mapChangeToActionType(changeType agent.ChangeType) agent.ActionType {
	switch changeType {
	case agent.ChangeTypeAdd:
		return agent.ActionTypeCreate
	case agent.ChangeTypeModify:
		return agent.ActionTypeModify
	case agent.ChangeTypeDelete:
		return agent.ActionTypeDelete
	default:
		return agent.ActionTypeModify
	}
}
