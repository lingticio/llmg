package rueidis

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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
	name    string
	rueidis rueidis.Client
	repo    om.Repository[Cached[*T]]
	options *rueidisJSONOptions
}

func RueidisJSON[T any](name string, rueidis rueidis.Client, callOptions ...RueidisJSONCallOption) *SemanticCacheRueidisJSON[T] {
	var t Cached[*T]

	opts := applyRueidisJSONCallOptions(&rueidisJSONOptions{}, callOptions)

	return &SemanticCacheRueidisJSON[T]{
		name:    name,
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

func (c *SemanticCacheRueidisJSON[T]) CacheVectors(ctx context.Context, doc *T, vectors []float64, seconds time.Duration) (*Cached[*T], error) {
	cached := c.newCached(doc, vectors)

	err := c.repo.Save(ctx, cached)
	if err != nil {
		return nil, err
	}

	secondsNum := int64(seconds.Seconds())
	if seconds >= 1 {
		cmd := c.rueidis.B().Expire().Key(fmt.Sprintf("%s:%s", c.name, cached.Key)).Seconds(secondsNum).Build()

		err = c.rueidis.Do(ctx, cmd).Error()
		if err != nil {
			return nil, err
		}
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

func (c *SemanticCacheRueidisJSON[T]) RetrieveTop3ByVectors(ctx context.Context, vectors []float64) ([]*Retrieved[*T], error) {
	return c.RetrieveByVectors(ctx, vectors, 3)
}

func (c *SemanticCacheRueidisJSON[T]) RetrieveTop10ByVectors(ctx context.Context, vectors []float64) ([]*Retrieved[*T], error) {
	return c.RetrieveByVectors(ctx, vectors, 10)
}

func (c *SemanticCacheRueidisJSON[T]) RetrieveFirstByVectors(ctx context.Context, vectors []float64) (*Retrieved[*T], error) {
	retrieved, err := c.RetrieveByVectors(ctx, vectors, 1)
	if err != nil {
		return nil, err
	}
	if len(retrieved) == 0 {
		return nil, nil
	}

	return retrieved[0], nil
}

func (c *SemanticCacheRueidisJSON[T]) RetrieveByVectors(ctx context.Context, vectors []float64, first int) ([]*Retrieved[*T], error) {
	err := c.ensureIndex(ctx, len(vectors))
	if err != nil {
		return nil, err
	}

	cmd := c.rueidis.B().
		FtSearch().Index(c.repo.IndexName()).Query("(*)=>[KNN "+strconv.FormatInt(int64(first), 10)+" @vec $V]").
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
