# Google Vertex AI Provider Example

This example demonstrates how to use the Google Vertex AI provider to access Gemini and partner models (Claude) through Google Cloud's enterprise AI platform.

## Features Demonstrated

1. **Basic Generation** - Using Application Default Credentials (ADC)
2. **Service Account Authentication** - Explicit service account configuration
3. **Streaming** - Real-time streaming responses
4. **Partner Models** - Using Claude models through Vertex AI
5. **Multimodal Input** - Sending images to Gemini models

## Prerequisites

1. Google Cloud Project with Vertex AI API enabled
2. Authentication configured (one of):
   - Application Default Credentials (ADC)
   - Service Account JSON file
3. Appropriate IAM permissions (`aiplatform.endpoints.predict` or `Vertex AI Service Agent` role)

## Environment Variables

```bash
# Required
export VERTEX_AI_PROJECT_ID="your-gcp-project-id"

# Optional (defaults shown)
export VERTEX_AI_LOCATION="us-central1"              # Region for Vertex AI
export VERTEX_AI_MODEL="gemini-1.5-flash-001"        # Model to use

# For service account authentication (optional)
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service-account.json"
```

## Authentication Setup

### Option 1: Application Default Credentials (ADC)

If running on Google Cloud (GCE, GKE, Cloud Run, etc.), ADC works automatically.

For local development:
```bash
# Authenticate with your Google account
gcloud auth application-default login

# Set the project
gcloud config set project YOUR_PROJECT_ID
```

### Option 2: Service Account

1. Create a service account:
```bash
gcloud iam service-accounts create vertex-ai-demo \
    --display-name="Vertex AI Demo Service Account"
```

2. Grant permissions:
```bash
gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
    --member="serviceAccount:vertex-ai-demo@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
    --role="roles/aiplatform.user"
```

3. Create and download key:
```bash
gcloud iam service-accounts keys create vertex-ai-key.json \
    --iam-account=vertex-ai-demo@YOUR_PROJECT_ID.iam.gserviceaccount.com
```

4. Set environment variable:
```bash
export GOOGLE_APPLICATION_CREDENTIALS="$(pwd)/vertex-ai-key.json"
```

## Running the Example

```bash
# With ADC
export VERTEX_AI_PROJECT_ID="your-project-id"
go run cmd/examples/provider-vertexai/main.go

# With specific region and model
export VERTEX_AI_PROJECT_ID="your-project-id"
export VERTEX_AI_LOCATION="europe-west4"
export VERTEX_AI_MODEL="gemini-1.5-pro-001"
go run cmd/examples/provider-vertexai/main.go

# Using Claude through Vertex AI
export VERTEX_AI_PROJECT_ID="your-project-id"
export VERTEX_AI_LOCATION="us-central1"  # Claude available in limited regions
go run cmd/examples/provider-vertexai/main.go
```

## Available Models

### Google Models
- `gemini-2.0-flash-preview-04-15` - Latest Gemini 2.0 Flash preview
- `gemini-1.5-pro-001` - Most capable model for complex tasks
- `gemini-1.5-flash-001` - Fast and efficient for high-volume tasks

### Partner Models (Claude via Vertex AI)
- `claude-3-opus@20240229` - Most capable Claude model
- `claude-3-7-sonnet@20241022` - Latest Claude Sonnet
- `claude-3-5-sonnet@20240620` - Balanced Claude model
- `claude-3-5-haiku@20241022` - Fast and efficient Claude model

Note: Partner models may not be available in all regions.

## Regional Availability

Vertex AI is available in multiple regions. Common ones include:
- `us-central1` (Iowa)
- `us-east1` (South Carolina)
- `us-west1` (Oregon)
- `europe-west1` (Belgium)
- `europe-west4` (Netherlands)
- `asia-east1` (Taiwan)
- `asia-northeast1` (Tokyo)

For the latest list, see: https://cloud.google.com/vertex-ai/docs/general/locations

## Cost Considerations

- Vertex AI charges based on model usage (per 1K characters or tokens)
- Pricing varies by model and region
- Partner models (Claude) have different pricing than Google models
- No charge for authentication or API calls that result in errors

See current pricing: https://cloud.google.com/vertex-ai/pricing

## Troubleshooting

### Authentication Errors
```
Error: failed to find default credentials
```
Solution: Run `gcloud auth application-default login` or set `GOOGLE_APPLICATION_CREDENTIALS`

### Permission Errors
```
Error: Permission 'aiplatform.endpoints.predict' denied
```
Solution: Ensure your service account or user has the `Vertex AI User` role

### Region Errors
```
Error: Model not available in region
```
Solution: Check model availability in your chosen region or try a different region

### Quota Errors
```
Error: Quota exceeded
```
Solution: Check your project quotas in the Google Cloud Console

## Example Output

```
=== Example 1: Basic Generation with ADC ===
Response from gemini-1.5-flash-001:
Here are 3 key benefits of using Vertex AI for enterprise applications:

1. **Unified Platform**: Consolidates ML workflows from data preparation to deployment
2. **Enterprise Security**: Built-in governance, IAM integration, and VPC support
3. **Scalability**: Automatic scaling and managed infrastructure for production workloads

=== Example 2: Generation with Service Account ===
Response: Machine learning is a subset of AI that enables systems to learn and improve from experience without being explicitly programmed.

=== Example 3: Streaming Generation ===
Streaming response: Cloud computing high above,
Data flows like morning mist bright,
Innovation soars.

=== Example 4: Partner Models (Claude) ===
Response from Claude via Vertex AI:
Claude is designed with a focus on being helpful, harmless, and honest...

=== Example 5: Multimodal Input ===
Multimodal response: The image shows a single red pixel.
```