package analyzer

// ArchitectureAnalyzer analyzes project architecture patterns
type ArchitectureAnalyzer struct{}

// NewArchitectureAnalyzer creates a new architecture analyzer
func NewArchitectureAnalyzer() *ArchitectureAnalyzer {
	return &ArchitectureAnalyzer{}
}

// DetectArchitecture detects the architecture pattern
func (a *ArchitectureAnalyzer) DetectArchitecture(fileStructure map[string]int) string {
	totalFiles := 0
	for _, count := range fileStructure {
		totalFiles += count
	}

	// Heuristics for architecture detection
	if totalFiles < 10 {
		return "simple"
	} else if totalFiles < 50 {
		return "modular"
	} else if totalFiles < 200 {
		return "layered"
	}

	return "complex"
}

// CalculateComplexity calculates project complexity
func (a *ArchitectureAnalyzer) CalculateComplexity(fileStructure map[string]int, dependencies map[string]string) string {
	totalFiles := 0
	for ext, count := range fileStructure {
		if ext != "no-extension" {
			totalFiles += count
		}
	}

	depCount := len(dependencies)

	// Calculate complexity score
	score := 0

	if totalFiles > 100 {
		score += 2
	} else if totalFiles > 50 {
		score += 1
	}

	if depCount > 20 {
		score += 2
	} else if depCount > 10 {
		score += 1
	}

	switch score {
	case 0, 1:
		return "low"
	case 2, 3:
		return "medium"
	default:
		return "high"
	}
}
