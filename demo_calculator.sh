#!/bin/bash

echo "=== Calculator Tool Demo ==="
echo

echo "1. Direct tool usage (no LLM):"
echo "-------------------------------"
./bin/agent-calculator | head -20
echo "... (truncated)"
echo

echo "2. LLM integration mode:"
echo "------------------------"
echo "Running: ./bin/agent-calculator llm"
echo "(This will use your configured LLM provider or mock if no API keys)"
echo

echo "3. Tool information mode:"
echo "-------------------------"
./bin/agent-calculator llm info
echo

echo "To run the full LLM example:"
echo "  ./bin/agent-calculator llm"
echo
echo "To enable debug logging:"
echo "  DEBUG=1 ./bin/agent-calculator llm"