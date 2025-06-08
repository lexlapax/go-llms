# Archived Documentation

This directory contains archived documentation that has been superseded by consolidated documents.

## Built-in Components Documentation

The following documents were consolidated into [docs/technical/built-in-components.md](/docs/technical/built-in-components.md):

- **BUILTIN_MIGRATION_BENEFITS.md** - Benefits comparison between common_tools.go and built-in tools
- **BUILTIN_COMPONENTS_DESIGN.md** - Original design document for built-in components
- **BUILTIN_COMPONENTS_IMPLEMENTATION_PLAN.md** - Detailed implementation plan with completion status
- **FEED_TOOLS_PLAN.md** - Implementation plan for feed processing tools

## Model Discovery Documentation

The following document was consolidated into [docs/user-guide/model-discovery.md](/docs/user-guide/model-discovery.md):

- **LIST_MODELS_ANALYSIS.md** - Architecture analysis for adding "List Models" capability to Go-LLMs

## Agent Framework Analysis Documentation

The following documents were used in the development of the agent architecture and are now consolidated into [docs/technical/agents.md](/docs/technical/agents.md):

- **analysis-agent-framework-claude.md** - Claude's analysis of agent frameworks and architectural patterns
- **analysis-agent-framework-gemini.md** - Gemini's comprehensive analysis of agent frameworks, including Google ADK patterns

## API Client Tool Documentation

The following documents were used in the development of the API client tool with GraphQL support:

- **GRAPHQL_API_CLIENT_DESIGN.md** - Overall GraphQL design with LLM-friendly approach
- **GRAPHQL_LIBRARY_ANALYSIS.md** - Analysis of various Go GraphQL libraries (selected gqlparser/v2)
- **GRAPHQL_PARAMETER_DESIGN.md** - GraphQL parameter integration strategy for the api_client tool

These documents are preserved for historical reference and to track the evolution of the built-in components system, model discovery features, agent architecture development, and API client tool enhancements.

## When to Reference These Documents

- To understand the original design decisions and rationale
- To track implementation progress and completion status
- To see detailed examples of specific implementations
- To understand the migration path from common_tools.go

For current documentation:
- Built-in components: [docs/technical/built-in-components.md](/docs/technical/built-in-components.md)
- Agent architecture: [docs/technical/agents.md](/docs/technical/agents.md)
- Model discovery: [docs/user-guide/model-discovery.md](/docs/user-guide/model-discovery.md)