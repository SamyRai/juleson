package templates

// Store defines the interface for template storage and retrieval.
type Store interface {
	LoadRegistry() (*Registry, error)
	LoadTemplate(path string) (*Template, error)
}
