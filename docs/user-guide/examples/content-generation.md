# Content Generation: Content Creation and Management

> **[Project Root](/) / [Documentation](../..) / [User Guide](../../user-guide) / [Examples](../../user-guide/examples) / Content Generation**

Build a comprehensive AI-powered content creation and management system. This example demonstrates automated content generation, SEO optimization, multi-format publishing, content scheduling, and quality assurance workflows using Go-LLMs.

## System Overview

This content generation system provides:

- **Multi-Format Content Creation** - Articles, social media posts, emails, documentation
- **SEO Optimization** - Keyword research, meta descriptions, content scoring
- **Brand Voice Consistency** - Tone analysis and brand guideline enforcement
- **Content Planning** - Editorial calendars and topic generation
- **Quality Assurance** - Automated proofreading and fact-checking
- **Multi-Channel Publishing** - Blog, social media, email campaigns
- **Performance Analytics** - Content engagement and conversion tracking

## Architecture

![Content Generation System Architecture](../../images/content-generation-architecture.svg)

### Components
1. **Content Planner** - Generates content ideas and editorial calendars
2. **Content Creator** - Produces written content in various formats
3. **SEO Optimizer** - Enhances content for search engine visibility
4. **Quality Checker** - Reviews content for grammar, tone, and accuracy
5. **Brand Guardian** - Ensures brand voice and guideline compliance
6. **Publisher** - Manages content distribution across channels
7. **Analytics Engine** - Tracks content performance and engagement

---

## Complete Implementation

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strconv"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    
    "github.com/lexlapax/go-llms/pkg/agent/core"
    "github.com/lexlapax/go-llms/pkg/agent/domain"
    "github.com/lexlapax/go-llms/pkg/agent/builtins/tools/web"
    "github.com/lexlapax/go-llms/pkg/agent/workflow"
    "github.com/lexlapax/go-llms/pkg/llm/provider"
    "github.com/lexlapax/go-llms/pkg/structured/processor"
)

// ContentGenerationSystem is the main system orchestrator
type ContentGenerationSystem struct {
    db                *sqlx.DB
    plannerAgent      *core.LLMAgent
    creatorAgent      *core.LLMAgent
    seoAgent          *core.LLMAgent
    qualityAgent      *core.LLMAgent
    brandAgent        *core.LLMAgent
    researchAgent     *core.LLMAgent
    workflowAgent     *workflow.SequentialAgent
    brandGuidelines   *BrandGuidelines
    seoConfig         *SEOConfig
    publishingChannels map[string]PublishingChannel
    config            *ContentConfig
    metrics           *ContentMetrics
}

type ContentConfig struct {
    DatabaseURL       string        `json:"database_url"`
    OpenAIKey        string        `json:"openai_key"`
    BrandName        string        `json:"brand_name"`
    Industry         string        `json:"industry"`
    TargetAudience   string        `json:"target_audience"`
    DefaultTone      string        `json:"default_tone"`
    ContentGuidelines string       `json:"content_guidelines"`
    MaxContentLength int           `json:"max_content_length"`
    SEOEnabled       bool          `json:"seo_enabled"`
    QualityThreshold float64       `json:"quality_threshold"`
}

// Content models
type ContentPiece struct {
    ID              int                    `json:"id" db:"id"`
    Title           string                 `json:"title" db:"title"`
    Content         string                 `json:"content" db:"content"`
    ContentType     string                 `json:"content_type" db:"content_type"` // article, social, email, documentation
    Status          string                 `json:"status" db:"status"` // draft, review, approved, published, archived
    AuthorID        string                 `json:"author_id" db:"author_id"`
    TopicID         *int                   `json:"topic_id" db:"topic_id"`
    Keywords        []string               `json:"keywords" db:"keywords"`
    Tags            []string               `json:"tags" db:"tags"`
    MetaDescription string                 `json:"meta_description" db:"meta_description"`
    SEOScore        float64                `json:"seo_score" db:"seo_score"`
    QualityScore    float64                `json:"quality_score" db:"quality_score"`
    BrandScore      float64                `json:"brand_score" db:"brand_score"`
    WordCount       int                    `json:"word_count" db:"word_count"`
    ReadingTime     int                    `json:"reading_time" db:"reading_time"` // minutes
    ScheduledFor    *time.Time             `json:"scheduled_for" db:"scheduled_for"`
    PublishedAt     *time.Time             `json:"published_at" db:"published_at"`
    CreatedAt       time.Time              `json:"created_at" db:"created_at"`
    UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
    Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
}

type ContentTopic struct {
    ID              int       `json:"id" db:"id"`
    Name            string    `json:"name" db:"name"`
    Description     string    `json:"description" db:"description"`
    Category        string    `json:"category" db:"category"`
    Priority        string    `json:"priority" db:"priority"` // low, medium, high, urgent
    Keywords        []string  `json:"keywords" db:"keywords"`
    TargetAudience  string    `json:"target_audience" db:"target_audience"`
    SearchVolume    int       `json:"search_volume" db:"search_volume"`
    Competition     float64   `json:"competition" db:"competition"`
    TrendScore      float64   `json:"trend_score" db:"trend_score"`
    LastUsed        *time.Time `json:"last_used" db:"last_used"`
    CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

type EditorialCalendar struct {
    ID          int        `json:"id" db:"id"`
    Date        time.Time  `json:"date" db:"date"`
    ContentType string     `json:"content_type" db:"content_type"`
    TopicID     int        `json:"topic_id" db:"topic_id"`
    Title       string     `json:"title" db:"title"`
    Status      string     `json:"status" db:"status"` // planned, assigned, in_progress, completed
    AssignedTo  *string    `json:"assigned_to" db:"assigned_to"`
    Channel     string     `json:"channel" db:"channel"`
    Priority    string     `json:"priority" db:"priority"`
    Notes       string     `json:"notes" db:"notes"`
    CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}

type BrandGuidelines struct {
    BrandName       string   `json:"brand_name"`
    ToneOfVoice     string   `json:"tone_of_voice"`
    WritingStyle    string   `json:"writing_style"`
    KeyMessages     []string `json:"key_messages"`
    ForbiddenWords  []string `json:"forbidden_words"`
    PreferredWords  []string `json:"preferred_words"`
    TargetAudience  string   `json:"target_audience"`
    BrandValues     []string `json:"brand_values"`
    DoList          []string `json:"do_list"`
    DontList        []string `json:"dont_list"`
}

type SEOConfig struct {
    PrimaryKeywords     []string `json:"primary_keywords"`
    SecondaryKeywords   []string `json:"secondary_keywords"`
    TargetKeywordDensity float64  `json:"target_keyword_density"`
    MetaDescriptionLength int     `json:"meta_description_length"`
    TitleLength         int      `json:"title_length"`
    HeadingStructure    bool     `json:"heading_structure"`
    InternalLinking     bool     `json:"internal_linking"`
    ImageAltText        bool     `json:"image_alt_text"`
}

// Content analysis results
type ContentAnalysis struct {
    SEOAnalysis    SEOAnalysis    `json:"seo_analysis"`
    QualityAnalysis QualityAnalysis `json:"quality_analysis"`
    BrandAnalysis  BrandAnalysis  `json:"brand_analysis"`
    OverallScore   float64        `json:"overall_score"`
    Recommendations []string      `json:"recommendations"`
    Issues         []string       `json:"issues"`
}

type SEOAnalysis struct {
    Score               float64  `json:"score"`
    KeywordDensity      float64  `json:"keyword_density"`
    TitleOptimization   float64  `json:"title_optimization"`
    MetaDescription     float64  `json:"meta_description"`
    HeadingStructure    float64  `json:"heading_structure"`
    ContentLength       float64  `json:"content_length"`
    ReadabilityScore    float64  `json:"readability_score"`
    Recommendations     []string `json:"recommendations"`
}

type QualityAnalysis struct {
    Score           float64  `json:"score"`
    GrammarErrors   int      `json:"grammar_errors"`
    SpellingErrors  int      `json:"spelling_errors"`
    ClarityScore    float64  `json:"clarity_score"`
    EngagementScore float64  `json:"engagement_score"`
    FactualAccuracy float64  `json:"factual_accuracy"`
    Issues          []string `json:"issues"`
    Suggestions     []string `json:"suggestions"`
}

type BrandAnalysis struct {
    Score           float64  `json:"score"`
    ToneConsistency float64  `json:"tone_consistency"`
    MessageAlignment float64 `json:"message_alignment"`
    StyleCompliance float64  `json:"style_compliance"`
    VoiceMatch      float64  `json:"voice_match"`
    Issues          []string `json:"issues"`
    Suggestions     []string `json:"suggestions"`
}

// Publishing channels
type PublishingChannel interface {
    Name() string
    Publish(ctx context.Context, content *ContentPiece) error
    Schedule(ctx context.Context, content *ContentPiece, publishTime time.Time) error
    GetAnalytics(ctx context.Context, contentID int) (*ChannelAnalytics, error)
}

type ChannelAnalytics struct {
    ContentID   int                    `json:"content_id"`
    Channel     string                 `json:"channel"`
    Views       int64                  `json:"views"`
    Engagement  int64                  `json:"engagement"`
    Shares      int64                  `json:"shares"`
    Comments    int64                  `json:"comments"`
    Conversions int64                  `json:"conversions"`
    CTR         float64                `json:"ctr"`
    Metrics     map[string]interface{} `json:"metrics"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

type ContentMetrics struct {
    ContentCreated       prometheus.Counter
    ContentPublished     prometheus.Counter
    SEOScoreDistribution prometheus.Histogram
    QualityScoreDistribution prometheus.Histogram
    GenerationTime       prometheus.Histogram
    ChannelPerformance   *prometheus.CounterVec
}

func NewContentGenerationSystem(config *ContentConfig) (*ContentGenerationSystem, error) {
    // Initialize database
    db, err := sqlx.Connect("postgres", config.DatabaseURL)
    if err != nil {
        return nil, fmt.Errorf("database connection failed: %w", err)
    }

    // Create LLM provider
    llm, err := provider.NewOpenAIProvider(
        provider.WithModel("gpt-4"),
        provider.WithMaxTokens(3000),
    )
    if err != nil {
        return nil, fmt.Errorf("LLM provider creation failed: %w", err)
    }

    // Create specialized agents
    plannerAgent := core.NewLLMAgent("content-planner", llm)
    creatorAgent := core.NewLLMAgent("content-creator", llm)
    seoAgent := core.NewLLMAgent("seo-optimizer", llm)
    qualityAgent := core.NewLLMAgent("quality-checker", llm)
    brandAgent := core.NewLLMAgent("brand-guardian", llm)
    researchAgent := core.NewLLMAgent("research-agent", llm)

    // Add web research tool to research agent
    webTool := web.NewWebFetchTool()
    researchAgent.AddTool(webTool)

    // Create workflow agent
    workflowAgent := workflow.NewSequentialAgent("content-workflow").
        AddAgent(plannerAgent).
        AddAgent(researchAgent).
        AddAgent(creatorAgent).
        AddAgent(seoAgent).
        AddAgent(qualityAgent).
        AddAgent(brandAgent)

    // Initialize brand guidelines
    brandGuidelines := &BrandGuidelines{
        BrandName:    config.BrandName,
        ToneOfVoice:  config.DefaultTone,
        WritingStyle: "Professional, clear, and engaging",
        KeyMessages:  []string{"Innovation", "Quality", "Customer-focused"},
        TargetAudience: config.TargetAudience,
    }

    // Initialize SEO config
    seoConfig := &SEOConfig{
        TargetKeywordDensity:  0.02, // 2%
        MetaDescriptionLength: 160,
        TitleLength:          60,
        HeadingStructure:     true,
        InternalLinking:      true,
        ImageAltText:        true,
    }

    // Initialize publishing channels
    publishingChannels := make(map[string]PublishingChannel)
    publishingChannels["blog"] = &BlogChannel{}
    publishingChannels["social"] = &SocialMediaChannel{}
    publishingChannels["email"] = &EmailChannel{}

    system := &ContentGenerationSystem{
        db:                 db,
        plannerAgent:       plannerAgent,
        creatorAgent:       creatorAgent,
        seoAgent:          seoAgent,
        qualityAgent:      qualityAgent,
        brandAgent:        brandAgent,
        researchAgent:     researchAgent,
        workflowAgent:     workflowAgent,
        brandGuidelines:   brandGuidelines,
        seoConfig:         seoConfig,
        publishingChannels: publishingChannels,
        config:            config,
        metrics:           initializeContentMetrics(),
    }

    // Initialize database schema
    if err := system.initializeSchema(); err != nil {
        return nil, fmt.Errorf("schema initialization failed: %w", err)
    }

    return system, nil
}

func (cgs *ContentGenerationSystem) initializeSchema() error {
    schema := `
    CREATE TABLE IF NOT EXISTS content_topics (
        id SERIAL PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        description TEXT,
        category VARCHAR(100),
        priority VARCHAR(20) DEFAULT 'medium',
        keywords TEXT[],
        target_audience TEXT,
        search_volume INTEGER DEFAULT 0,
        competition DECIMAL(3,2) DEFAULT 0.0,
        trend_score DECIMAL(3,2) DEFAULT 0.0,
        last_used TIMESTAMP,
        created_at TIMESTAMP DEFAULT NOW()
    );

    CREATE TABLE IF NOT EXISTS content_pieces (
        id SERIAL PRIMARY KEY,
        title TEXT NOT NULL,
        content TEXT NOT NULL,
        content_type VARCHAR(50) NOT NULL,
        status VARCHAR(50) DEFAULT 'draft',
        author_id VARCHAR(255),
        topic_id INTEGER REFERENCES content_topics(id),
        keywords TEXT[],
        tags TEXT[],
        meta_description TEXT,
        seo_score DECIMAL(3,2) DEFAULT 0.0,
        quality_score DECIMAL(3,2) DEFAULT 0.0,
        brand_score DECIMAL(3,2) DEFAULT 0.0,
        word_count INTEGER DEFAULT 0,
        reading_time INTEGER DEFAULT 0,
        scheduled_for TIMESTAMP,
        published_at TIMESTAMP,
        created_at TIMESTAMP DEFAULT NOW(),
        updated_at TIMESTAMP DEFAULT NOW(),
        metadata JSONB
    );

    CREATE TABLE IF NOT EXISTS editorial_calendar (
        id SERIAL PRIMARY KEY,
        date DATE NOT NULL,
        content_type VARCHAR(50) NOT NULL,
        topic_id INTEGER REFERENCES content_topics(id),
        title TEXT,
        status VARCHAR(50) DEFAULT 'planned',
        assigned_to VARCHAR(255),
        channel VARCHAR(50),
        priority VARCHAR(20) DEFAULT 'medium',
        notes TEXT,
        created_at TIMESTAMP DEFAULT NOW()
    );

    CREATE TABLE IF NOT EXISTS content_analytics (
        id SERIAL PRIMARY KEY,
        content_id INTEGER REFERENCES content_pieces(id),
        channel VARCHAR(50),
        views BIGINT DEFAULT 0,
        engagement BIGINT DEFAULT 0,
        shares BIGINT DEFAULT 0,
        comments BIGINT DEFAULT 0,
        conversions BIGINT DEFAULT 0,
        ctr DECIMAL(5,4) DEFAULT 0.0,
        metrics JSONB,
        updated_at TIMESTAMP DEFAULT NOW()
    );

    CREATE INDEX IF NOT EXISTS idx_content_status ON content_pieces(status);
    CREATE INDEX IF NOT EXISTS idx_content_type ON content_pieces(content_type);
    CREATE INDEX IF NOT EXISTS idx_calendar_date ON editorial_calendar(date);
    CREATE INDEX IF NOT EXISTS idx_analytics_content ON content_analytics(content_id, channel);
    `

    _, err := cgs.db.Exec(schema)
    return err
}

// Core content generation workflow
func (cgs *ContentGenerationSystem) GenerateContent(ctx context.Context, request ContentRequest) (*ContentPiece, error) {
    start := time.Now()
    cgs.metrics.ContentCreated.Inc()

    // Step 1: Research the topic
    research, err := cgs.researchTopic(ctx, request.Topic, request.Keywords)
    if err != nil {
        log.Printf("Research failed: %v", err)
        research = &TopicResearch{
            Topic:       request.Topic,
            KeyPoints:   []string{},
            Sources:     []string{},
            TrendData:   map[string]interface{}{},
        }
    }

    // Step 2: Create content outline
    outline, err := cgs.createOutline(ctx, request, research)
    if err != nil {
        return nil, fmt.Errorf("outline creation failed: %w", err)
    }

    // Step 3: Generate content
    content, err := cgs.generateContent(ctx, request, outline, research)
    if err != nil {
        return nil, fmt.Errorf("content generation failed: %w", err)
    }

    // Step 4: Optimize for SEO
    if cgs.config.SEOEnabled {
        seoContent, err := cgs.optimizeForSEO(ctx, content, request.Keywords)
        if err != nil {
            log.Printf("SEO optimization failed: %v", err)
        } else {
            content = seoContent
        }
    }

    // Step 5: Check quality
    qualityAnalysis, err := cgs.checkQuality(ctx, content)
    if err != nil {
        log.Printf("Quality check failed: %v", err)
        qualityAnalysis = &QualityAnalysis{Score: 0.5}
    }

    // Step 6: Validate brand compliance
    brandAnalysis, err := cgs.validateBrand(ctx, content)
    if err != nil {
        log.Printf("Brand validation failed: %v", err)
        brandAnalysis = &BrandAnalysis{Score: 0.5}
    }

    // Step 7: Calculate metrics
    wordCount := len(strings.Fields(content))
    readingTime := calculateReadingTime(wordCount)

    // Step 8: Create content piece
    contentPiece := &ContentPiece{
        Title:           request.Title,
        Content:         content,
        ContentType:     request.ContentType,
        Status:          "draft",
        AuthorID:        request.AuthorID,
        Keywords:        request.Keywords,
        Tags:            request.Tags,
        QualityScore:    qualityAnalysis.Score,
        BrandScore:      brandAnalysis.Score,
        WordCount:       wordCount,
        ReadingTime:     readingTime,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
        Metadata:        request.Metadata,
    }

    // Step 9: Store in database
    if err := cgs.storeContent(ctx, contentPiece); err != nil {
        return nil, fmt.Errorf("content storage failed: %w", err)
    }

    // Record metrics
    cgs.metrics.GenerationTime.Observe(time.Since(start).Seconds())
    cgs.metrics.QualityScoreDistribution.Observe(qualityAnalysis.Score)

    return contentPiece, nil
}

type ContentRequest struct {
    Title       string                 `json:"title" binding:"required"`
    Topic       string                 `json:"topic" binding:"required"`
    ContentType string                 `json:"content_type" binding:"required"`
    Keywords    []string               `json:"keywords"`
    Tags        []string               `json:"tags"`
    AuthorID    string                 `json:"author_id"`
    TargetLength int                   `json:"target_length"`
    Tone        string                 `json:"tone"`
    Audience    string                 `json:"audience"`
    Channel     string                 `json:"channel"`
    Metadata    map[string]interface{} `json:"metadata"`
}

type TopicResearch struct {
    Topic       string                 `json:"topic"`
    KeyPoints   []string               `json:"key_points"`
    Sources     []string               `json:"sources"`
    TrendData   map[string]interface{} `json:"trend_data"`
    Competitors []string               `json:"competitors"`
    Keywords    []string               `json:"keywords"`
    Statistics  []string               `json:"statistics"`
}

func (cgs *ContentGenerationSystem) researchTopic(ctx context.Context, topic string, keywords []string) (*TopicResearch, error) {
    prompt := fmt.Sprintf(`Research the topic "%s" and provide comprehensive information for content creation.

Keywords to focus on: %s

Please provide:
1. Key points to cover in the content
2. Current trends and statistics
3. Relevant sources and references
4. Additional keywords and related topics
5. Competitive landscape insights

Return your research as JSON with the following structure:
{
    "topic": "%s",
    "key_points": ["point1", "point2", ...],
    "sources": ["source1", "source2", ...],
    "trend_data": {"trend1": "data1", ...},
    "competitors": ["comp1", "comp2", ...],
    "keywords": ["keyword1", "keyword2", ...],
    "statistics": ["stat1", "stat2", ...]
}`,
        topic, strings.Join(keywords, ", "), topic)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := cgs.researchAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var research TopicResearch
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        if err := json.Unmarshal([]byte(response), &research); err != nil {
            return nil, fmt.Errorf("failed to parse research: %w", err)
        }
    }

    return &research, nil
}

func (cgs *ContentGenerationSystem) createOutline(ctx context.Context, request ContentRequest, research *TopicResearch) (*ContentOutline, error) {
    prompt := fmt.Sprintf(`Create a detailed content outline based on the following information:

Title: %s
Topic: %s
Content Type: %s
Target Length: %d words
Tone: %s
Target Audience: %s
Channel: %s

Research Data:
Key Points: %s
Keywords: %s

Brand Guidelines:
- Brand Name: %s
- Tone of Voice: %s
- Target Audience: %s

Please create a structured outline that includes:
1. Introduction hook
2. Main sections with subsections
3. Key points to cover in each section
4. Conclusion elements
5. Call-to-action suggestions

Return the outline as JSON with proper structure for content creation.`,
        request.Title, request.Topic, request.ContentType, request.TargetLength,
        request.Tone, request.Audience, request.Channel,
        strings.Join(research.KeyPoints, ", "), strings.Join(request.Keywords, ", "),
        cgs.brandGuidelines.BrandName, cgs.brandGuidelines.ToneOfVoice, cgs.brandGuidelines.TargetAudience)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := cgs.plannerAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var outline ContentOutline
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        if err := json.Unmarshal([]byte(response), &outline); err != nil {
            return nil, fmt.Errorf("failed to parse outline: %w", err)
        }
    }

    return &outline, nil
}

type ContentOutline struct {
    Introduction string              `json:"introduction"`
    Sections     []OutlineSection    `json:"sections"`
    Conclusion   string              `json:"conclusion"`
    CallToAction string              `json:"call_to_action"`
    Keywords     []string            `json:"keywords"`
}

type OutlineSection struct {
    Title      string   `json:"title"`
    Subsections []string `json:"subsections"`
    KeyPoints  []string `json:"key_points"`
    WordTarget int      `json:"word_target"`
}

func (cgs *ContentGenerationSystem) generateContent(ctx context.Context, request ContentRequest, outline *ContentOutline, research *TopicResearch) (string, error) {
    prompt := fmt.Sprintf(`Write high-quality content based on the following specifications:

Title: %s
Content Type: %s
Target Length: %d words
Tone: %s
Target Audience: %s

Content Outline:
%s

Research Information:
Key Points: %s
Statistics: %s

Brand Guidelines:
- Brand Name: %s
- Tone of Voice: %s
- Key Messages: %s
- Writing Style: %s

SEO Requirements:
- Primary Keywords: %s
- Target keyword density: %.1f%%

Please write engaging, well-structured content that:
1. Follows the provided outline
2. Incorporates research insights and statistics
3. Maintains consistent brand voice and tone
4. Includes relevant keywords naturally
5. Engages the target audience
6. Provides actionable value
7. Includes proper headings and structure
8. Ends with a compelling call-to-action

Write the complete content now:`,
        request.Title, request.ContentType, request.TargetLength, request.Tone, request.Audience,
        formatOutline(outline),
        strings.Join(research.KeyPoints, ", "), strings.Join(research.Statistics, ", "),
        cgs.brandGuidelines.BrandName, cgs.brandGuidelines.ToneOfVoice,
        strings.Join(cgs.brandGuidelines.KeyMessages, ", "), cgs.brandGuidelines.WritingStyle,
        strings.Join(request.Keywords, ", "), cgs.seoConfig.TargetKeywordDensity*100)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := cgs.creatorAgent.Run(ctx, state)
    if err != nil {
        return "", err
    }

    if len(result.Messages) == 0 {
        return "", fmt.Errorf("no content generated")
    }

    content := result.Messages[len(result.Messages)-1].TextContent()
    return content, nil
}

func (cgs *ContentGenerationSystem) optimizeForSEO(ctx context.Context, content string, keywords []string) (string, error) {
    prompt := fmt.Sprintf(`Optimize the following content for SEO while maintaining quality and readability:

Content:
%s

Target Keywords: %s

SEO Requirements:
- Keyword density: %.1f%%
- Meta description length: %d characters
- Title optimization
- Proper heading structure (H1, H2, H3)
- Internal linking opportunities
- Image alt text suggestions

Please provide:
1. SEO-optimized version of the content
2. Suggested meta description
3. SEO analysis and recommendations

Focus on natural keyword integration and user experience.`,
        content, strings.Join(keywords, ", "),
        cgs.seoConfig.TargetKeywordDensity*100, cgs.seoConfig.MetaDescriptionLength)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := cgs.seoAgent.Run(ctx, state)
    if err != nil {
        return content, err // Return original content if SEO optimization fails
    }

    if len(result.Messages) == 0 {
        return content, nil
    }

    response := result.Messages[len(result.Messages)-1].TextContent()
    
    // Extract optimized content (simplified - in production, use structured parsing)
    lines := strings.Split(response, "\n")
    var optimizedContent strings.Builder
    contentStarted := false
    
    for _, line := range lines {
        if strings.Contains(strings.ToLower(line), "optimized content") {
            contentStarted = true
            continue
        }
        if contentStarted && strings.Contains(strings.ToLower(line), "meta description") {
            break
        }
        if contentStarted {
            optimizedContent.WriteString(line + "\n")
        }
    }

    if optimizedContent.Len() > 0 {
        return strings.TrimSpace(optimizedContent.String()), nil
    }
    
    return content, nil
}

func (cgs *ContentGenerationSystem) checkQuality(ctx context.Context, content string) (*QualityAnalysis, error) {
    prompt := fmt.Sprintf(`Analyze the quality of this content and provide detailed feedback:

Content:
%s

Please evaluate:
1. Grammar and spelling accuracy
2. Clarity and readability
3. Engagement factor
4. Factual accuracy (based on general knowledge)
5. Structure and flow
6. Value to the reader

Provide a quality score (0.0-1.0) and specific recommendations for improvement.

Return analysis as JSON:
{
    "score": 0.0-1.0,
    "grammar_errors": 0,
    "spelling_errors": 0,
    "clarity_score": 0.0-1.0,
    "engagement_score": 0.0-1.0,
    "factual_accuracy": 0.0-1.0,
    "issues": ["issue1", "issue2"],
    "suggestions": ["suggestion1", "suggestion2"]
}`, content)

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := cgs.qualityAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var analysis QualityAnalysis
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        if err := json.Unmarshal([]byte(response), &analysis); err != nil {
            return nil, fmt.Errorf("failed to parse quality analysis: %w", err)
        }
    }

    return &analysis, nil
}

func (cgs *ContentGenerationSystem) validateBrand(ctx context.Context, content string) (*BrandAnalysis, error) {
    prompt := fmt.Sprintf(`Analyze this content for brand compliance:

Content:
%s

Brand Guidelines:
- Brand Name: %s
- Tone of Voice: %s
- Writing Style: %s
- Key Messages: %s
- Target Audience: %s
- Brand Values: %s

Evaluate:
1. Tone consistency with brand voice
2. Message alignment with brand values
3. Style compliance with guidelines
4. Overall brand voice match

Return analysis as JSON:
{
    "score": 0.0-1.0,
    "tone_consistency": 0.0-1.0,
    "message_alignment": 0.0-1.0,
    "style_compliance": 0.0-1.0,
    "voice_match": 0.0-1.0,
    "issues": ["issue1", "issue2"],
    "suggestions": ["suggestion1", "suggestion2"]
}`,
        content,
        cgs.brandGuidelines.BrandName, cgs.brandGuidelines.ToneOfVoice,
        cgs.brandGuidelines.WritingStyle, strings.Join(cgs.brandGuidelines.KeyMessages, ", "),
        cgs.brandGuidelines.TargetAudience, strings.Join(cgs.brandGuidelines.BrandValues, ", "))

    state := domain.NewState()
    state.AddMessage(domain.NewTextMessage(domain.RoleUser, prompt))

    result, err := cgs.brandAgent.Run(ctx, state)
    if err != nil {
        return nil, err
    }

    var analysis BrandAnalysis
    if len(result.Messages) > 0 {
        response := result.Messages[len(result.Messages)-1].TextContent()
        if err := json.Unmarshal([]byte(response), &analysis); err != nil {
            return nil, fmt.Errorf("failed to parse brand analysis: %w", err)
        }
    }

    return &analysis, nil
}

// Database operations
func (cgs *ContentGenerationSystem) storeContent(ctx context.Context, content *ContentPiece) error {
    query := `INSERT INTO content_pieces 
              (title, content, content_type, status, author_id, topic_id, keywords, tags,
               meta_description, seo_score, quality_score, brand_score, word_count, 
               reading_time, metadata)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
              RETURNING id, created_at`

    metadataJSON, _ := json.Marshal(content.Metadata)
    
    err := cgs.db.QueryRowContext(ctx, query,
        content.Title, content.Content, content.ContentType, content.Status,
        content.AuthorID, content.TopicID, pq.Array(content.Keywords), pq.Array(content.Tags),
        content.MetaDescription, content.SEOScore, content.QualityScore, content.BrandScore,
        content.WordCount, content.ReadingTime, metadataJSON,
    ).Scan(&content.ID, &content.CreatedAt)

    return err
}

// HTTP API handlers
func (cgs *ContentGenerationSystem) GenerateContentHandler(c *gin.Context) {
    var request ContentRequest
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
    defer cancel()

    content, err := cgs.GenerateContent(ctx, request)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"content": content})
}

func (cgs *ContentGenerationSystem) GetContentHandler(c *gin.Context) {
    contentID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
        return
    }

    var content ContentPiece
    query := `SELECT * FROM content_pieces WHERE id = $1`
    
    err = cgs.db.GetContext(c.Request.Context(), &content, query, contentID)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"content": content})
}

func (cgs *ContentGenerationSystem) PublishContentHandler(c *gin.Context) {
    contentID, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid content ID"})
        return
    }

    var request struct {
        Channel     string     `json:"channel" binding:"required"`
        ScheduleFor *time.Time `json:"schedule_for"`
    }

    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Get content
    var content ContentPiece
    query := `SELECT * FROM content_pieces WHERE id = $1`
    err = cgs.db.GetContext(c.Request.Context(), &content, query, contentID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Content not found"})
        return
    }

    // Get publishing channel
    channel, exists := cgs.publishingChannels[request.Channel]
    if !exists {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel"})
        return
    }

    // Publish or schedule
    var publishErr error
    if request.ScheduleFor != nil {
        publishErr = channel.Schedule(c.Request.Context(), &content, *request.ScheduleFor)
        content.ScheduledFor = request.ScheduleFor
    } else {
        publishErr = channel.Publish(c.Request.Context(), &content)
        now := time.Now()
        content.PublishedAt = &now
        content.Status = "published"
    }

    if publishErr != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": publishErr.Error()})
        return
    }

    // Update content status
    updateQuery := `UPDATE content_pieces SET status = $1, published_at = $2, scheduled_for = $3, updated_at = $4 WHERE id = $5`
    _, err = cgs.db.ExecContext(c.Request.Context(), updateQuery,
        content.Status, content.PublishedAt, content.ScheduledFor, time.Now(), contentID)
    if err != nil {
        log.Printf("Failed to update content status: %v", err)
    }

    cgs.metrics.ContentPublished.Inc()
    cgs.metrics.ChannelPerformance.WithLabelValues(request.Channel, "publish").Inc()

    c.JSON(http.StatusOK, gin.H{
        "message": "Content published successfully",
        "content": content,
}
}

// Publishing channel implementations
type BlogChannel struct{}

func (bc *BlogChannel) Name() string { return "blog" }

func (bc *BlogChannel) Publish(ctx context.Context, content *ContentPiece) error {
    // Implementation would integrate with blog platform (WordPress, etc.)
    log.Printf("Publishing to blog: %s", content.Title)
    return nil
}

func (bc *BlogChannel) Schedule(ctx context.Context, content *ContentPiece, publishTime time.Time) error {
    // Implementation would schedule publication
    log.Printf("Scheduling blog post: %s for %v", content.Title, publishTime)
    return nil
}

func (bc *BlogChannel) GetAnalytics(ctx context.Context, contentID int) (*ChannelAnalytics, error) {
    // Implementation would fetch analytics from blog platform
    return &ChannelAnalytics{
        ContentID:  contentID,
        Channel:    "blog",
        Views:      1500,
        Engagement: 120,
        UpdatedAt:  time.Now(),
    }, nil
}

type SocialMediaChannel struct{}

func (smc *SocialMediaChannel) Name() string { return "social" }

func (smc *SocialMediaChannel) Publish(ctx context.Context, content *ContentPiece) error {
    log.Printf("Publishing to social media: %s", content.Title)
    return nil
}

func (smc *SocialMediaChannel) Schedule(ctx context.Context, content *ContentPiece, publishTime time.Time) error {
    log.Printf("Scheduling social post: %s for %v", content.Title, publishTime)
    return nil
}

func (smc *SocialMediaChannel) GetAnalytics(ctx context.Context, contentID int) (*ChannelAnalytics, error) {
    return &ChannelAnalytics{
        ContentID:  contentID,
        Channel:    "social",
        Views:      5000,
        Engagement: 350,
        Shares:     45,
        UpdatedAt:  time.Now(),
    }, nil
}

type EmailChannel struct{}

func (ec *EmailChannel) Name() string { return "email" }

func (ec *EmailChannel) Publish(ctx context.Context, content *ContentPiece) error {
    log.Printf("Sending email: %s", content.Title)
    return nil
}

func (ec *EmailChannel) Schedule(ctx context.Context, content *ContentPiece, publishTime time.Time) error {
    log.Printf("Scheduling email: %s for %v", content.Title, publishTime)
    return nil
}

func (ec *EmailChannel) GetAnalytics(ctx context.Context, contentID int) (*ChannelAnalytics, error) {
    return &ChannelAnalytics{
        ContentID:  contentID,
        Channel:    "email",
        Views:      2500,
        Engagement: 180,
        CTR:        0.042,
        UpdatedAt:  time.Now(),
    }, nil
}

// Utility functions
func calculateReadingTime(wordCount int) int {
    // Average reading speed: 200 words per minute
    return (wordCount + 199) / 200 // Round up
}

func formatOutline(outline *ContentOutline) string {
    var formatted strings.Builder
    formatted.WriteString("Introduction: " + outline.Introduction + "\n\n")
    
    for _, section := range outline.Sections {
        formatted.WriteString("Section: " + section.Title + "\n")
        for _, subsection := range section.Subsections {
            formatted.WriteString("  - " + subsection + "\n")
        }
        formatted.WriteString("\n")
    }
    
    formatted.WriteString("Conclusion: " + outline.Conclusion + "\n")
    formatted.WriteString("Call to Action: " + outline.CallToAction + "\n")
    
    return formatted.String()
}

func initializeContentMetrics() *ContentMetrics {
    metrics := &ContentMetrics{
        ContentCreated: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "content_created_total",
            Help: "Total number of content pieces created",
        }),
        ContentPublished: prometheus.NewCounter(prometheus.CounterOpts{
            Name: "content_published_total",
            Help: "Total number of content pieces published",
        }),
        SEOScoreDistribution: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "content_seo_score",
            Help: "Distribution of SEO scores",
        }),
        QualityScoreDistribution: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "content_quality_score",
            Help: "Distribution of quality scores",
        }),
        GenerationTime: prometheus.NewHistogram(prometheus.HistogramOpts{
            Name: "content_generation_time_seconds",
            Help: "Time taken to generate content",
        }),
        ChannelPerformance: prometheus.NewCounterVec(
            prometheus.CounterOpts{
                Name: "content_channel_performance_total",
                Help: "Content performance by channel",
            },
            []string{"channel", "action"},
        ),
    }

    prometheus.MustRegister(
        metrics.ContentCreated,
        metrics.ContentPublished,
        metrics.SEOScoreDistribution,
        metrics.QualityScoreDistribution,
        metrics.GenerationTime,
        metrics.ChannelPerformance,
    )

    return metrics
}

func main() {
    config := &ContentConfig{
        DatabaseURL:       "postgres://user:pass@localhost/content_db?sslmode=disable",
        BrandName:        "TechCorp",
        Industry:         "Technology",
        TargetAudience:   "Business professionals and tech enthusiasts",
        DefaultTone:      "Professional, informative, and engaging",
        ContentGuidelines: "Focus on providing actionable insights and practical value",
        MaxContentLength: 3000,
        SEOEnabled:       true,
        QualityThreshold: 0.8,
    }

    system, err := NewContentGenerationSystem(config)
    if err != nil {
        log.Fatal("Failed to create content system:", err)
    }

    // Setup HTTP server
    r := gin.Default()

    // API routes
    api := r.Group("/api/v1")
    {
        api.POST("/content", system.GenerateContentHandler)
        api.GET("/content/:id", system.GetContentHandler)
        api.POST("/content/:id/publish", system.PublishContentHandler)
        api.GET("/metrics", gin.WrapH(promhttp.Handler()))
    }

    // Health check
    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "healthy",
            "timestamp": time.Now(),
}
}

    log.Println("Content Generation System starting on :8080")
    log.Fatal(r.Run(":8080"))
}
```

## Usage Examples

### Generate Blog Article

```bash
curl -X POST http://localhost:8080/api/v1/content \
  -H "Content-Type: application/json" \
  -d '{
    "title": "The Future of AI in Content Marketing",
    "topic": "AI content marketing trends",
    "content_type": "article",
    "keywords": ["AI content marketing", "automation", "personalization"],
    "tags": ["AI", "marketing", "technology"],
    "author_id": "author123",
    "target_length": 1500,
    "tone": "professional",
    "audience": "marketing professionals",
    "channel": "blog"
  }'
```

### Publish Content

```bash
curl -X POST http://localhost:8080/api/v1/content/123/publish \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "blog"
  }'
```

### Schedule Content

```bash
curl -X POST http://localhost:8080/api/v1/content/123/publish \
  -H "Content-Type: application/json" \
  -d '{
    "channel": "social",
    "schedule_for": "2024-02-01T10:00:00Z"
  }'
```

## Key Features Demonstrated

1. **Multi-Agent Content Workflow** - Research, planning, creation, optimization
2. **SEO Optimization** - Automated keyword integration and content optimization
3. **Quality Assurance** - Grammar, readability, and engagement analysis
4. **Brand Compliance** - Automatic brand voice and guideline validation
5. **Multi-Channel Publishing** - Blog, social media, email distribution
6. **Content Analytics** - Performance tracking across channels
7. **Editorial Calendar** - Content planning and scheduling
8. **Database Integration** - Persistent storage for content and analytics

## Production Considerations

- **Content Approval Workflow** - Human review before publication
- **A/B Testing** - Test different content variations
- **Content Templates** - Reusable templates for different content types
- **Plagiarism Detection** - Ensure content originality
- **Compliance Checks** - Legal and regulatory compliance
- **Performance Monitoring** - Content engagement and conversion tracking

This content generation system showcases how to build a comprehensive AI-powered content creation platform using Go-LLMs, demonstrating workflow orchestration, quality control, and multi-channel publishing capabilities.