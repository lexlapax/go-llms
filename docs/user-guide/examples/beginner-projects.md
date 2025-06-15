# Beginner Projects

> **[User Guide](../README.md) / Examples / Beginner Projects**

Get hands-on experience with go-llms by building these 5 beginner-friendly projects. Each project introduces new concepts and builds your confidence with AI development.

## Prerequisites

- Completed [Quick Start](../getting-started/quickstart.md)
- Understanding of [Key Concepts](../getting-started/key-concepts.md)
- API key for your chosen provider

## Project Overview

| Project | What You'll Learn | Time | Difficulty |
|---------|------------------|------|------------|
| [Text Summarizer](#1-text-summarizer) | Basic prompting, structured output | 15 min | ⭐ |
| [Language Translator](#2-language-translator) | Schemas, validation, error handling | 20 min | ⭐ |
| [Code Explainer](#3-code-explainer) | File handling, code analysis | 25 min | ⭐⭐ |
| [Smart Calculator](#4-smart-calculator) | Tool creation, natural language processing | 30 min | ⭐⭐ |
| [Email Assistant](#5-email-assistant) | Templates, workflow agents, real application | 35 min | ⭐⭐⭐ |

## 1. Text Summarizer ⭐

**Goal:** Build a tool that summarizes long text into key points.

**What you'll learn:** Basic prompting, structured output, schema validation

### The Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/schema"
)

func main() {
    // Setup
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("Please set OPENAI_API_KEY environment variable")
    }
    
    provider := provider.NewOpenAIProvider(apiKey, "gpt-4")
    
    // Create summarizer agent
    summarizer := core.NewLLMAgent("summarizer", "gpt-4", core.LLMDeps{
        Provider: provider,
    })
    
    // Define what we want back
    summarySchema := &schema.Schema{
        Type: "object",
        Properties: map[string]*schema.Schema{
            "main_topic": {
                Type: "string",
                Description: "The main topic of the text",
            },
            "key_points": {
                Type: "array",
                Items: &schema.Schema{Type: "string"},
                Description: "3-5 key points from the text",
            },
            "summary": {
                Type: "string",
                Description: "One paragraph summary",
            },
        },
        Required: []string{"main_topic", "key_points", "summary"},
    }
    
    // Configure the agent
    summarizer.SetSystemPrompt(`You are an expert at summarizing text.
    Extract the main topic, key points, and create a concise summary.
    Be accurate and capture the most important information.`)
    
    summarizer.SetSchema(summarySchema)
    
    // Example text to summarize
    longText := `
    Artificial Intelligence (AI) has been transforming industries across the globe 
    at an unprecedented pace. In healthcare, AI algorithms are being used to diagnose 
    diseases earlier and more accurately than ever before. Machine learning models 
    can now detect cancer in medical imaging with higher accuracy rates than human 
    radiologists in some cases.
    
    In the automotive industry, self-driving cars powered by AI are becoming a reality. 
    Companies like Tesla, Google, and traditional automakers are investing billions 
    in autonomous vehicle technology. These vehicles use computer vision, sensor fusion, 
    and deep learning to navigate complex traffic situations.
    
    The financial sector has also embraced AI for fraud detection, algorithmic trading, 
    and risk assessment. Banks are using AI to analyze transaction patterns and detect 
    suspicious activities in real-time, significantly reducing financial fraud.
    
    However, with these advances come important ethical considerations. Issues around 
    job displacement, privacy, bias in AI algorithms, and the need for transparent 
    AI decision-making are becoming increasingly important. Governments and organizations 
    worldwide are working to establish frameworks for responsible AI development and deployment.
    `
    
    // Summarize the text
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf("Please summarize this text: %s", longText))
    
    result, err := summarizer.Run(context.Background(), state)
    if err != nil {
        log.Fatal("Error summarizing:", err)
    }
    
    // Get structured output
    if summary, exists := result.Get("structured_output"); exists {
        summaryData := summary.(map[string]interface{})
        
        fmt.Println("📄 TEXT SUMMARIZER RESULTS")
        fmt.Println("==========================")
        fmt.Printf("📌 Main Topic: %s\n\n", summaryData["main_topic"])
        
        fmt.Println("🔍 Key Points:")
        if keyPoints, ok := summaryData["key_points"].([]interface{}); ok {
            for i, point := range keyPoints {
                fmt.Printf("  %d. %s\n", i+1, point)
            }
        }
        
        fmt.Printf("\n📝 Summary:\n%s\n", summaryData["summary"])
    }
}
```

### Expected Output

```
📄 TEXT SUMMARIZER RESULTS
==========================
📌 Main Topic: AI transformation across industries and ethical considerations

🔍 Key Points:
  1. AI is improving healthcare diagnosis and medical imaging analysis
  2. Autonomous vehicles are becoming reality with major industry investment
  3. Financial sector uses AI for fraud detection and risk assessment
  4. Ethical concerns include job displacement, privacy, and algorithmic bias
  5. Need for responsible AI development frameworks

📝 Summary:
Artificial Intelligence is rapidly transforming multiple industries including healthcare, automotive, and finance through improved diagnostics, autonomous vehicles, and fraud detection. However, this progress raises important ethical considerations around job displacement, privacy, and algorithmic bias, prompting the need for responsible AI development frameworks.
```

### What You Learned

✅ How to use schemas for structured output  
✅ Creating agents with specific purposes  
✅ Handling complex text processing  
✅ Working with nested data structures  

## 2. Language Translator ⭐

**Goal:** Build a multi-language translator with confidence scoring.

**What you'll learn:** Input validation, error handling, multiple providers

### The Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/schema"
)

func main() {
    // Setup provider
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("Please set OPENAI_API_KEY environment variable")
    }
    
    provider := provider.NewOpenAIProvider(apiKey, "gpt-4")
    
    // Create translator agent
    translator := core.NewLLMAgent("translator", "gpt-4", core.LLMDeps{
        Provider: provider,
    })
    
    // Define translation schema
    translationSchema := &schema.Schema{
        Type: "object",
        Properties: map[string]*schema.Schema{
            "detected_language": {
                Type: "string",
                Description: "The detected language of the input text",
            },
            "target_language": {
                Type: "string", 
                Description: "The target language for translation",
            },
            "translation": {
                Type: "string",
                Description: "The translated text",
            },
            "confidence": {
                Type: "number",
                Minimum: &[]float64{0}[0],
                Maximum: &[]float64{1}[0],
                Description: "Confidence score of the translation (0-1)",
            },
            "notes": {
                Type: "string",
                Description: "Any notes about the translation (idioms, cultural context, etc.)",
            },
        },
        Required: []string{"detected_language", "target_language", "translation", "confidence"},
    }
    
    // Configure translator
    translator.SetSystemPrompt(`You are an expert translator who speaks many languages fluently.
    
    Your tasks:
    1. Detect the source language of the input text
    2. Translate it accurately to the target language
    3. Provide a confidence score (0-1) for your translation
    4. Add notes if there are cultural nuances, idioms, or context considerations
    
    Be accurate and maintain the original tone and meaning.`)
    
    translator.SetSchema(translationSchema)
    
    // Translation examples
    translations := []struct {
        text       string
        targetLang string
    }{
        {"Hello, how are you today?", "Spanish"},
        {"Je suis très heureux de vous rencontrer", "English"},
        {"この本はとても面白いです", "English"},
        {"Wie geht es dir heute?", "French"},
        {"¡Qué hermoso día hace hoy!", "English"},
    }
    
    fmt.Println("🌍 MULTI-LANGUAGE TRANSLATOR")
    fmt.Println("============================")
    
    for i, example := range translations {
        fmt.Printf("\n--- Translation %d ---\n", i+1)
        fmt.Printf("Original: %s\n", example.text)
        fmt.Printf("Target: %s\n", example.targetLang)
        
        // Create translation request
        state := domain.NewState()
        state.Set("user_input", fmt.Sprintf(
            "Translate this text to %s: %s", 
            example.targetLang, 
            example.text,
        ))
        
        result, err := translator.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Translation error: %v\n", err)
            continue
        }
        
        if output, exists := result.Get("structured_output"); exists {
            data := output.(map[string]interface{})
            
            fmt.Printf("Detected: %s\n", data["detected_language"])
            fmt.Printf("Translation: %s\n", data["translation"])
            fmt.Printf("Confidence: %.2f\n", data["confidence"])
            
            if notes, ok := data["notes"].(string); ok && notes != "" {
                fmt.Printf("Notes: %s\n", notes)
            }
        }
    }
    
    // Interactive mode
    fmt.Println("\n🔄 Interactive Mode (type 'quit' to exit)")
    fmt.Println("Format: text | target_language")
    fmt.Println("Example: Hello world | French")
    
    for {
        fmt.Print("\nTranslate: ")
        var input string
        fmt.Scanln(&input)
        
        if input == "quit" {
            break
        }
        
        // Parse input
        parts := strings.Split(input, "|")
        if len(parts) != 2 {
            fmt.Println("Please use format: text | target_language")
            continue
        }
        
        text := strings.TrimSpace(parts[0])
        targetLang := strings.TrimSpace(parts[1])
        
        if text == "" || targetLang == "" {
            fmt.Println("Both text and target language are required")
            continue
        }
        
        // Translate
        state := domain.NewState()
        state.Set("user_input", fmt.Sprintf("Translate this text to %s: %s", targetLang, text))
        
        result, err := translator.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        if output, exists := result.Get("structured_output"); exists {
            data := output.(map[string]interface{})
            fmt.Printf("→ %s (confidence: %.2f)\n", 
                data["translation"], 
                data["confidence"])
        }
    }
}
```

### Expected Output

```
🌍 MULTI-LANGUAGE TRANSLATOR
============================

--- Translation 1 ---
Original: Hello, how are you today?
Target: Spanish
Detected: English
Translation: Hola, ¿cómo estás hoy?
Confidence: 0.95

--- Translation 2 ---
Original: Je suis très heureux de vous rencontrer
Target: English
Detected: French
Translation: I am very happy to meet you
Confidence: 0.98

🔄 Interactive Mode (type 'quit' to exit)
Format: text | target_language
Example: Hello world | French

Translate: Good morning | Japanese
→ おはようございます (confidence: 0.92)
```

### What You Learned

✅ Input parsing and validation  
✅ Confidence scoring for AI outputs  
✅ Interactive command-line applications  
✅ Handling multiple languages  

## 3. Code Explainer ⭐⭐

**Goal:** Build a tool that reads code files and explains what they do.

**What you'll learn:** File I/O, code analysis, detailed prompting

### The Code

```go
package main

import (
    "context"
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "strings"

    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/schema"
)

func main() {
    // Setup
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("Please set OPENAI_API_KEY environment variable")
    }
    
    provider := provider.NewOpenAIProvider(apiKey, "gpt-4")
    
    // Create code explainer agent
    explainer := core.NewLLMAgent("code-explainer", "gpt-4", core.LLMDeps{
        Provider: provider,
    })
    
    // Define explanation schema
    explanationSchema := &schema.Schema{
        Type: "object",
        Properties: map[string]*schema.Schema{
            "language": {
                Type: "string",
                Description: "Programming language detected",
            },
            "purpose": {
                Type: "string",
                Description: "What this code does in one sentence",
            },
            "breakdown": {
                Type: "array",
                Items: &schema.Schema{
                    Type: "object",
                    Properties: map[string]*schema.Schema{
                        "section": {Type: "string"},
                        "explanation": {Type: "string"},
                    },
                },
                Description: "Step-by-step breakdown of the code",
            },
            "key_concepts": {
                Type: "array",
                Items: &schema.Schema{Type: "string"},
                Description: "Important programming concepts used",
            },
            "complexity": {
                Type: "string",
                Enum: []interface{}{"beginner", "intermediate", "advanced"},
                Description: "Complexity level",
            },
        },
        Required: []string{"language", "purpose", "breakdown", "key_concepts", "complexity"},
    }
    
    // Configure explainer
    explainer.SetSystemPrompt(`You are an expert programming tutor who explains code clearly.
    
    Your job:
    1. Identify the programming language
    2. Explain the overall purpose
    3. Break down the code section by section
    4. Identify key programming concepts
    5. Rate the complexity level
    
    Be clear, educational, and assume the reader is learning.`)
    
    explainer.SetSchema(explanationSchema)
    
    // Check command line arguments
    if len(os.Args) < 2 {
        fmt.Println("Usage: go run code-explainer.go <file-path>")
        fmt.Println("Or create a sample file to analyze...")
        
        // Create a sample file
        sampleCode := `def fibonacci(n):
    """Calculate the nth Fibonacci number using recursion."""
    if n <= 1:
        return n
    else:
        return fibonacci(n-1) + fibonacci(n-2)

# Test the function
for i in range(10):
    print(f"F({i}) = {fibonacci(i)}")
`
        
        err := ioutil.WriteFile("sample_fibonacci.py", []byte(sampleCode), 0644)
        if err != nil {
            log.Fatal("Error creating sample file:", err)
        }
        
        fmt.Println("Created sample_fibonacci.py")
        fmt.Println("Run: go run code-explainer.go sample_fibonacci.py")
        return
    }
    
    filePath := os.Args[1]
    
    // Read the code file
    code, err := ioutil.ReadFile(filePath)
    if err != nil {
        log.Fatal("Error reading file:", err)
    }
    
    fileName := filepath.Base(filePath)
    
    fmt.Printf("🔍 CODE EXPLAINER - Analyzing %s\n", fileName)
    fmt.Println("=====================================")
    
    // Analyze the code
    state := domain.NewState()
    state.Set("user_input", fmt.Sprintf(
        "Please explain this code file (%s):\n\n%s", 
        fileName, 
        string(code),
    ))
    
    result, err := explainer.Run(context.Background(), state)
    if err != nil {
        log.Fatal("Error analyzing code:", err)
    }
    
    if output, exists := result.Get("structured_output"); exists {
        data := output.(map[string]interface{})
        
        // Display results
        fmt.Printf("📝 Language: %s\n", data["language"])
        fmt.Printf("🎯 Purpose: %s\n\n", data["purpose"])
        
        fmt.Println("📋 Code Breakdown:")
        if breakdown, ok := data["breakdown"].([]interface{}); ok {
            for i, item := range breakdown {
                if section, ok := item.(map[string]interface{}); ok {
                    fmt.Printf("  %d. %s\n", i+1, section["section"])
                    fmt.Printf("     %s\n\n", section["explanation"])
                }
            }
        }
        
        fmt.Println("🔑 Key Concepts:")
        if concepts, ok := data["key_concepts"].([]interface{}); ok {
            for _, concept := range concepts {
                fmt.Printf("  • %s\n", concept)
            }
        }
        
        fmt.Printf("\n⚡ Complexity: %s\n", data["complexity"])
    }
    
    // Interactive mode for multiple files
    fmt.Println("\n🔄 Want to analyze another file? (y/n)")
    var response string
    fmt.Scanln(&response)
    
    if strings.ToLower(response) == "y" {
        fmt.Print("Enter file path: ")
        var newPath string
        fmt.Scanln(&newPath)
        
        if newPath != "" {
            // Recursive call with new arguments
            os.Args[1] = newPath
            main()
        }
    }
}
```

### Example Analysis Output

```
🔍 CODE EXPLAINER - Analyzing sample_fibonacci.py
=====================================
📝 Language: Python
🎯 Purpose: Calculates and prints the first 10 Fibonacci numbers using a recursive function

📋 Code Breakdown:
  1. Function Definition
     Defines a recursive function 'fibonacci' that calculates the nth Fibonacci number

  2. Base Case
     Handles the base cases where n <= 1, returning n directly

  3. Recursive Case  
     For n > 1, returns the sum of fibonacci(n-1) + fibonacci(n-2)

  4. Documentation
     Includes a docstring explaining what the function does

  5. Test Loop
     Uses a for loop to calculate and print Fibonacci numbers from 0 to 9

🔑 Key Concepts:
  • Recursion
  • Base cases
  • Function documentation
  • String formatting (f-strings)
  • For loops

⚡ Complexity: intermediate
```

### What You Learned

✅ File I/O operations  
✅ Code analysis and explanation  
✅ Command-line argument handling  
✅ Structured data processing  

## 4. Smart Calculator ⭐⭐

**Goal:** Build a calculator that understands natural language math questions.

**What you'll learn:** Tool creation, natural language processing, complex prompting

### The Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math"
    "os"
    "strconv"
    "strings"

    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/agent/tools"
    "github.com/lexlapax/go-llms/pkg/schema"
)

// Create math calculation tool
func createMathTool() domain.Tool {
    mathSchema := &schema.Schema{
        Type: "object",
        Properties: map[string]*schema.Schema{
            "operation": {
                Type: "string",
                Enum: []interface{}{"add", "subtract", "multiply", "divide", "power", "sqrt", "sin", "cos", "tan", "log"},
                Description: "Mathematical operation to perform",
            },
            "numbers": {
                Type: "array",
                Items: &schema.Schema{Type: "number"},
                Description: "Numbers to use in the calculation",
            },
        },
        Required: []string{"operation", "numbers"},
    }
    
    return tools.NewTool(
        "math_calculator",
        "Perform mathematical calculations",
        mathSchema,
        func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
            operation := params["operation"].(string)
            numbersInterface := params["numbers"].([]interface{})
            
            // Convert to float64 slice
            numbers := make([]float64, len(numbersInterface))
            for i, n := range numbersInterface {
                switch v := n.(type) {
                case float64:
                    numbers[i] = v
                case int:
                    numbers[i] = float64(v)
                default:
                    return nil, fmt.Errorf("invalid number: %v", n)
                }
            }
            
            switch operation {
            case "add":
                result := 0.0
                for _, n := range numbers {
                    result += n
                }
                return result, nil
                
            case "subtract":
                if len(numbers) < 2 {
                    return nil, fmt.Errorf("subtraction requires at least 2 numbers")
                }
                result := numbers[0]
                for _, n := range numbers[1:] {
                    result -= n
                }
                return result, nil
                
            case "multiply":
                result := 1.0
                for _, n := range numbers {
                    result *= n
                }
                return result, nil
                
            case "divide":
                if len(numbers) != 2 {
                    return nil, fmt.Errorf("division requires exactly 2 numbers")
                }
                if numbers[1] == 0 {
                    return nil, fmt.Errorf("cannot divide by zero")
                }
                return numbers[0] / numbers[1], nil
                
            case "power":
                if len(numbers) != 2 {
                    return nil, fmt.Errorf("power requires exactly 2 numbers")
                }
                return math.Pow(numbers[0], numbers[1]), nil
                
            case "sqrt":
                if len(numbers) != 1 {
                    return nil, fmt.Errorf("square root requires exactly 1 number")
                }
                if numbers[0] < 0 {
                    return nil, fmt.Errorf("cannot take square root of negative number")
                }
                return math.Sqrt(numbers[0]), nil
                
            case "sin":
                if len(numbers) != 1 {
                    return nil, fmt.Errorf("sin requires exactly 1 number")
                }
                return math.Sin(numbers[0]), nil
                
            case "cos":
                if len(numbers) != 1 {
                    return nil, fmt.Errorf("cos requires exactly 1 number")
                }
                return math.Cos(numbers[0]), nil
                
            case "tan":
                if len(numbers) != 1 {
                    return nil, fmt.Errorf("tan requires exactly 1 number")
                }
                return math.Tan(numbers[0]), nil
                
            case "log":
                if len(numbers) != 1 {
                    return nil, fmt.Errorf("log requires exactly 1 number")
                }
                if numbers[0] <= 0 {
                    return nil, fmt.Errorf("cannot take log of non-positive number")
                }
                return math.Log(numbers[0]), nil
                
            default:
                return nil, fmt.Errorf("unsupported operation: %s", operation)
            }
        },
    )
}

func main() {
    // Setup
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("Please set OPENAI_API_KEY environment variable")
    }
    
    provider := provider.NewOpenAIProvider(apiKey, "gpt-4")
    
    // Create smart calculator agent
    calculator := core.NewLLMAgent("smart-calculator", "gpt-4", core.LLMDeps{
        Provider: provider,
    })
    
    // Configure calculator
    calculator.SetSystemPrompt(`You are a smart calculator that understands natural language math questions.
    
    When users ask math questions:
    1. Parse what calculation they want
    2. Use the math_calculator tool to perform the calculation
    3. Explain the result in a friendly way
    
    You can handle:
    - Basic operations: addition, subtraction, multiplication, division
    - Advanced operations: powers, square roots, trigonometry, logarithms
    - Word problems: "If I have 15 apples and give away 7..."
    - Unit conversions: "How many minutes in 2.5 hours?"
    
    Always show your work and explain the calculation step by step.`)
    
    // Add math tool
    calculator.AddTool(createMathTool())
    
    fmt.Println("🧮 SMART CALCULATOR")
    fmt.Println("===================")
    fmt.Println("Ask me any math question in natural language!")
    fmt.Println("Examples:")
    fmt.Println("  • What's 15 times 23?")
    fmt.Println("  • What's the square root of 144?")
    fmt.Println("  • If I have 50 dollars and spend 17.50, how much is left?")
    fmt.Println("  • What's 2 to the power of 8?")
    fmt.Println()
    
    // Example calculations
    examples := []string{
        "What's 15 times 23?",
        "What's the square root of 144?", 
        "If I have 50 dollars and spend 17.50, how much is left?",
        "What's 2 to the power of 8?",
        "What's the sine of 30 degrees?",
    }
    
    fmt.Println("📊 Example Calculations:")
    for i, example := range examples {
        fmt.Printf("\n--- Example %d ---\n", i+1)
        fmt.Printf("Question: %s\n", example)
        
        state := domain.NewState()
        state.Set("user_input", example)
        
        result, err := calculator.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        if response, exists := result.Get("response"); exists {
            fmt.Printf("Answer: %s\n", response)
        }
    }
    
    // Interactive mode
    fmt.Println("\n🔄 Interactive Mode (type 'quit' to exit)")
    
    for {
        fmt.Print("\nMath question: ")
        var input string
        fmt.Scanln(&input)
        
        if input == "quit" {
            fmt.Println("👋 Thanks for using Smart Calculator!")
            break
        }
        
        if input == "" {
            continue
        }
        
        state := domain.NewState()
        state.Set("user_input", input)
        
        result, err := calculator.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        if response, exists := result.Get("response"); exists {
            fmt.Printf("🧮 %s\n", response)
        }
    }
}
```

### Expected Output

```
🧮 SMART CALCULATOR
===================
Ask me any math question in natural language!
Examples:
  • What's 15 times 23?
  • What's the square root of 144?
  • If I have 50 dollars and spend 17.50, how much is left?
  • What's 2 to the power of 8?

📊 Example Calculations:

--- Example 1 ---
Question: What's 15 times 23?
Answer: I'll calculate 15 times 23 for you. 15 × 23 = 345

--- Example 2 ---
Question: What's the square root of 144?
Answer: The square root of 144 is 12. This is because 12 × 12 = 144.

🔄 Interactive Mode (type 'quit' to exit)

Math question: How much is 25% of 80?
🧮 To find 25% of 80, I need to multiply 80 by 0.25. 80 × 0.25 = 20. So 25% of 80 is 20.
```

### What You Learned

✅ Creating custom tools with complex logic  
✅ Natural language processing for math  
✅ Error handling in tools  
✅ Interactive applications with AI  

## 5. Email Assistant ⭐⭐⭐

**Goal:** Build an assistant that helps write different types of emails.

**What you'll learn:** Templates, workflow agents, real-world applications

### The Code

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "strings"

    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/domain"
    "github.com/lexlapax/go-llms/pkg/schema"
)

func main() {
    // Setup
    apiKey := os.Getenv("OPENAI_API_KEY")
    if apiKey == "" {
        log.Fatal("Please set OPENAI_API_KEY environment variable")
    }
    
    provider := provider.NewOpenAIProvider(apiKey, "gpt-4")
    
    // Create email analysis agent
    analyzer := core.NewLLMAgent("email-analyzer", "gpt-4", core.LLMDeps{
        Provider: provider,
    })
    
    analysisSchema := &schema.Schema{
        Type: "object",
        Properties: map[string]*schema.Schema{
            "email_type": {
                Type: "string",
                Enum: []interface{}{"business", "personal", "complaint", "thank_you", "request", "follow_up"},
                Description: "Type of email to write",
            },
            "tone": {
                Type: "string", 
                Enum: []interface{}{"formal", "casual", "friendly", "professional", "urgent"},
                Description: "Desired tone for the email",
            },
            "key_points": {
                Type: "array",
                Items: &schema.Schema{Type: "string"},
                Description: "Main points to include in the email",
            },
        },
        Required: []string{"email_type", "tone", "key_points"},
    }
    
    analyzer.SetSystemPrompt(`You are an email writing expert. Analyze the user's request and determine:
    1. What type of email they want to write
    2. What tone is appropriate
    3. What key points should be included
    
    Be practical and consider professional communication standards.`)
    
    analyzer.SetSchema(analysisSchema)
    
    // Create email writer agent
    writer := core.NewLLMAgent("email-writer", "gpt-4", core.LLMDeps{
        Provider: provider,
    })
    
    emailSchema := &schema.Schema{
        Type: "object",
        Properties: map[string]*schema.Schema{
            "subject": {
                Type: "string",
                Description: "Email subject line",
            },
            "body": {
                Type: "string",
                Description: "Email body content",
            },
            "tips": {
                Type: "array",
                Items: &schema.Schema{Type: "string"}, 
                Description: "Tips for improving the email",
            },
        },
        Required: []string{"subject", "body"},
    }
    
    writer.SetSystemPrompt(`You are an expert email writer. Based on the analysis provided:
    1. Write a compelling subject line
    2. Write a well-structured email body
    3. Provide tips for improvement
    
    Match the requested tone and include all key points naturally.
    Follow email best practices for structure and clarity.`)
    
    writer.SetSchema(emailSchema)
    
    // Create email review agent  
    reviewer := core.NewLLMAgent("email-reviewer", "gpt-4", core.LLMDeps{
        Provider: provider,
    })
    
    reviewSchema := &schema.Schema{
        Type: "object",
        Properties: map[string]*schema.Schema{
            "score": {
                Type: "number",
                Minimum: &[]float64{1}[0],
                Maximum: &[]float64{10}[0],
                Description: "Quality score from 1-10",
            },
            "strengths": {
                Type: "array",
                Items: &schema.Schema{Type: "string"},
                Description: "What's good about this email",
            },
            "improvements": {
                Type: "array", 
                Items: &schema.Schema{Type: "string"},
                Description: "Suggested improvements",
            },
            "final_version": {
                Type: "string",
                Description: "Improved version of the email body",
            },
        },
        Required: []string{"score", "strengths", "improvements", "final_version"},
    }
    
    reviewer.SetSystemPrompt(`You are an email quality reviewer. Review the email and:
    1. Give it a score from 1-10
    2. Identify strengths
    3. Suggest specific improvements
    4. Provide an improved final version
    
    Focus on clarity, professionalism, and effectiveness.`)
    
    reviewer.SetSchema(reviewSchema)
    
    // Create workflow that combines all agents
    emailWorkflow := workflow.NewSequentialAgent("email-assistant", []domain.BaseAgent{
        analyzer,
        writer, 
        reviewer,
    })
    
    fmt.Println("📧 EMAIL ASSISTANT")
    fmt.Println("==================")
    fmt.Println("I'll help you write professional emails!")
    fmt.Println("Tell me what kind of email you need to write.")
    fmt.Println()
    
    // Example email requests
    examples := []string{
        "I need to write a thank you email to my manager for approving my vacation request",
        "I want to complain to customer service about a defective product I received", 
        "I need to follow up with a client about a project proposal I sent last week",
    }
    
    fmt.Println("📋 Example Email Requests:")
    for i, example := range examples {
        fmt.Printf("\n--- Example %d ---\n", i+1)
        fmt.Printf("Request: %s\n", example)
        
        state := domain.NewState()
        state.Set("user_input", example)
        
        result, err := emailWorkflow.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        // Get analysis
        if analysis, exists := result.Get("analysis"); exists {
            analysisData := analysis.(map[string]interface{})
            fmt.Printf("Type: %s | Tone: %s\n", 
                analysisData["email_type"], 
                analysisData["tone"])
        }
        
        // Get email draft
        if email, exists := result.Get("email"); exists {
            emailData := email.(map[string]interface{})
            fmt.Printf("Subject: %s\n", emailData["subject"])
            fmt.Printf("Body preview: %s...\n", 
                strings.TrimSpace(emailData["body"].(string))[:100])
        }
        
        // Get review
        if review, exists := result.Get("review"); exists {
            reviewData := review.(map[string]interface{})
            fmt.Printf("Quality Score: %.1f/10\n", reviewData["score"])
        }
    }
    
    // Interactive mode
    fmt.Println("\n🔄 Interactive Mode (type 'quit' to exit)")
    fmt.Println("Describe the email you need to write:")
    
    for {
        fmt.Print("\nEmail request: ")
        var input string
        fmt.Scanln(&input)
        
        if input == "quit" {
            fmt.Println("👋 Happy emailing!")
            break
        }
        
        if input == "" {
            continue
        }
        
        state := domain.NewState()
        state.Set("user_input", input)
        
        fmt.Println("⏳ Processing your request...")
        
        result, err := emailWorkflow.Run(context.Background(), state)
        if err != nil {
            fmt.Printf("❌ Error: %v\n", err)
            continue
        }
        
        // Display full results
        fmt.Println("\n📊 EMAIL ANALYSIS:")
        if analysis, exists := result.Get("analysis"); exists {
            analysisData := analysis.(map[string]interface{})
            fmt.Printf("  Type: %s\n", analysisData["email_type"])
            fmt.Printf("  Tone: %s\n", analysisData["tone"])
            fmt.Printf("  Key Points: %v\n", analysisData["key_points"])
        }
        
        fmt.Println("\n📝 EMAIL DRAFT:")
        if email, exists := result.Get("email"); exists {
            emailData := email.(map[string]interface{})
            fmt.Printf("  Subject: %s\n\n", emailData["subject"])
            fmt.Printf("  Body:\n%s\n", emailData["body"])
        }
        
        fmt.Println("\n📋 QUALITY REVIEW:")
        if review, exists := result.Get("review"); exists {
            reviewData := review.(map[string]interface{})
            fmt.Printf("  Score: %.1f/10\n", reviewData["score"])
            
            if strengths, ok := reviewData["strengths"].([]interface{}); ok {
                fmt.Println("  Strengths:")
                for _, strength := range strengths {
                    fmt.Printf("    ✓ %s\n", strength)
                }
            }
            
            if improvements, ok := reviewData["improvements"].([]interface{}); ok {
                fmt.Println("  Improvements:")
                for _, improvement := range improvements {
                    fmt.Printf("    → %s\n", improvement)
                }
            }
            
            fmt.Printf("\n  Final Version:\n%s\n", reviewData["final_version"])
        }
    }
}
```

### Expected Output

```
📧 EMAIL ASSISTANT
==================
I'll help you write professional emails!
Tell me what kind of email you need to write.

📋 Example Email Requests:

--- Example 1 ---
Request: I need to write a thank you email to my manager for approving my vacation request
Type: thank_you | Tone: professional
Subject: Thank you for approving my vacation request
Body preview: Dear [Manager's Name], I wanted to take a moment to express my sincere gratitude for approving...
Quality Score: 8.5/10

🔄 Interactive Mode (type 'quit' to exit)
Describe the email you need to write:

Email request: I need to ask my colleague for help with a project deadline
⏳ Processing your request...

📊 EMAIL ANALYSIS:
  Type: request
  Tone: professional
  Key Points: [project assistance, deadline urgency, collaboration]

📝 EMAIL DRAFT:
  Subject: Request for Project Assistance - [Project Name]

  Body:
  Hi [Colleague's Name],

  I hope this email finds you well. I'm reaching out to request your assistance with [Project Name], which has an upcoming deadline.

  [Rest of email body...]

📋 QUALITY REVIEW:
  Score: 8.0/10
  Strengths:
    ✓ Clear subject line
    ✓ Professional tone
    ✓ Specific request
  Improvements:
    → Add specific deadline date
    → Mention what kind of help is needed
```

### What You Learned

✅ Building complex workflows with multiple agents  
✅ Sequential agent processing  
✅ Real-world application development  
✅ Quality assurance with AI review  

## 🎉 Congratulations!

You've completed all 5 beginner projects and learned:

- **Text Processing** - Summarization and structured output
- **Language Skills** - Translation with confidence scoring  
- **Code Analysis** - File reading and code explanation
- **Tool Creation** - Custom tools for calculations
- **Workflow Design** - Multi-agent email assistance

## What's Next?

Ready to level up? Try these paths:

### 🚀 **Intermediate Projects**
- [Intermediate Projects](intermediate-projects.md) - Build larger applications
- [Advanced Projects](advanced-projects.md) - Complex multi-agent systems

### 📖 **Guides**
- [Building Chat Apps](../guides/building-chat-apps.md) - Complete chat applications
- [Custom Tools](../tools/custom-tools.md) - Advanced tool development
- [Agent Communication](../guides/agent-communication.md) - Multi-agent coordination

### 🔧 **Production Skills**  
- [Error Handling](../guides/error-handling.md) - Robust error management
- [Performance Tips](../guides/performance.md) - Optimization strategies
- [Testing Strategies](../advanced/testing-strategies.md) - Testing AI applications

---

**Feeling confident?** → [Try intermediate projects](intermediate-projects.md) | **Want to dive deeper?** → [Build chat applications](../guides/building-chat-apps.md)