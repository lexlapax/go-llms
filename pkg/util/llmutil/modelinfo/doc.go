// ABOUTME: Model information package providing LLM inventory and capability discovery.
// ABOUTME: Aggregates model data from providers with caching and service abstractions.
// Package modelinfo provides structures and functions to fetch, aggregate,
// and cache information about available Large Language Models (LLMs)
// from various providers.
//
// It defines the core domain models for model inventory, includes fetchers
// for specific providers (OpenAI, Google, Anthropic), a service to
// aggregate this information, and a caching mechanism to store and
// retrieve the inventory efficiently.
package modelinfo
