# Workflow Composition Example

This example demonstrates advanced patterns for composing complex workflows from simpler workflow components. It shows how to build sophisticated agent systems by combining different workflow types.

## Patterns Demonstrated

### 1. Nested Workflows
Shows how to use workflows as steps within other workflows, creating hierarchical execution structures:
- Sub-workflows for validation, processing, and reporting
- Main workflow orchestrating the sub-workflows
- Clean separation of concerns

### 2. Pipeline Composition
Demonstrates building complex pipelines by chaining workflows:
- Information gathering (parallel workflow)
- Analysis (conditional workflow based on data type)
- Synthesis (sequential workflow)
- Each stage passes its results to the next

### 3. Complex Orchestration
Shows advanced composition mixing different workflow types:
- Loop workflows for retry logic
- Parallel workflows for concurrent processing
- Conditional workflows for branching logic
- Nested workflows for quality checks
- All composed into a sophisticated document processing system

### 4. Dynamic Composition
Demonstrates building workflows programmatically from configuration:
- Load workflow definitions from config
- Dynamically create workflow instances
- Compose them into a main workflow
- Useful for configurable, plugin-based systems

## Key Concepts

1. **Workflows as Building Blocks**: Workflows can contain other workflows, enabling modular design.

2. **Composition Patterns**:
   - **Hierarchical**: Workflows within workflows
   - **Sequential**: Pipeline-style composition
   - **Mixed**: Combining different workflow types

3. **State Flow**: State is passed through the entire workflow hierarchy, with each component able to read and modify it.

4. **Reusability**: Component workflows can be reused in different compositions.

## Running the Example

```bash
go run main.go
```

The example uses mock agents to demonstrate the patterns without requiring API keys.

## Use Cases

1. **Data Processing Pipelines**: ETL workflows with validation, transformation, and loading stages
2. **Document Processing**: Complex document analysis with multiple processing steps
3. **Research Workflows**: Multi-stage research with gathering, analysis, and synthesis
4. **Quality Assurance**: Workflows with validation loops and quality checks
5. **Microservice Orchestration**: Coordinating multiple services in complex business processes

## Best Practices

1. **Modular Design**: Create small, focused workflows that do one thing well
2. **Clear Interfaces**: Define clear state contracts between workflow components
3. **Error Handling**: Use appropriate error strategies for each workflow type
4. **Testing**: Test component workflows independently before composing
5. **Documentation**: Document the expected state inputs/outputs for each workflow