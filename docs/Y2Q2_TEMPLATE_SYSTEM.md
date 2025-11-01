# Jules Automation - Y2Q2 Template System

This document explains the Y2Q2 (YAML-to-Query) inspired template system used in Jules automation for easy task and context management.

## 🎯 **Y2Q2 Philosophy**

The Y2Q2 template system is inspired by the concept of converting structured YAML configurations into executable queries and tasks. It provides:

- **Declarative Configuration**: Define automation tasks in YAML
- **Dynamic Context**: Automatic project analysis and context extraction
- **Template Reusability**: Shareable and customizable automation templates
- **Version Control**: Template versioning and compatibility management

## 📋 **Template Structure**

### **Core Components**

```yaml
# Template metadata
metadata:
  name: "template-name"
  version: "1.0.0"
  description: "Template description"
  author: "Author name"
  tags: ["tag1", "tag2"]
  category: "category"

# Template configuration
config:
  strategy: "strategy-name"
  max_concurrent_tasks: 3
  timeout: "300s"
  requires_approval: true
  backup_enabled: true

# Context extraction rules
context:
  project_analysis:
    - "analyze_directory_structure"
    - "identify_dependencies"
    - "analyze_code_patterns"
  
  file_patterns:
    include: ["**/*.go", "**/*.py"]
    exclude: ["**/node_modules/**"]

# Task definition
tasks:
  - name: "task-name"
    type: "task-type"
    description: "Task description"
    depends_on: ["previous-task"]
    jules_prompt: "Prompt for Jules AI"
    context_vars:
      VariableName: "{{.ProjectPath}}"

# Validation rules
validation:
  pre_execution: ["check_git_status"]
  post_execution: ["run_tests"]

# Output configuration
output:
  format: "markdown"
  include: ["summary", "recommendations"]
  files:
    - path: "{{.ProjectPath}}/REPORT.md"
      template: "report_template.md"
```

## 🔧 **Context Variables**

### **Built-in Variables**
- `{{.ProjectPath}}` - Path to the project
- `{{.ProjectName}}` - Name of the project
- `{{.ProjectType}}` - Type of project (go, python, js, etc.)
- `{{.Languages}}` - Detected programming languages
- `{{.Frameworks}}` - Detected frameworks
- `{{.Architecture}}` - Detected architecture pattern
- `{{.Complexity}}` - Project complexity assessment
- `{{.GitStatus}}` - Git repository status
- `{{.Timestamp}}` - Current timestamp

### **Custom Variables**
- `{{.CustomParam}}` - User-defined parameters
- `{{.FocusAreas}}` - Areas to focus on
- `{{.ExcludePatterns}}` - Patterns to exclude
- `{{.TargetArchitecture}}` - Target architecture

## 🎨 **Template Categories**

### **Reorganization Templates**
- **modular-restructure**: Convert to modular architecture
- **layered-architecture**: Implement layered architecture
- **microservices-split**: Split into microservices
- **monorepo-organization**: Organize monorepo structure

### **Testing Templates**
- **test-generation**: Generate comprehensive tests
- **test-coverage**: Improve test coverage
- **integration-tests**: Add integration tests
- **performance-tests**: Add performance tests

### **Refactoring Templates**
- **code-cleanup**: General code cleanup
- **dependency-update**: Update dependencies
- **api-modernization**: Modernize API design
- **performance-optimization**: Optimize performance

### **Documentation Templates**
- **api-documentation**: Generate API documentation
- **readme-generation**: Create comprehensive README
- **architecture-docs**: Document architecture
- **deployment-guide**: Create deployment guide

## 🚀 **Usage Examples**

### **Basic Template Usage**

```bash
# List available templates
jules-cli template list

# Show template details
jules-cli template show modular-restructure

# Execute template
jules-cli execute template modular-restructure ./my-project

# Execute with custom parameters
jules-cli execute template-with-params modular-restructure ./my-project \
  target-layers="presentation,business,data" \
  preserve-tests=true
```

### **MCP Integration**

```json
{
  "method": "tools/call",
  "params": {
    "name": "execute_template",
    "arguments": {
      "template_name": "modular-restructure",
      "project_path": "./my-project",
      "custom_params": {
        "target_layers": "presentation,business,data",
        "preserve_tests": "true"
      }
    }
  }
}
```

## 🔍 **Context Extraction**

### **Project Analysis**
The system automatically extracts project context:

```yaml
context:
  project_analysis:
    - "analyze_directory_structure"    # Analyze folder structure
    - "identify_dependencies"          # Find package dependencies
    - "analyze_code_patterns"          # Detect code patterns
    - "detect_architectural_issues"    # Find architectural problems
    - "analyze_import_graph"          # Map import relationships
```

### **File Pattern Matching**
Define which files to include/exclude:

```yaml
file_patterns:
  include:
    - "**/*.go"           # Go source files
    - "**/*.py"           # Python source files
    - "**/*.js"           # JavaScript files
    - "**/*.ts"           # TypeScript files
    - "**/*.java"         # Java source files
  exclude:
    - "**/node_modules/**" # Node.js dependencies
    - "**/vendor/**"       # Go dependencies
    - "**/.git/**"         # Git metadata
    - "**/dist/**"         # Build outputs
    - "**/build/**"        # Build outputs
```

## 📊 **Template Registry**

The template registry maintains metadata about all available templates:

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
    features:
      - "multi-phase-execution"
      - "approval-workflow"
      - "backup-enabled"
      - "architecture-validation"
    complexity: "high"
    estimated_duration: "2-4 hours"
```

## 🎯 **Task Types**

### **Analysis Tasks**
- **analyze_current_structure**: Analyze project structure
- **identify_dependencies**: Map dependencies
- **detect_issues**: Find problems and issues
- **assess_complexity**: Calculate complexity metrics

### **Planning Tasks**
- **create_plan**: Generate execution plan
- **risk_assessment**: Assess risks and mitigation
- **timeline_estimation**: Estimate execution time
- **resource_planning**: Plan required resources

### **Execution Tasks**
- **execute_changes**: Apply changes to codebase
- **update_dependencies**: Update project dependencies
- **refactor_code**: Refactor existing code
- **generate_tests**: Create test files

### **Validation Tasks**
- **run_tests**: Execute test suite
- **check_build**: Verify build process
- **validate_imports**: Check import statements
- **verify_functionality**: Test functionality

## 🔄 **Workflow Integration**

### **Multi-Phase Execution**
Templates can execute in multiple phases with approval points:

```yaml
tasks:
  - name: "phase_1_analysis"
    type: "analysis"
    description: "Analyze current state"
  
  - name: "phase_2_planning"
    type: "planning"
    depends_on: ["phase_1_analysis"]
    requires_approval: true
  
  - name: "phase_3_execution"
    type: "execution"
    depends_on: ["phase_2_planning"]
    requires_approval: true
```

### **Approval Workflow**
Critical changes can require user approval:

```yaml
config:
  requires_approval: true

tasks:
  - name: "critical_change"
    requires_approval: true
    jules_prompt: |
      This is a critical change that requires approval.
      Please review the proposed changes before proceeding.
```

### **Backup and Rollback**
Templates can automatically create backups:

```yaml
config:
  backup_enabled: true

validation:
  pre_execution:
    - "check_git_status"
    - "create_backup"
  post_execution:
    - "run_tests"
    - "validate_changes"
```

## 📈 **Advanced Features**

### **Dynamic Context Injection**
Templates can inject dynamic context based on project analysis:

```yaml
context_vars:
  ProjectPath: "{{.ProjectPath}}"
  FocusAreas: "{{.FocusAreas}}"
  TargetArchitecture: "{{.TargetArchitecture}}"
  CustomParam: "{{.CustomParam}}"
```

### **Conditional Execution**
Tasks can be conditionally executed based on context:

```yaml
tasks:
  - name: "conditional_task"
    type: "execution"
    condition: "{{.ProjectType}} == 'go'"
    jules_prompt: |
      This task only runs for Go projects.
      Project type: {{.ProjectType}}
```

### **Template Composition**
Templates can be composed from other templates:

```yaml
dependencies:
  - "base-template"
  - "testing-template"

tasks:
  - name: "inherit_base_tasks"
    type: "composition"
    template: "base-template"
```

## 🎨 **Best Practices**

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

## 🔧 **Custom Template Creation**

### **Creating Custom Templates**

```bash
# Create new template
jules-cli template create my-custom-template custom "My custom automation template"

# Edit template
jules-cli template edit my-custom-template

# Validate template
jules-cli template validate my-custom-template
```

### **Template Structure Example**

```yaml
metadata:
  name: "my-custom-template"
  version: "1.0.0"
  description: "My custom automation template"
  author: "Your Name"
  tags: ["custom", "automation"]
  category: "custom"

config:
  strategy: "custom"
  max_concurrent_tasks: 2
  timeout: "180s"
  requires_approval: false
  backup_enabled: true

context:
  project_analysis:
    - "analyze_custom_patterns"
    - "detect_custom_issues"
  
  file_patterns:
    include: ["**/*.custom"]
    exclude: ["**/exclude/**"]

tasks:
  - name: "custom_analysis"
    type: "analysis"
    description: "Perform custom analysis"
    jules_prompt: |
      Analyze the project in {{.ProjectPath}} for custom patterns.
      Focus on: {{.FocusAreas}}
      
      Provide recommendations for:
      1. Custom pattern improvements
      2. Issue resolution
      3. Optimization opportunities

  - name: "custom_execution"
    type: "execution"
    description: "Execute custom changes"
    depends_on: ["custom_analysis"]
    jules_prompt: |
      Based on the analysis, implement custom improvements:
      
      1. Apply recommended changes
      2. Update custom patterns
      3. Optimize performance
      4. Validate functionality
      
      Ensure all changes are properly tested and documented.

validation:
  pre_execution:
    - "check_custom_requirements"
  post_execution:
    - "run_custom_tests"
    - "validate_custom_patterns"

output:
  format: "markdown"
  include: ["summary", "custom_analysis", "recommendations"]
  files:
    - path: "{{.ProjectPath}}/CUSTOM_REPORT.md"
      template: "custom_report.md"
```

---

*Y2Q2 Template System - Converting YAML configurations into executable automation tasks*
