# Workflow Agents Package

This package contains the new implementation of workflow agents for the go-llms project.

## Overview

The workflow agents in this package are designed to orchestrate complex multi-step processes and multi-agent interactions. This is part of the Agent Architecture Restructuring Phase 3, building upon the core infrastructure established in Phases 1 and 1.5.

## Key Components

- **WorkflowAgent**: The main interface for workflow agents that extends BaseAgent with workflow-specific capabilities
- **Multi-agent orchestration**: Support for coordinating multiple agents to accomplish complex tasks
- **State management**: Advanced state handling for workflow execution
- **Handoff mechanisms**: Seamless delegation between agents using the Handoff interface

## Architecture

This implementation follows the patterns established in the core agent infrastructure:
- Uses the BaseAgent interface as foundation
- Integrates with the State management system
- Leverages the Event system for workflow coordination
- Supports Guardrails for workflow validation
- Implements TracingHook for workflow observability

## Status

This is a new implementation as part of Phase 3 of the Agent Architecture Restructuring project. The previous workflow package has been removed and this represents a clean, architecture-aligned approach to workflow agents.