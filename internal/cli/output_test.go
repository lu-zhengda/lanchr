package cli

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestFprintJSON(t *testing.T) {
	t.Run("simple struct", func(t *testing.T) {
		var buf bytes.Buffer
		input := map[string]string{"key": "value"}

		if err := fprintJSON(&buf, input); err != nil {
			t.Fatalf("fprintJSON() error = %v", err)
		}

		var got map[string]string
		if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
			t.Fatalf("failed to parse output: %v", err)
		}
		if got["key"] != "value" {
			t.Errorf("got key=%q, want %q", got["key"], "value")
		}
	})

	t.Run("indented output", func(t *testing.T) {
		var buf bytes.Buffer
		input := map[string]int{"a": 1}

		if err := fprintJSON(&buf, input); err != nil {
			t.Fatalf("fprintJSON() error = %v", err)
		}

		// Indented JSON should contain newlines and spaces.
		output := buf.String()
		if output == `{"a":1}`+"\n" {
			t.Error("expected indented JSON, got compact")
		}
	})

	t.Run("nil value", func(t *testing.T) {
		var buf bytes.Buffer
		if err := fprintJSON(&buf, nil); err != nil {
			t.Fatalf("fprintJSON() error = %v", err)
		}
		if got := buf.String(); got != "null\n" {
			t.Errorf("got %q, want %q", got, "null\n")
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		var buf bytes.Buffer
		if err := fprintJSON(&buf, []string{}); err != nil {
			t.Fatalf("fprintJSON() error = %v", err)
		}
		if got := buf.String(); got != "[]\n" {
			t.Errorf("got %q, want %q", got, "[]\n")
		}
	})
}
