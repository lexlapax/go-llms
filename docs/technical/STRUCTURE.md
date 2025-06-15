# Technical Documentation Structure

> **[Documentation Home](README.md) / Structure Overview**

## Documentation Organization

This document outlines the complete structure of the go-llms technical documentation. Each section is designed for specific audiences and use cases.

## Complete Directory Structure

```
docs/technical-new/
├── README.md                    # Main navigation and overview
├── STRUCTURE.md                # This file - documentation map
├── architecture.md             # System design and principles
├── core-concepts.md           # Key abstractions and patterns
│
├── providers/                 # LLM Provider documentation
│   ├── README.md             # Provider documentation hub
│   ├── overview.md           # Provider system architecture
│   ├── implementing-providers.md  # Custom provider guide
│   ├── provider-registry.md  # Dynamic registration
│   └── metadata.md           # Capabilities and discovery
│
├── agents/                    # Agent framework documentation
│   ├── README.md             # Agent documentation hub
│   ├── overview.md           # Agent system architecture
│   ├── llm-agents.md         # AI-powered agents
│   ├── workflow-agents.md    # Orchestration patterns
│   ├── multi-agent-systems.md # Coordination patterns
│   └── state-management.md   # State and data flow
│
├── tools/                     # Tool system documentation
│   ├── README.md             # Tool documentation hub
│   ├── overview.md           # Tool architecture
│   ├── creating-tools.md     # Tool development guide
│   ├── tool-discovery.md     # Runtime registration
│   └── built-in-tools.md     # Available tools
│
├── development/              # Development guides
│   ├── README.md            # Development hub
│   ├── contributing.md      # Contribution guidelines
│   ├── testing.md           # Testing guide (created)
│   ├── api-design.md        # API design principles
│   └── best-practices.md    # Development best practices
│
├── advanced/                 # Advanced topics
│   ├── README.md            # Advanced topics hub
│   ├── performance.md       # Performance optimization
│   ├── event-system.md      # Event architecture (created)
│   ├── error-handling.md    # Error patterns
│   ├── schema-system.md     # JSON Schema details
│   └── bridge-integration.md # External integrations
│
└── api-reference/           # API documentation
    ├── README.md           # API reference hub
    ├── providers.md        # Provider interfaces
    ├── agents.md           # Agent interfaces
    ├── tools.md            # Tool interfaces
    └── types.md            # Type definitions
```

## Documentation Standards

![Package Structure](../images/package-structure.svg)
*Figure 1: Package structure and organization showing how the codebase maps to the documentation structure*

### 1. Consistent Navigation
Every document includes:
- Breadcrumb navigation at the top
- Links to related documents
- Clear section hierarchy

### 2. Document Structure
Each document follows:
```markdown
# Title
> **[Navigation](../README.md) / Current Page**

## Overview
Brief introduction and purpose

## Main Content
Organized sections with examples

## Best Practices
Practical recommendations

## Next Steps
Links to related topics
```

### 3. Code Examples
- Practical, runnable examples
- Comments explaining key points
- Error handling demonstrated
- Testing patterns included

### 4. Progressive Disclosure
- Start with overview
- Progress to details
- Advanced topics separate
- Examples throughout

## Target Audiences

### For Contributors
Primary documents:
1. [Architecture Overview](architecture.md)
2. [Core Concepts](core-concepts.md)
3. [Contributing Guide](development/contributing.md)
4. [Testing Guide](development/testing.md)
5. [API Design](development/api-design.md)

### For Provider Implementers
Primary documents:
1. [Provider Overview](providers/overview.md)
2. [Implementing Providers](providers/implementing-providers.md)
3. [Provider Registry](providers/provider-registry.md)
4. [Provider Metadata](providers/metadata.md)

### For Tool Developers
Primary documents:
1. [Tool Overview](tools/overview.md)
2. [Creating Tools](tools/creating-tools.md)
3. [Tool Discovery](tools/tool-discovery.md)
4. [Built-in Tools](tools/built-in-tools.md)

### For Advanced Users
Primary documents:
1. [Agent System](agents/overview.md)
2. [Workflow Patterns](agents/workflow-agents.md)
3. [Event System](advanced/event-system.md)
4. [Performance Guide](advanced/performance.md)

## Documentation Coverage

### Core Components ✅
- [x] Architecture overview
- [x] Core concepts explanation
- [x] Provider system
- [x] Agent framework
- [x] Tool system
- [x] State management

### Development ✅
- [x] Testing infrastructure
- [ ] Contributing workflow
- [ ] API design principles
- [ ] Best practices

### Advanced Topics ✅
- [x] Event system
- [ ] Performance optimization
- [ ] Error handling patterns
- [ ] Schema validation
- [ ] Bridge integration

### API Reference ✅
- [ ] Provider interfaces
- [ ] Agent interfaces
- [ ] Tool interfaces
- [ ] Type definitions

## Key Features Documented

### v0.3.5 Features
- ✅ Provider abstraction layer
- ✅ Agent framework with tools
- ✅ Workflow orchestration
- ✅ Event system
- ✅ State management
- ✅ Tool discovery
- ✅ Testing infrastructure
- ⏳ Schema validation (partial)
- ⏳ Bridge integration (partial)
- ⏳ Error handling system (partial)

### Documentation Quality
- ✅ Consistent formatting
- ✅ Comprehensive examples
- ✅ Clear navigation
- ✅ Progressive learning path
- ✅ Multiple audience support
- ✅ Practical focus

## Migration from Old Documentation

### Improvements Made
1. **Consistent Navigation**: All documents now have breadcrumbs
2. **Better Organization**: Clear hierarchy and grouping
3. **More Examples**: Practical code throughout
4. **Audience Focus**: Sections for different users
5. **Complete Coverage**: All major features documented

### Mapping Old to New
| Old Document | New Location |
|--------------|--------------|
| `architectur.md` | `architecture.md` |
| `provider-*.md` | `providers/*.md` |
| `agent-core.md` | `agents/llm-agents.md` |
| `workflow-*.md` | `agents/workflow-agents.md` |
| `tools.md`, `tool-*.md` | `tools/*.md` |
| `testing-*.md` | `development/testing.md` |

## Next Steps

1. **Complete Remaining Sections**: Fill in the remaining documented but not yet created files
2. **Add More Examples**: Expand code examples in each section
3. **Create Tutorials**: Step-by-step guides for common tasks
4. **API Reference**: Generate from code documentation
5. **Review and Polish**: Final consistency check

## Contributing to Documentation

See [Contributing Guide](development/contributing.md#documentation) for:
- Documentation standards
- Review process
- Style guide
- Example templates