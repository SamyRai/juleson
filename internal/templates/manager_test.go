package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		manager, err := NewManager("../../templates/builtin", "", false)
		require.NoError(t, err)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.registry)
		assert.True(t, len(manager.registry.Templates) > 0)
	})
}

func TestListTemplates(t *testing.T) {
	manager, err := NewManager("../../templates/builtin", "", false)
	require.NoError(t, err)

	templates := manager.ListTemplates()
	assert.True(t, len(templates) > 0)
}

func TestListTemplatesByCategory(t *testing.T) {
	manager, err := NewManager("../../templates/builtin", "", false)
	require.NoError(t, err)

	templates := manager.ListTemplatesByCategory("testing")
	assert.True(t, len(templates) > 0)
	for _, tmpl := range templates {
		assert.Equal(t, "testing", tmpl.Category)
	}
}

func TestSearchTemplates(t *testing.T) {
	manager, err := NewManager("../../templates/builtin", "", false)
	require.NoError(t, err)

	results := manager.SearchTemplates("test")
	assert.True(t, len(results) > 0)
}

func TestLoadTemplate(t *testing.T) {
	manager, err := NewManager("../../templates/builtin", "", false)
	require.NoError(t, err)

	template, err := manager.LoadTemplate("test-generation")
	require.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, "test-generation", template.Metadata.Name)
}

func TestValidateTemplate(t *testing.T) {
	manager, err := NewManager("../../templates/builtin", "", false)
	require.NoError(t, err)

	template, err := manager.LoadTemplate("test-generation")
	require.NoError(t, err)

	err = manager.ValidateTemplate(template)
	assert.NoError(t, err)
}

func TestCreateTemplate(t *testing.T) {
	manager, err := NewManager("../../templates/builtin", "", false)
	require.NoError(t, err)

	template, err := manager.CreateTemplate("test-template", "testing", "A test template")
	require.NoError(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, "test-template", template.Metadata.Name)
	assert.Equal(t, "testing", template.Metadata.Category)
}
