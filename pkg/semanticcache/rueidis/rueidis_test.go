package rueidis

import (
	"context"
	"testing"

	"github.com/lingticio/llmg/internal/datastore"
	"github.com/nekomeowww/xo"
	"github.com/stretchr/testify/require"
)

type Example struct {
	Title string `json:"title"` // both NewHashRepository and NewJSONRepository use json tag as field name
	Name  string `json:"name"`
}

func TestSemanticCache(t *testing.T) {
	r, err := datastore.NewRueidis()()
	require.NoError(t, err)
	require.NotNil(t, r)

	json := RueidisJSON[Example]("my_example3", r)

	_, err = json.CacheVectors(context.Background(), &Example{Title: "my_title", Name: "my_name"}, []float64{1, 1, 1, 1, 1, 1, 1, 1})
	require.NoError(t, err)

	records, err := json.RetrieveVectors(context.Background(), []float64{1, 1, 1, 1, 1, 1, 1, 1})
	require.NoError(t, err)
	require.NotNil(t, records)

	xo.PrintJSON(records)
}
