package esdsl

import (
	"encoding/json"
	"fmt"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
)

type ESParser struct{}

func (esp *ESParser) Parse(q string) (bq query.Query, err error) {
	var raw map[string]any
	if err := json.Unmarshal([]byte(q), &raw); err != nil {
		return nil, fmt.Errorf("invalid query json: %w", err)
	}

	bq = extractQuery(raw, "")

	if bq == nil {
		return nil, fmt.Errorf("unsupported query format: %+v", raw)
	}

	return bq, nil
}

func extractQuery(q map[string]any, parentPath string) query.Query {
	// handle match query
	if match, ok := q["match"].(map[string]any); ok {
		return parseMatchWithParent(match, parentPath)
	}

	// handle match_phrase query
	if matchPhrase, ok := q["match_phrase"].(map[string]any); ok {
		return parseMatchPhraseWithParent(matchPhrase, parentPath)
	}

	// handle bool query
	if boolPart, ok := q["bool"].(map[string]any); ok {
		return parseBoolWithParent(boolPart, parentPath)
	}

	// handle nested query
	if nested, ok := q["nested"].(map[string]any); ok {
		path, _ := nested["path"].(string)
		inner, _ := nested["query"].(map[string]any)
		if inner == nil || path == "" {
			return nil
		}
		// prepend nested path as parent
		return extractQuery(inner, path)
	}

	return nil
}

func parseMatchWithParent(match map[string]any, parentPath string) query.Query {
	for field, v := range match {
		vmap, _ := v.(map[string]any)
		term, _ := vmap["query"].(string)
		boost, _ := vmap["boost"].(float64)

		mq := bleve.NewMatchQuery(term)

		// prepend parent path if exists
		fieldName := field
		if parentPath != "" {
			fieldName = parentPath + "." + field
		}
		mq.SetField(fieldName)
		if boost > 0 {
			mq.SetBoost(boost)
		}
		return mq
	}
	return nil
}

func parseMatchPhraseWithParent(match map[string]any, parentPath string) query.Query {
	for field, v := range match {
		vmap, _ := v.(map[string]any)
		term, _ := vmap["query"].(string)
		boost, _ := vmap["boost"].(float64)

		mq := bleve.NewMatchPhraseQuery(term)

		// prepend parent path if exists
		fieldName := field
		if parentPath != "" {
			fieldName = parentPath + "." + field
		}
		mq.SetField(fieldName)
		if boost > 0 {
			mq.SetBoost(boost)
		}
		return mq
	}
	return nil
}

func parseBoolWithParent(boolPart map[string]any, parentPath string) query.Query {
	bq := bleve.NewBooleanQuery()

	// must clauses
	if mustArr, ok := boolPart["must"].([]any); ok {
		for _, m := range mustArr {
			if mm, ok := m.(map[string]any); ok {
				if mq := extractQuery(mm, parentPath); mq != nil {
					bq.AddMust(mq)
				}
			}
		}
	}

	// should clauses
	if shouldArr, ok := boolPart["should"].([]any); ok {
		for _, s := range shouldArr {
			if sm, ok := s.(map[string]any); ok {
				if sq := extractQuery(sm, parentPath); sq != nil {
					bq.AddShould(sq)
				}
			}
		}
	}

	return bq
}
