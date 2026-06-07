package model_test

import (
	"testing"

	"github.com/wzhejunqiu/data-nexus/internal/model"
)

func TestDefaultCSVFormat(t *testing.T) {
	def := model.DefaultCSVFormat()
	if def.Delimiter != "," || def.QuoteChar != `"` || !def.HasHeader {
		t.Fatalf("unexpected defaults: %+v", def)
	}
	if def.LineEnding != "crlf" || def.Encoding != "utf-8" {
		t.Fatalf("unexpected line ending or encoding: %+v", def)
	}
}

func TestCSVFormatOptionsNormalized(t *testing.T) {
	custom := model.CSVFormatOptions{
		Delimiter:        ";",
		QuoteChar:        "'",
		HasHeader:        false,
		NullValue:        "NULL",
		LineEnding:       "lf",
		Encoding:         "utf-8-bom",
		CommentChar:      "#",
		LazyQuotes:       true,
		TrimLeadingSpace: true,
	}
	n := custom.Normalized()
	if n.Delimiter != ";" || n.QuoteChar != "'" || n.HasHeader {
		t.Fatalf("unexpected normalized delimiter/quote/header: %+v", n)
	}
	if n.NullValue != "NULL" || n.CommentChar != "#" || !n.LazyQuotes || !n.TrimLeadingSpace {
		t.Fatalf("unexpected normalized optional fields: %+v", n)
	}
	if n.LineEnding != "lf" || n.Encoding != "utf-8-bom" {
		t.Fatalf("unexpected normalized line ending or encoding: %+v", n)
	}

	empty := model.CSVFormatOptions{}.Normalized()
	def := model.DefaultCSVFormat()
	if empty.Delimiter != def.Delimiter || empty.QuoteChar != def.QuoteChar {
		t.Fatalf("empty options should backfill defaults: %+v", empty)
	}
}
