# Jules Automation Templates

This directory contains YAML-based templates for Jules automation tasks. Templates use a structured format inspired by Y2Q2 (YAML-to-Query) patterns for easy task definition and context management.

## 📁 **Directory Structure**

```
templates/
├── builtin/           # Built-in templates (shipped with the tool)
│   ├── reorganization/
│   ├── refactoring/
│   ├── testing/
│   └── documentation/
├── custom/            # User-defined custom templates
└── registry/          # Template registry and metadata
```

## 🎯 **Template Format**

Templates use YAML with the following structure:

```yaml
# Template metadata
metadata:
  name: "project-reorganization"
  version: "1.0.0"
  description: "Reorganize project structure using modular architecture"
  author: "Jules Automation"
  tags: ["reorganization", "architecture", "modular"]
  category: "reorganization"

# Template configuration
config:
  strategy: "modular"
  max_concurrent_tasks: 3
  timeout: "600s"
  requires_approval: true

# Context extraction rules
context:
  project_analysis:
    - "analyze_directory_structure"
    - "identify_dependencies"
    - "analyze_code_patterns"
    - "detect_architectural_issues"
  
  file_patterns:
    include:
      - "**/*.go"
      - "**/*.py"
      - "**/*.js"
      - "**/*.ts"
      - "**/*.java"
    exclude:
      - "**/node_modules/**"
      - "**/vendor/**"
      - "**/.git/**"

# Task definition
tasks:
  - name: "analyze_current_structure"
    type: "analysis"
    description: "Analyze current project structure and identify issues"
    jules_prompt: |
      Analyze the current project structure in {{.ProjectPath}} and identify:
      1. Architectural issues and inconsistencies
      2. Dependency problems
      3. Code organization problems
      4. Potential improvements
      
      Focus on: {{.FocusAreas}}
      Exclude: {{.ExcludePatterns}}
    
    context_vars:
      ProjectPath: "{{.ProjectPath}}"
      FocusAreas: "{{.FocusAreas}}"
      ExcludePatterns: "{{.ExcludePatterns}}"

  - name: "create_reorganization_plan"
    type: "planning"
    description: "Create detailed reorganization plan"
    depends_on: ["analyze_current_structure"]
    jules_prompt: |
      Based on the analysis of {{.ProjectPath}}, create a detailed reorganization plan:
      
      1. **Target Architecture**: {{.TargetArchitecture}}
      2. **Migration Strategy**: Step-by-step approach
      3. **Risk Assessment**: Potential issues and mitigation
      4. **Timeline**: Estimated time for each phase
      
      Consider:
      - Minimal disruption to existing functionality
      - Backward compatibility
      - Testing strategy
      - Documentation updates
      
      Output format: Structured plan with phases and tasks

  - name: "execute_reorganization"
    type: "execution"
    description: "Execute the reorganization plan"
    depends_on: ["create_reorganization_plan"]
    requires_approval: true
    jules_prompt: |
      Execute the reorganization plan for {{.ProjectPath}}:
      
      **Phase 1**: Create new directory structure
      **Phase 2**: Move files according to plan
      **Phase 3**: Update imports and references
      **Phase 4**: Update configuration files
      **Phase 5**: Run tests and validate
      
      For each change:
      1. Explain what you're doing
      2. Show the changes
      3. Verify the change works
      4. Update documentation if needed
      
      Be careful to:
      - Maintain functionality
      - Update all references
      - Run tests after each phase
      - Commit changes incrementally

# Validation rules
validation:
  pre_execution:
    - "check_git_status"
    - "verify_backup"
    - "validate_project_structure"
  
  post_execution:
    - "run_tests"
    - "check_build"
    - "validate_imports"
    - "verify_functionality"

# Output configuration
output:
  format: "markdown"
  include:
    - "summary"
    - "detailed_changes"
    - "test_results"
    - "recommendations"
  
  files:
    - path: "{{.ProjectPath}}/REORGANIZATION_REPORT.md"
      template: "reorganization_report.md"
    - path: "{{.ProjectPath}}/CHANGELOG.md"
      template: "changelog.md"
```

## 🔧 **Template Variables**

Templates support dynamic variables that are populated at runtime:

### **Built-in Variables**
- `{{.ProjectPath}}` - Path to the project
- `{{.ProjectName}}` - Name of the project
- `{{.ProjectType}}` - Type of project (go, python, js, etc.)
- `{{.Timestamp}}` - Current timestamp
- `{{.User}}` - Current user

### **Context Variables**
- `{{.FocusAreas}}` - Areas to focus on
- `{{.ExcludePatterns}}` - Patterns to exclude
- `{{.TargetArchitecture}}` - Target architecture
- `{{.CustomParams}}` - Custom parameters

## 📋 **Template Categories**

### **Reorganization Templates**
- `modular-restructure` - Convert to modular architecture
- `layered-architecture` - Implement layered architecture
- `microservices-split` - Split into microservices
- `monorepo-organization` - Organize monorepo structure

### **Refactoring Templates**
- `code-cleanup` - General code cleanup
- `dependency-update` - Update dependencies
- `api-modernization` - Modernize API design
- `performance-optimization` - Optimize performance

### **Testing Templates**
- `test-generation` - Generate comprehensive tests
- `test-coverage` - Improve test coverage
- `integration-tests` - Add integration tests
- `e2e-tests` - Add end-to-end tests

### **Documentation Templates**
- `api-documentation` - Generate API documentation
- `readme-generation` - Create comprehensive README
- `architecture-docs` - Document architecture
- `deployment-guide` - Create deployment guide

## 🚀 **Usage Examples**

### **Using Built-in Templates**

```bash
# Use modular reorganization template
./bin/jules-cli template run --template modular-restructure --project ./my-project

# Use custom parameters
./bin/jules-cli template run --template layered-architecture \
  --project ./my-project \
  --param target-layers="presentation,business,data" \
  --param preserve-tests=true
```

### **Creating Custom Templates**

```bash
# Create new template
./bin/jules-cli template create --name "my-custom-template" \
  --category "custom" \
  --description "My custom automation template"

# Edit template
./bin/jules-cli template edit --name "my-custom-template"
```

### **Template Registry Management**

```bash
# List available templates
./bin/jules-cli template list

# Show template details
./bin/jules-cli template show --name modular-restructure

# Validate template
./bin/jules-cli template validate --name modular-restructure

# Import template from URL
./bin/jules-cli template import --url https://example.com/template.yaml
```

## 🔍 **Context Extraction**

Templates can automatically extract project context:

### **Project Analysis**
- Directory structure analysis
- Dependency mapping
- Code pattern detection
- Architectural assessment

### **File Analysis**
- Language detection
- Framework identification
- Configuration analysis
- Test coverage analysis

### **Git Integration**
- Commit history analysis
- Branch structure
- Contributor analysis
- Change patterns

## 📊 **Template Registry**

The template registry (`templates/registry/registry.yaml`) maintains metadata about all available templates:

```yaml
templates:
  - name: "modular-restructure"
    version: "1.0.0"
    category: "reorganization"
    description: "Convert project to modular architecture"
    author: "Jules Automation"
    tags: ["modular", "architecture", "reorganization"]
    file: "builtin/reorganization/modular-restructure.yaml"
    dependencies: []
    compatibility:
      languages: ["go", "python", "javascript", "typescript"]
      frameworks: ["react", "vue", "angular", "express", "gin", "fastapi"]
    
  - name: "test-generation"
    version: "1.0.0"
    category: "testing"
    description: "Generate comprehensive test suite"
    author: "Jules Automation"
    tags: ["testing", "coverage", "automation"]
    file: "builtin/testing/test-generation.yaml"
    dependencies: ["project-analysis"]
    compatibility:
      languages: ["go", "python", "javascript", "typescript", "java"]
      frameworks: ["jest", "pytest", "testing", "junit"]
```

## 🎯 **Best Practices**

### **Template Design**
- Use clear, descriptive names
- Include comprehensive metadata
- Provide detailed descriptions
- Use semantic versioning
- Include compatibility information

### **Context Extraction**
- Be specific about file patterns
- Include relevant analysis steps
- Exclude unnecessary files
- Consider performance implications

### **Task Definition**
- Use clear, actionable prompts
- Include validation steps
- Provide rollback strategies
- Consider approval workflows

### **Output Configuration**
- Use structured output formats
- Include comprehensive reports
- Generate actionable recommendations
- Maintain audit trails

---

*Template system designed for Jules automation with Y2Q2-inspired YAML patterns*
