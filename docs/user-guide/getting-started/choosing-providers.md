# Choosing Providers: Your AI Partner Selection Guide

> **[Project Root](/) / [Documentation](/docs/) / [User Guide](/docs/user-guide/) / [Getting Started](/docs/user-guide/getting-started/) / Choosing Providers**

Select the right LLM provider for your project with this comprehensive comparison guide. Each provider has unique strengths, costs, and use cases.

## Quick Provider Selector

**🚀 Just want to get started?** Use this quick selector:

- **New to AI?** → **OpenAI** (most reliable, best docs)
- **Privacy focused?** → **Ollama** (runs locally)
- **Cost sensitive?** → **OpenRouter** (68 free models)
- **Advanced reasoning?** → **Anthropic** (Claude is excellent for analysis)
- **Multimodal needs?** → **Google Gemini** (great vision and speed)
- **Enterprise/compliance?** → **Vertex AI** (Google Cloud integration)

## Provider Comparison Matrix

![Provider Comparison](../../images/provider-comparison.svg)

| Provider | Best For | Cost Range | Setup Difficulty | Model Quality | Special Features |
|----------|----------|------------|------------------|---------------|------------------|
| **OpenAI** | General use, reliability | $ - $$$ | Easy | Excellent | Function calling, vision, latest GPT models |
| **Anthropic** | Analysis, reasoning, safety | $$ - $$$ | Easy | Excellent | Long context, constitutional AI, latest Claude |
| **Google Gemini** | Speed, multimodal | $ - $$ | Easy | Very Good | Fast inference, integrated Google services |
| **Ollama** | Privacy, local hosting | Free* | Medium | Good | No API costs, full privacy, offline capable |
| **OpenRouter** | Model variety, cost | Free - $$$ | Easy | Varies | 400+ models, many free, single API |
| **Vertex AI** | Enterprise, compliance | $$ - $$$ | Hard | Excellent | Google Cloud, enterprise features, SLAs |

*Hardware costs apply for Ollama

## Detailed Provider Guides

### OpenAI - The Reliable Choice
*Best for: Beginners, production applications, function calling*

#### ✅ **Strengths**
- **Proven Reliability** - Battle-tested in production environments
- **Excellent Documentation** - Comprehensive guides and examples
- **Function Calling** - Best-in-class tool integration
- **Model Variety** - GPT-4o, GPT-4 Turbo, GPT-4o-mini for different budgets
- **Vision Capabilities** - Strong image understanding and generation
- **Strong Ecosystem** - Extensive third-party integrations

#### ⚠️ **Considerations**
- **Cost** - Can be expensive for high-volume applications
- **Privacy** - Data sent to OpenAI (though not used for training)
- **Rate Limits** - May hit limits on free tier

#### 🛠️ **Setup**
```bash
# Get API key from https://platform.openai.com/api-keys
export OPENAI_API_KEY="sk-..."

# Optional: Organization (for team accounts)
export OPENAI_ORGANIZATION="org-..."
```

#### 🎯 **Best Models**
- **GPT-4o** - Latest, multimodal, best quality
- **GPT-4o-mini** - Fast, cost-effective, good quality
- **GPT-4 Turbo** - High capability, longer context

#### 💰 **Cost Guide**
- **GPT-4o-mini**: $0.15/$0.60 per 1M tokens (input/output)
- **GPT-4o**: $2.50/$10.00 per 1M tokens (input/output)
- **GPT-4 Turbo**: $10.00/$30.00 per 1M tokens (input/output)

#### 📖 **Usage Example**
```go
provider := provider.NewOpenAIProvider(apiKey, "gpt-4o-mini")
agent := core.NewLLMAgent("assistant", "gpt-4o-mini", core.LLMDeps{Provider: provider})
```

### Anthropic - The Reasoning Expert
*Best for: Analysis, research, complex reasoning, safety-critical applications*

#### ✅ **Strengths**
- **Superior Reasoning** - Excellent for complex analysis and problem-solving
- **Long Context** - 200K+ token context windows
- **Constitutional AI** - Built-in safety and ethical guardrails
- **Honest Responses** - More likely to admit uncertainty
- **Latest Claude Models** - Claude 3.5 Sonnet is state-of-the-art
- **Research Focus** - Cutting-edge AI safety research

#### ⚠️ **Considerations**
- **Cost** - Premium pricing for premium quality
- **Speed** - Slightly slower than some competitors
- **Availability** - Limited regional availability

#### 🛠️ **Setup**
```bash
# Get API key from https://console.anthropic.com/
export ANTHROPIC_API_KEY="sk-ant-..."
```

#### 🎯 **Best Models**
- **Claude 3.5 Sonnet** - Latest, excellent reasoning
- **Claude 3.5 Haiku** - Fast, cost-effective
- **Claude 3 Opus** - Maximum capability (when available)

#### 💰 **Cost Guide**
- **Claude 3.5 Haiku**: $0.25/$1.25 per 1M tokens
- **Claude 3.5 Sonnet**: $3.00/$15.00 per 1M tokens
- **Claude 3 Opus**: $15.00/$75.00 per 1M tokens

#### 📖 **Usage Example**
```go
provider := provider.NewAnthropicProvider(apiKey, "claude-3-5-sonnet-latest")
agent := core.NewLLMAgent("analyst", "claude-3-5-sonnet-latest", core.LLMDeps{Provider: provider})
```

### Google Gemini - The Speed Demon
*Best for: Fast responses, multimodal content, Google ecosystem integration*

#### ✅ **Strengths**
- **Blazing Fast** - Some of the fastest inference speeds
- **Multimodal Native** - Excellent vision and understanding
- **Google Integration** - Works with Google services
- **Cost Effective** - Competitive pricing
- **Gemini 2.0** - Latest models with strong capabilities
- **Free Tier** - Generous free usage limits

#### ⚠️ **Considerations**
- **Newer** - Less battle-tested than OpenAI/Anthropic
- **Function Calling** - Still evolving, not as mature
- **Context Length** - Shorter than Anthropic

#### 🛠️ **Setup**
```bash
# Get API key from https://makersuite.google.com/app/apikey
export GEMINI_API_KEY="AI..."
```

#### 🎯 **Best Models**
- **Gemini 2.0 Flash Lite** - Latest, balanced performance
- **Gemini 1.5 Pro** - High capability, longer context
- **Gemini 1.5 Flash** - Fast, cost-effective

#### 💰 **Cost Guide**
- **Gemini 2.0 Flash**: $0.075/$0.30 per 1M tokens
- **Gemini 1.5 Pro**: $1.25/$5.00 per 1M tokens
- **Gemini 1.5 Flash**: $0.075/$0.30 per 1M tokens

#### 📖 **Usage Example**
```go
provider := provider.NewGeminiProvider(apiKey, "gemini-2.0-flash-latest")
agent := core.NewLLMAgent("assistant", "gemini-2.0-flash-latest", core.LLMDeps{Provider: provider})
```

### Ollama - The Privacy Champion
*Best for: Local development, privacy, cost control, offline usage*

#### ✅ **Strengths**
- **Complete Privacy** - Everything runs locally
- **No API Costs** - Only hardware and electricity
- **Offline Capable** - Works without internet
- **Model Variety** - Llama, Mistral, CodeLlama, and more
- **Easy Setup** - Simple installation and management
- **Development Friendly** - Perfect for testing and development

#### ⚠️ **Considerations**
- **Hardware Requirements** - Needs GPU for good performance
- **Model Quality** - Generally lower than commercial APIs
- **Setup Complexity** - Requires local installation
- **No Support** - Community-driven support only

#### 🛠️ **Setup**
```bash
# Install Ollama
curl -fsSL https://ollama.ai/install.sh | sh

# Pull models
ollama pull llama3.2:3b      # 3B parameter model
ollama pull codellama:7b     # Code-focused model
ollama pull mistral:7b       # General purpose

# Set host (optional)
export OLLAMA_HOST="http://localhost:11434"
```

#### 🎯 **Best Models**
- **llama3.2:3b** - Fast, good for development
- **codellama:7b** - Excellent for code tasks
- **mistral:7b** - Good general purpose model
- **llama3.1:8b** - Balanced performance/quality

#### 💰 **Cost Guide**
- **Free to run** (after hardware investment)
- **Hardware needs**: 8GB+ RAM for 3B models, 16GB+ for 7B models
- **GPU recommended**: NVIDIA with 8GB+ VRAM for good performance

#### 📖 **Usage Example**
```go
provider := provider.NewOllamaProvider("http://localhost:11434", "llama3.2:3b")
agent := core.NewLLMAgent("local-assistant", "llama3.2:3b", core.LLMDeps{Provider: provider})
```

### OpenRouter - The Model Marketplace
*Best for: Exploring different models, cost optimization, accessing specialized models*

#### ✅ **Strengths**
- **Huge Model Selection** - 400+ models from different providers
- **Many Free Models** - 68 models with free tiers
- **Single API** - Access multiple providers through one interface
- **Cost Comparison** - Easy to compare model costs
- **Latest Models** - Quick access to new releases
- **Flexible Billing** - Pay-per-use with detailed tracking

#### ⚠️ **Considerations**
- **Quality Variance** - Models vary significantly in capability
- **Rate Limits** - Free models have usage limits
- **Support** - Varying levels of model-specific support
- **Latency** - Additional network hop may increase response time

#### 🛠️ **Setup**
```bash
# Get API key from https://openrouter.ai/keys
export OPENROUTER_API_KEY="sk-or-..."

# Optional: Site URL for usage tracking
export OPENROUTER_SITE_URL="https://yoursite.com"
```

#### 🎯 **Best Models**
- **Free Tier**: google/gemma-2-9b-it, mistralai/mistral-7b-instruct
- **Premium**: anthropic/claude-3.5-sonnet, openai/gpt-4o
- **Specialized**: coding, vision, or domain-specific models

#### 💰 **Cost Guide**
- **Free models**: $0 (with usage limits)
- **Budget models**: $0.10-$1.00 per 1M tokens
- **Premium models**: Match or slightly above original provider pricing

#### 📖 **Usage Example**
```go
provider := provider.NewOpenRouterProvider(apiKey, "anthropic/claude-3.5-sonnet")
agent := core.NewLLMAgent("assistant", "claude-3.5-sonnet", core.LLMDeps{Provider: provider})
```

### Vertex AI - The Enterprise Option
*Best for: Enterprise deployments, Google Cloud integration, compliance requirements*

#### ✅ **Strengths**
- **Enterprise Ready** - SLAs, compliance, enterprise support
- **Google Cloud Integration** - Seamless GCP integration
- **Security** - Advanced security and compliance features
- **Scaling** - Enterprise-grade scaling and management
- **Model Garden** - Access to multiple model families
- **Private Deployment** - Option for private model hosting

#### ⚠️ **Considerations**
- **Complexity** - Most complex setup process
- **Cost** - Can be expensive with GCP overhead
- **Learning Curve** - Requires GCP knowledge
- **Overkill** - May be excessive for simple applications

#### 🛠️ **Setup**
```bash
# Install gcloud CLI and authenticate
gcloud auth application-default login

# Or use service account
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"

# Set project
export GOOGLE_CLOUD_PROJECT="your-project-id"
```

#### 🎯 **Best Models**
- **Gemini Pro** - Google's latest models
- **Claude** - Anthropic models via Vertex AI
- **Code models** - Specialized coding models

#### 💰 **Cost Guide**
- **Pricing**: Similar to Google AI Studio but with GCP billing
- **Additional costs**: GCP infrastructure and management

#### 📖 **Usage Example**
```go
provider := provider.NewVertexAIProvider("your-project-id", "us-central1", "gemini-pro")
agent := core.NewLLMAgent("enterprise-agent", "gemini-pro", core.LLMDeps{Provider: provider})
```

## Decision Framework

### By Use Case

#### 🚀 **Getting Started / Learning**
1. **OpenAI** - Most tutorials and examples
2. **Ollama** - No API costs for experimentation
3. **OpenRouter** - Try many models easily

#### 🏢 **Production Applications**
1. **OpenAI** - Proven reliability
2. **Anthropic** - For reasoning-heavy tasks
3. **Vertex AI** - For enterprise requirements

#### 💰 **Cost-Sensitive Projects**
1. **Ollama** - Free after hardware investment
2. **OpenRouter** - Many free models
3. **Gemini** - Competitive pricing

#### 🔒 **Privacy-Sensitive Applications**
1. **Ollama** - Complete local control
2. **Vertex AI** - Enterprise privacy controls
3. **On-premises deployments** - Through enterprise agreements

#### ⚡ **Speed-Critical Applications**
1. **Gemini** - Fastest inference
2. **OpenAI GPT-4o-mini** - Good speed/quality balance
3. **Local Ollama** - No network latency

### By Technical Requirements

#### 🛠️ **Function Calling / Tool Use**
1. **OpenAI** - Most mature implementation
2. **Anthropic** - Good tool support
3. **Gemini** - Evolving capabilities

#### 👀 **Vision / Multimodal**
1. **Gemini** - Native multimodal
2. **OpenAI GPT-4o** - Excellent vision
3. **Claude 3** - Good vision capabilities

#### 📄 **Long Context**
1. **Anthropic** - 200K+ tokens
2. **Gemini** - Up to 2M tokens (some models)
3. **OpenAI** - 128K tokens

#### 🧠 **Complex Reasoning**
1. **Anthropic Claude** - Superior reasoning
2. **OpenAI GPT-4o** - Strong analytical capabilities
3. **Gemini Pro** - Good reasoning performance

## Multi-Provider Strategies

### Provider Fallback Chain
```go
// Try premium provider first, fallback to alternatives
agent, err := core.NewAgentFromString("assistant", "anthropic/claude-3.5-sonnet")
if err != nil {
    agent, err = core.NewAgentFromString("assistant", "openai/gpt-4o")
    if err != nil {
        agent, _ = core.NewAgentFromString("assistant", "ollama/llama3.2:3b")
    }
}
```

### Cost Optimization
```go
// Use cheaper models for simple tasks, premium for complex ones
func selectProvider(taskComplexity string) string {
    switch taskComplexity {
    case "simple":
        return "openai/gpt-4o-mini"  // Fast and cheap
    case "complex":
        return "anthropic/claude-3.5-sonnet"  // Best reasoning
    case "multimodal":
        return "gemini/gemini-2.0-flash"  // Best vision
    default:
        return "openai/gpt-4o"  // Balanced default
    }
}
```

### Development vs Production
```go
// Development: Use local/free models
if os.Getenv("ENV") == "development" {
    provider = "ollama/llama3.2:3b"
} else {
    // Production: Use reliable commercial APIs
    provider = "openai/gpt-4o"
}
```

## Getting Started Recommendations

### For Different Experience Levels

#### 🌱 **Complete Beginner**
1. Start with **OpenAI** (best documentation)
2. Try **Ollama** for cost-free experimentation
3. Use GPT-4o-mini for learning (fast and cheap)

#### 🚀 **Experienced Developer** 
1. Start with **Anthropic** for challenging tasks
2. Use **OpenRouter** to explore model variety
3. Set up **Ollama** for local development

#### 🏢 **Enterprise Team**
1. Evaluate **Vertex AI** for compliance needs
2. Test **OpenAI** for reliability requirements
3. Consider **Anthropic** for safety-critical applications

### Quick Start Commands

```bash
# OpenAI (recommended first choice)
export OPENAI_API_KEY="your-key"
go run first_ai_app.go

# Anthropic (for reasoning tasks)
export ANTHROPIC_API_KEY="your-key"
go run reasoning_app.go

# Ollama (for local development)
ollama pull llama3.2:3b
go run local_ai_app.go

# OpenRouter (for exploration)
export OPENROUTER_API_KEY="your-key"
go run multi_model_app.go
```

## Next Steps

🎯 **Ready to choose?** Here's what to do next:

1. **[Set up your chosen provider](installation.md#environment-configuration)**
2. **[Build your first app](first-steps.md)**
3. **[Explore provider-specific guides](../guides/provider-setup.md)**

### Learn More

- **[Provider Setup Guide](../guides/provider-setup.md)** - Detailed configuration for each provider
- **[Multi-Provider Strategies](../guides/multi-provider-strategies.md)** - Using multiple providers together
- **[Provider Comparison](../reference/provider-comparison.md)** - Complete feature matrix
- **[Local Providers](../guides/local-providers.md)** - Deep dive into Ollama and local hosting

---

**Still unsure?** Start with **OpenAI** for reliability or **Ollama** for cost-free experimentation. You can always switch later! 🚀