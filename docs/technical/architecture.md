# Go-LLMs Architecture

> **[Documentation Home](/REFERENCE.md) / [Technical Documentation](/docs/technical/) / Architecture**

This document describes the architecture of the Go-LLMs library, providing insight into its design principles, component structure, and how different parts interact.

## Table of Contents

1. [Overview](#overview)
2. [Architecture Principles](#architecture-principles)
3. [Project Structure](#project-structure)
4. [Core Components](#core-components)
5. [Data Flow](#data-flow)
6. [Agent Workflow](#agent-workflow)
7. [Multi-Provider Strategies](#multi-provider-strategies)

## Overview

Go-LLMs is a Go library for creating LLM-powered applications with structured outputs and type safety. It aims to port the core functionality of pydantic-ai to Go while embracing Go's idioms and strengths.

The library is built around several key components that work together to provide a comprehensive solution for working with LLMs in Go applications:

![Go-LLMs Architecture Overview](/docs/images/architecture_overview.svg)

## Architecture Principles

Go-LLMs is built following these key architectural principles:

1. **Vertical Slicing**: Code is organized by feature rather than by layer, allowing for better cohesion and easier feature development.

2. **Interface-Based Design**: Core functionality is defined through interfaces, enabling loose coupling and easier testing.

3. **Clean Architecture**: The code follows clean architecture principles with clear domain boundaries and separation of concerns.

4. **Dependency Injection**: Components accept their dependencies from the outside, promoting testability and flexibility.

5. **Immutable Data Model**: Domain objects are designed to be immutable where possible, reducing the risk of shared state issues.

## Project Structure

The architecture follows a vertical slicing approach where code is organized by feature:

```
go-llms/
   cmd/                       # Application entry points
      examples/              # Example applications
   internal/                  # Internal packages
   pkg/                       # Public packages
      schema/                # Schema definition and validation
         domain/            # Core domain models and interfaces
         validation/        # Validation implementation
         adapter/           # Schema generation from Go structs
      llm/                   # LLM integration
         domain/            # Core domain models and interfaces
         provider/          # Provider implementations (OpenAI, Anthropic)
         prompt/            # Prompt templates and formatting
      structured/            # Structured output processing
         domain/            # Core domain models and interfaces
         processor/         # JSON extraction and validation
         adapter/           # External format adapters
      agent/                 # Agent orchestration
          domain/            # Core domain models and interfaces
          tools/             # Tool implementations
          workflow/          # Agent execution flow
   examples/                  # Usage examples
```

Each feature follows a common structure:
- **domain/**: Defines core domain models and interfaces
- **implementation/**: Contains the concrete implementations
- **adapter/**: Provides adapters for external integrations

## Core Components

### Schema Validation

The schema validation component provides:
- JSON Schema-compatible validation
- Type coercion to handle different input formats
- Structural validation for nested objects and arrays
- Format validation for common types (email, date-time, etc.)
- Advanced validation features (conditional validation, etc.)

### LLM Provider

The LLM provider component offers:
- Unified interface for different LLM providers (OpenAI, Anthropic, etc.)
- Streaming support for real-time responses
- Concurrent processing with multi-provider strategies
- Error handling and graceful degradation
- Caching and performance optimization

### Structured Output

The structured output component enables:
- Extracting structured data from LLM responses
- Schema-based validation of extracted data
- Type conversion to Go structs
- Prompt enhancement with schema information

### Agent System

The agent system provides:
- Tool integration for external actions
- Workflow management for complex interactions
- Message history management
- Monitoring and metrics through hooks
- Caching for improved performance

### Model Discovery

The model discovery system provides:
- Automatic model information fetching from provider APIs
- Capability detection for multimodal support, function calling, streaming
- Intelligent caching system for performance optimization
- Unified model inventory across all providers
- Flexible filtering and discovery based on capabilities

## Data Flow

This diagram shows how data flows through the system when generating structured outputs:

![Go-LLMs Data Flow](/docs/images/data_flow.svg)

The data flow follows these steps:

1. The application sends a request to an LLM provider
2. The provider formats the request for the external LLM API
3. The raw response from the LLM is returned to the provider
4. The structured output processor extracts structured data
5. The schema validator validates the extracted data
6. The valid result is returned to the application

## Agent Workflow

When using agents with tools, the flow of execution follows this pattern:

![Go-LLMs Agent Workflow](/docs/images/agent_workflow.svg)

The agent coordinates between:
- The LLM provider for generating responses
- Tools for performing specific operations
- Hooks for monitoring and logging
- User input/output handling

The key components in the agent workflow are:

1. **Agent**: Orchestrates the interaction between tools and the LLM
2. **Tools**: Provide functionality that the LLM can invoke
3. **LLM Provider**: Generates responses based on the conversation
4. **Message Manager**: Maintains conversation history
5. **Tool Executor**: Executes tool calls from the LLM
6. **Hooks**: Monitor and log agent activity

## Multi-Provider Strategies

Go-LLMs supports multiple strategies for working with several LLM providers simultaneously:

![Go-LLMs Multi-Provider Strategies](/docs/images/multi_provider.svg)

The available strategies are:

1. **Fastest Strategy**: Uses the first provider to respond
   - Sends requests to all providers concurrently
   - Returns the first successful response
   - Cancels other requests when the first response is received

2. **Primary Strategy**: Tries the primary provider first, with fallbacks
   - Sends the request to the primary provider
   - If the primary fails, tries fallback providers in sequence
   - Provides deterministic behavior for consistent responses

3. **Consensus Strategy**: Compares results from multiple providers to determine the best response
   - Sends requests to all providers concurrently
   - Compares the responses using similarity algorithms
   - Returns the most commonly occurring response
   - Can be weighted based on provider configurations

For more details on the multi-provider implementations, see the [Multi-Provider Guide](/docs/user-guide/multi-provider.md).