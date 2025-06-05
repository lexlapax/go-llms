# State Persistence Example

This example demonstrates how to implement state persistence for agents, allowing conversations and data to be saved and restored across multiple sessions.

## Features

- Save agent state to JSON files
- Load previous conversation history
- Track session count
- Continue conversations across program runs

## Running the Example

```bash
go run main.go
```

Run the program multiple times to see the conversation history being preserved. Each run will:

1. Load any existing state from `agent_state.json`
2. Display previous conversation history
3. Accept new input from the user
4. Run the agent with the current state
5. Save the updated state back to file

## Key Concepts

### Persistent State Wrapper

The example creates a `PersistentState` type that wraps `domain.State` with save/load functionality:

```go
type PersistentState struct {
    *domain.State
    filename string
}
```

### State Serialization

State is serialized to JSON format, preserving:
- Conversation history
- Session count
- Any other custom state data

### Conversation History

The example maintains a conversation history as a slice of maps:

```go
conversationHistory := []map[string]string{
    {"user": "Hello", "agent": "Hi there!"},
    {"user": "How are you?", "agent": "I'm doing well, thanks!"},
}
```

## Use Cases

This pattern is useful for:

1. **Chat Applications**: Maintain conversation context across sessions
2. **Multi-Step Workflows**: Save progress through complex workflows
3. **User Preferences**: Store user-specific settings and preferences
4. **Analytics**: Track usage patterns and session data
5. **Debugging**: Save and replay agent interactions

## Extensions

You could extend this example to:

- Use a database instead of JSON files
- Implement encryption for sensitive data
- Add compression for large state files
- Support multiple conversation threads
- Implement state versioning and migration