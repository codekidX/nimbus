package nimbus

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/codekidx/nimbus/dialect"
)

// D is document type for easier identification for nimbus Document
// which is used for providing corupus
type D map[string]any

// Engine is the test search engine
type Engine struct {
	idx      bleve.Index
	dsl      dialect.DSL
	d        dialect.Dialect
	idxField string
}

// Result is a minimal search result for assertions
type Result struct {
	ID     string
	Score  float64
	Fields map[string]any
}

// New creates a new in-memory Nimbus engine
func New(dsl dialect.DSL) (*Engine, error) {
	mapping := bleve.NewIndexMapping()
	idx, err := bleve.NewMemOnly(mapping)
	if err != nil {
		return nil, err
	}

	d := dialect.GetDialect(dsl)
	if d == nil {
		return nil, errors.New("no such dialect")
	}

	return &Engine{idx: idx, dsl: dsl, d: d, idxField: "id"}, nil
}

// WithCorpus indexes a slice of documents (maps or structs).
// Assumes each document has a unique "id" field.
func (e *Engine) WithCorpus(docs []any) (*Engine, error) {
	for _, doc := range docs {
		// Marshal/unmarshal to map for easy field access
		b, _ := json.Marshal(doc)
		var m map[string]any
		_ = json.Unmarshal(b, &m)

		id, ok := m[e.idxField].(string)
		if !ok {
			return nil, fmt.Errorf("document missing 'id' field")
		}
		if err := e.idx.Index(id, m); err != nil {
			return nil, err
		}
	}
	return e, nil
}

// WithDialect takes in the dialect of the input query language
// to evaluate
func (e *Engine) WithDialect(d dialect.DSL) *Engine {
	e.dsl = d
	return e
}

// WithIndexField takes the field name to index during corpus
// indexing stage
func (e *Engine) WithIndexField(f string) *Engine {
	e.idxField = f
	return e
}

// Query takes a small subset of ES DSL as JSON string.
// Supports:
//
//	{ "match": { "title": { "query": "sony", "boost": 2.0 } } }
//
//	{ "bool": {
//	    "must": [
//	      { "match": { "title": { "query": "sony" } } }
//	    ],
//	    "should": [
//	      { "match": { "description": { "query": "camera" } } }
//	    ]
//	  }}
func (e *Engine) Query(q string, size int) ([]Result, error) {
	bq, err := e.d.Parse(q)
	if err != nil {
		return nil, err
	}

	req := bleve.NewSearchRequestOptions(bq, size, 0, false)
	req.Fields = []string{"*"} // return all fields
	res, err := e.idx.Search(req)
	if err != nil {
		return nil, err
	}

	var out []Result
	for _, hit := range res.Hits {
		out = append(out, Result{
			ID:     hit.ID,
			Score:  hit.Score,
			Fields: hit.Fields,
		})
	}
	return out, nil
}
