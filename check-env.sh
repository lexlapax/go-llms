#!/bin/bash

echo "Checking LLM provider environment variables..."
echo

# OpenAI
if [ -n "$OPENAI_API_KEY" ]; then
    echo "✓ OPENAI_API_KEY is set (length: ${#OPENAI_API_KEY})"
else
    echo "✗ OPENAI_API_KEY is not set"
fi

if [ -n "$GO_LLMS_OPENAI_API_KEY" ]; then
    echo "✓ GO_LLMS_OPENAI_API_KEY is set (length: ${#GO_LLMS_OPENAI_API_KEY})"
else
    echo "✗ GO_LLMS_OPENAI_API_KEY is not set"
fi

echo

# Anthropic
if [ -n "$ANTHROPIC_API_KEY" ]; then
    echo "✓ ANTHROPIC_API_KEY is set (length: ${#ANTHROPIC_API_KEY})"
else
    echo "✗ ANTHROPIC_API_KEY is not set"
fi

if [ -n "$GO_LLMS_ANTHROPIC_API_KEY" ]; then
    echo "✓ GO_LLMS_ANTHROPIC_API_KEY is set (length: ${#GO_LLMS_ANTHROPIC_API_KEY})"
else
    echo "✗ GO_LLMS_ANTHROPIC_API_KEY is not set"
fi

echo

# Gemini/Google
if [ -n "$GEMINI_API_KEY" ]; then
    echo "✓ GEMINI_API_KEY is set (length: ${#GEMINI_API_KEY})"
else
    echo "✗ GEMINI_API_KEY is not set"
fi

if [ -n "$GO_LLMS_GEMINI_API_KEY" ]; then
    echo "✓ GO_LLMS_GEMINI_API_KEY is set (length: ${#GO_LLMS_GEMINI_API_KEY})"
else
    echo "✗ GO_LLMS_GEMINI_API_KEY is not set"
fi

echo
echo "Note: You only need one of each pair to be set."
echo "The GO_LLMS_* versions take precedence if both are set."