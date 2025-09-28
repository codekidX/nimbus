// nimbus/example_test.go
package nimbus

import (
	"testing"

	"github.com/codekidx/nimbus/dialect"
)

func TestNimbusMatchBoost(t *testing.T) {
	engine, _ := New(dialect.ESDSL)
	_, err := engine.WithCorpus([]any{
		D{"id": "1", "title": "Sony camera"},
		D{"id": "2", "title": "Canon camera"},
	})
	if err != nil {
		t.Fatalf("failed to load corpus: %v", err)
	}

	// Boost Sony title matches higher
	query := `{
		"match": {
			"title": { "query": "sony", "boost": 2.0 }
		}
	}`

	results, err := engine.Query(query, 10)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	// Assert top hit
	if results[0].ID != "1" {
		t.Fatalf("expected doc 1 (Sony) to be top hit, got %s", results[0].ID)
	}
}

func TestNimbusMatchPhraseBoost(t *testing.T) {
	engine, _ := New(dialect.ESDSL)
	_, err := engine.WithCorpus([]any{
		D{"id": "1", "title": "Sony camera"},
		D{"id": "2", "title": "Canon camera"},
	})
	if err != nil {
		t.Fatalf("failed to load corpus: %v", err)
	}

	// Boost Sony title matches higher
	query := `{
		"match_phrase": {
			"title": { "query": "canon camera", "boost": 2.0 }
		}
	}`

	results, err := engine.Query(query, 10)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result")
	}

	// Assert top hit
	if results[0].ID != "2" {
		t.Fatalf("expected doc 2 (Canon) to be top hit, got %s", results[0].ID)
	}
}

func TestNimbusBoolQuery(t *testing.T) {
	engine, _ := New(dialect.ESDSL)
	_, _ = engine.WithCorpus([]any{
		D{"id": "1", "title": "Sony camera", "description": "Best camera for travel"},
		D{"id": "2", "title": "Canon DSLR", "description": "Professional photography"},
		D{"id": "3", "title": "Nikon lens", "description": "Great for wildlife"},
	})

	query := `{
		"bool": {
			"must": [
				{ "match": { "title": { "query": "sony" } } }
			],
			"should": [
				{ "match": { "description": { "query": "travel" } } }
			]
		}
	}`

	results, err := engine.Query(query, 10)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if len(results) == 0 || results[0].ID != "1" {
		t.Fatalf("expected doc 1 (Sony camera) to rank highest, got %+v", results)
	}
}

func TestNimbusNestedQuery(t *testing.T) {
	engine, _ := New(dialect.ESDSL)
	_, _ = engine.WithCorpus([]any{
		D{
			"id": "1",
			"user": D{
				"name": "Alice",
				"address": D{
					"city": "Paris",
				},
			},
		},
		D{
			"id": "2",
			"user": D{
				"name": "Bob",
				"address": D{
					"city": "London",
				},
			},
		},
	})

	query := `{
		"nested": {
			"path": "user.address",
			"query": {
				"match": { "city": { "query": "Paris" } }
			}
		}
	}`

	results, _ := engine.Query(query, 10)
	if len(results) != 1 || results[0].ID != "1" {
		t.Fatalf("expected doc 1 (Alice, Paris), got %+v", results)
	}
}
