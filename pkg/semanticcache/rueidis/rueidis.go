package rueidis

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/redis/rueidis"
	"github.com/redis/rueidis/om"
	"github.com/samber/lo"
)

type rueidisJSONOptions struct {
	dimension int
}

func WithDimension(dimension int) RueidisJSONCallOption {
	return func(o *rueidisJSONOptions) {
		o.dimension = dimension
	}
}

type RueidisJSONCallOption func(*rueidisJSONOptions)

func applyRueidisJSONCallOptions(defaultOpts *rueidisJSONOptions, opts []RueidisJSONCallOption) *rueidisJSONOptions {
	for _, o := range opts {
		o(defaultOpts)
	}

	return defaultOpts
}

type Cached[T any] struct {
	Key    string    `json:"key" redis:",key"` // the redis:",key" is required to indicate which field is the ULID key
	Ver    int64     `json:"ver" redis:",ver"` // the redis:",ver" is required to do optimistic locking to prevent lost update
	Vec    []float64 `json:"vec"`
	Object T         `json:"object"`
}

type Retrieved[T any] struct {
	Key    string  `json:"key"`
	Score  float64 `json:"score"`
	Object T       `json:"object"`
}

type SemanticCacheRueidisJSON[T any] struct {
	rueidis rueidis.Client
	repo    om.Repository[Cached[*T]]
	options *rueidisJSONOptions
}

func RueidisJSON[T any](name string, rueidis rueidis.Client, callOptions ...RueidisJSONCallOption) *SemanticCacheRueidisJSON[T] {
	var t Cached[*T]

	opts := applyRueidisJSONCallOptions(&rueidisJSONOptions{}, callOptions)

	return &SemanticCacheRueidisJSON[T]{
		rueidis: rueidis,
		repo:    om.NewJSONRepository(name, t, rueidis),
		options: opts,
	}
}

func (c *SemanticCacheRueidisJSON[T]) newCached(doc *T, vectors []float64) *Cached[*T] {
	entity := c.repo.NewEntity()

	return &Cached[*T]{
		Key:    entity.Key,
		Ver:    entity.Ver,
		Vec:    vectors,
		Object: doc,
	}
}

func (c *SemanticCacheRueidisJSON[T]) CacheVectors(ctx context.Context, doc *T, vectors []float64) (*Cached[*T], error) {
	cached := c.newCached(doc, vectors)

	err := c.repo.Save(ctx, cached)
	if err != nil {
		return nil, err
	}

	return cached, nil
}

func (c *SemanticCacheRueidisJSON[T]) createIndex(ctx context.Context, dimension int) error {
	dim := dimension
	if c.options.dimension > 0 {
		dim = c.options.dimension
	}

	err := c.repo.CreateIndex(ctx, func(schema om.FtCreateSchema) rueidis.Completed {
		return schema.
			FieldName("$.vec").As("vec").Vector("FLAT", 6, "TYPE", "FLOAT64", "DIM", strconv.FormatInt(int64(dim), 10), "DISTANCE_METRIC", "COSINE").
			Build()
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *SemanticCacheRueidisJSON[T]) ensureIndex(ctx context.Context, dimension int) error {
	cmd := c.rueidis.B().FtList().Build()

	indexes, err := c.rueidis.Do(ctx, cmd).AsStrSlice()
	if err != nil {
		return err
	}

	indexName := c.repo.IndexName()
	for _, idx := range indexes {
		if idx == indexName {
			return nil
		}
	}

	return c.createIndex(ctx, dimension)
}

func (c *SemanticCacheRueidisJSON[T]) RetrieveVectors(ctx context.Context, vectors []float64) ([]*Retrieved[*T], error) {
	err := c.ensureIndex(ctx, len(vectors))
	if err != nil {
		return nil, err
	}

	cmd := c.rueidis.B().
		FtSearch().Index(c.repo.IndexName()).Query("(*)=>[KNN 3 @vec $V]").
		Return("1").Identifier("$.object").
		Sortby("__vec_score").
		Params().Nargs(2).NameValue().NameValue("V", rueidis.VectorString64(vectors)).
		Dialect(2).
		Build()

	_, records, err := c.rueidis.Do(ctx, cmd).AsFtSearch()
	if err != nil {
		return nil, err
	}

	return lo.Map(records, func(record rueidis.FtSearchDoc, _ int) *Retrieved[*T] {
		var object T
		_ = json.Unmarshal([]byte(record.Doc["$.object"]), &object)

		return &Retrieved[*T]{Key: record.Key, Score: record.Score, Object: &object}
	}), nil
}
