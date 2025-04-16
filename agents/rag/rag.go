package rag

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/document"
	"github.com/bububa/atomic-agents/components/embedder"
	"github.com/bububa/atomic-agents/components/vectordb"
	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/instructor-go"
)

type Options struct {
	name              string
	enhanceQueryAgent agents.TypeableAgent[schema.String, schema.String]
	embedder          embedder.Embedder
	chunker           embedder.Chunker
	vectordb          vectordb.Engine
	contextGenerator  func(string, []vectordb.Record) string
	searchOptions     []vectordb.SearchOption
}

type RAG[O schema.Schema] struct {
	agent agents.TypeableAgent[schema.String, O]
	Options
}

var (
	_ agents.TypeableAgent[schema.String, schema.String]   = (*RAG[schema.String])(nil)
	_ agents.StreamableAgent[schema.String, schema.String] = (*RAG[schema.String])(nil)
	_ agents.AnonymousAgent                                = (*RAG[schema.String])(nil)
	_ agents.AnonymousStreamableAgent                      = (*RAG[schema.String])(nil)
)

type Option func(*Options)

func WithName(name string) Option {
	return func(r *Options) {
		r.name = name
	}
}

func WithChunker(chunker embedder.Chunker) Option {
	return func(r *Options) {
		r.chunker = chunker
	}
}

func WithEmbedder(e embedder.Embedder) Option {
	return func(r *Options) {
		r.embedder = e
	}
}

func WithVectorDB(v vectordb.Engine) Option {
	return func(r *Options) {
		r.vectordb = v
	}
}

func WithEhanceQueryAgent(v agents.TypeableAgent[schema.String, schema.String]) Option {
	return func(r *Options) {
		r.enhanceQueryAgent = v
	}
}

func WithContextGenerator(fn func(string, []vectordb.Record) string) Option {
	return func(r *Options) {
		r.contextGenerator = fn
	}
}

func WithSearchOptions(opts ...vectordb.SearchOption) Option {
	return func(r *Options) {
		r.searchOptions = opts
	}
}

func NewRAG[O schema.Schema](agent agents.TypeableAgent[schema.String, O], opts ...Option) *RAG[O] {
	ret := new(RAG[O])
	ret.agent = agent
	for _, opt := range opts {
		opt(&ret.Options)
	}
	if ret.contextGenerator == nil {
		ret.contextGenerator = defaultContextGenerator
	}
	return ret
}

func (r *RAG[O]) Name() string {
	return r.name
}

func (r *RAG[O]) SetSearchOptions(opts ...vectordb.SearchOption) {
	r.searchOptions = opts
}

func (r *RAG[O]) AddDocuments(ctx context.Context, collectionName string, docs ...document.Document) (*components.LLMUsage, error) {
	totalUsage := new(components.LLMUsage)
	for _, doc := range docs {
		var parts []string
		content := doc.String()
		if r.chunker != nil {
			parts = r.chunker.SplitText(content)
		} else {
			parts = []string{content}
		}
		usage := new(components.LLMUsage)
		embeddings, err := r.embedder.BatchEmbed(ctx, parts, usage)
		totalUsage.Merge(usage)
		if err != nil {
			return totalUsage, err
		}
		records := make([]vectordb.Record, 0, len(embeddings))
		for _, embedding := range embeddings {
			embedding.Meta = doc.Meta()
			records = append(records, vectordb.Record{Embedding: embedding})
		}
		if err := r.vectordb.Insert(ctx, collectionName, records...); err != nil {
			return totalUsage, err
		}
	}
	return totalUsage, nil
}

func (r *RAG[O]) Search(ctx context.Context, query string, opts ...vectordb.SearchOption) ([]vectordb.Record, *components.LLMUsage, error) {
	embedding := new(embedder.Embedding)
	usage := new(components.LLMUsage)
	if err := r.embedder.Embed(ctx, query, embedding, usage); err != nil {
		return nil, nil, err
	}
	records, err := r.vectordb.Search(ctx, embedding.Embedding, opts...)
	if err != nil {
		return nil, usage, err
	}
	return records, usage, nil
}

func (r *RAG[O]) Run(ctx context.Context, query *schema.String, output *O, llmResp *components.LLMResponse) error {
	enhancedQuery, err := r.generateEnhancedQuery(ctx, query, llmResp)
	if err != nil {
		return err
	}
	usage := new(components.LLMUsage)
	usage.Merge(llmResp.Usage)
	records, searchUsage, err := r.Search(ctx, enhancedQuery, r.searchOptions...)
	usage.Merge(searchUsage)
	if err != nil {
		return err
	}
	if len(records) == 0 {
		return fmt.Errorf("no relevant information to answer question: %s", query.String())
	}
	input := schema.NewString(r.contextGenerator(query.String(), records))
	err = r.agent.Run(ctx, input, output, llmResp)
	usage.Merge(llmResp.Usage)
	llmResp.Usage = usage
	return err
}

func (r *RAG[O]) RunAnonymous(ctx context.Context, query any, llmResp *components.LLMResponse) (any, error) {
	if query == nil {
		return nil, errors.New("nil input")
	}
	var input *schema.String
	switch t := query.(type) {
	case string:
		input = schema.NewString(t)
	case schema.String:
		input = &t
	case *string:
		input = schema.NewString(*t)
	case *schema.String:
		input = t
	default:
		return nil, errors.New("invalid input schema")
	}
	output := new(O)
	if err := r.Run(ctx, input, output, llmResp); err != nil {
		return nil, err
	}
	return output, nil
}

func (r *RAG[O]) Stream(ctx context.Context, query *schema.String) (<-chan instructor.StreamData, agents.MergeResponse, error) {
	streamAgent, ok := r.agent.(agents.StreamableAgent[schema.String, O])
	if !ok {
		return nil, nil, errors.New("RAG agent is not streamable")
	}
	llmResp := new(components.LLMResponse)
	enhancedQuery, err := r.generateEnhancedQuery(ctx, query, llmResp)
	if err != nil {
		return nil, nil, err
	}
	records, searchUsage, err := r.Search(ctx, enhancedQuery, r.searchOptions...)
	if llmResp.Usage == nil {
		llmResp.Usage = searchUsage
	} else {
		llmResp.Usage.Merge(searchUsage)
	}
	mergeResp := func(resp *components.LLMResponse) {
		*resp = *llmResp
	}
	if err != nil {
		return nil, mergeResp, err
	}
	if len(records) == 0 {
		return nil, mergeResp, fmt.Errorf("no relevant information to answer question: %s", query.String())
	}
	input := schema.NewString(r.contextGenerator(query.String(), records))
	ch, mergeFn, err := streamAgent.Stream(ctx, input)
	mergeResp = func(resp *components.LLMResponse) {
		if mergeFn != nil {
			mergeFn(resp)
		}
		if resp.Usage == nil {
			resp.Usage = llmResp.Usage
		} else {
			resp.Usage.Merge(llmResp.Usage)
		}
	}
	return ch, mergeResp, err
}

func (r *RAG[O]) StreamAnonymous(ctx context.Context, query any) (<-chan instructor.StreamData, agents.MergeResponse, error) {
	if query == nil {
		return nil, nil, errors.New("nil input")
	}
	var input *schema.String
	switch t := query.(type) {
	case string:
		input = schema.NewString(t)
	case schema.String:
		input = &t
	case *string:
		input = schema.NewString(*t)
	case *schema.String:
		input = t
	default:
		return nil, nil, errors.New("invalid input schema")
	}
	return r.Stream(ctx, input)
}

func (r *RAG[O]) generateEnhancedQuery(ctx context.Context, query *schema.String, llmResp *components.LLMResponse) (string, error) {
	if r.enhanceQueryAgent == nil {
		return query.String(), nil
	}
	var out schema.String
	if err := r.enhanceQueryAgent.Run(ctx, query, &out, llmResp); err != nil {
		return "", err
	}
	return out.String(), nil
}

func defaultContextGenerator(query string, records []vectordb.Record) string {
	// Build context from search results
	sb := new(strings.Builder)
	sb.WriteString("Based on the following information:\n\n")

	// Track total relevance score
	totalScore := 0.0
	for i, record := range records {
		fmt.Fprintf(sb, "%d. %s\n", i+1, record.Embedding.Object)
		if meta := record.Embedding.Meta; meta != nil {
			for k, v := range meta {
				fmt.Fprintf(sb, "  - %s: %s\n", k, v)
			}
		}
		fmt.Fprintf(sb, " - Score: %.3f\n", record.Score)
		totalScore += record.Score
	}

	fmt.Fprintf(sb, "\nPlease provide a comprehensive answer to this question: %s", query)
	return sb.String()
}
