package dialect

import (
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/codekidx/nimbus/dialect/esdsl"
)

type DSL uint8

const (
	ESDSL DSL = iota
)

type Dialect interface {
	Parse(q string) (query query.Query, err error)
}

func GetDialect(d DSL) Dialect {
	switch d {
	case ESDSL:
		return &esdsl.ESParser{}
	}

	return nil
}
