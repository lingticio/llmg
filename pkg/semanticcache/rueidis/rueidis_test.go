package rueidis

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/lingticio/llmg/internal/datastore"
	"github.com/nekomeowww/xo"
	"github.com/samber/lo"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Example struct {
	Query      string `json:"query"`
	Completion string `json:"completion"`
}

func TestSemanticCache(t *testing.T) {
	r, err := datastore.NewRueidis()()
	require.NoError(t, err)
	require.NotNil(t, r)

	json := RueidisJSON[Example]("chat_semantic_cache", r, WithDimension(1536))

	config := openai.DefaultConfig(os.Getenv("OPENAI_API_KEY"))
	config.BaseURL = os.Getenv("OPENAI_API_BASEURL")

	openAI := openai.NewClientWithConfig(config)

	// Cache
	{
		embedding, err := openAI.CreateEmbeddings(context.Background(), openai.EmbeddingRequestStrings{
			Input:          []string{"When was ChatGPT released?", "Where is the headquarters of OpenAI?"},
			Model:          openai.AdaEmbeddingV2,
			EncodingFormat: openai.EmbeddingEncodingFormatFloat,
		})
		require.NoError(t, err)
		require.Len(t, embedding.Data, 2)

		ebd1 := lo.Map(embedding.Data[0].Embedding, func(item float32, _ int) float64 {
			return float64(item)
		})

		ebd2 := lo.Map(embedding.Data[1].Embedding, func(item float32, _ int) float64 {
			return float64(item)
		})

		_, err = json.CacheVectors(context.Background(), &Example{Query: "Birthday of ChatGPT"}, ebd1, time.Second*2)
		require.NoError(t, err)

		_, err = json.CacheVectors(context.Background(), &Example{Query: "Apple's headquarters"}, ebd2, time.Second*2)
		require.NoError(t, err)
	}

	// Retrieve
	{
		embeddingQuery, err := openAI.CreateEmbeddings(context.Background(), openai.EmbeddingRequestStrings{
			Input:          []string{"When is the birthday of ChatGPT?"},
			Model:          openai.AdaEmbeddingV2,
			EncodingFormat: openai.EmbeddingEncodingFormatFloat,
		})
		require.NoError(t, err)
		require.Len(t, embeddingQuery.Data, 1)

		ebdQuery := lo.Map(embeddingQuery.Data[0].Embedding, func(item float32, _ int) float64 {
			return float64(item)
		})

		retrieved, err := json.RetrieveTop10ByVectors(context.Background(), ebdQuery)
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		require.Len(t, retrieved, 2)
		require.NotZero(t, retrieved[0].Score)
		require.NotZero(t, retrieved[1].Score)

		assert.Equal(t, "Birthday of ChatGPT", retrieved[0].Object.Query)
		assert.Equal(t, "Apple's headquarters", retrieved[1].Object.Query)

		xo.PrintJSON(retrieved)
	}
}
