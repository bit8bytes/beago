package redact_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/bit8bytes/beago/redact"
)

func TestAnonymize(t *testing.T) {
	m := &redact.Mask{}
	m.Add("Anna Müller")
	m.Add("Berlin Grundschule")

	in := strings.NewReader("Student Anna Müller attends Berlin Grundschule.")
	var out bytes.Buffer
	if err := m.Anonymize().Handle(context.Background(), in, &out); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if strings.Contains(got, "Anna Müller") || strings.Contains(got, "Berlin Grundschule") {
		t.Errorf("real data still present: %s", got)
	}
	if !strings.Contains(got, "TOKEN_0") || !strings.Contains(got, "TOKEN_1") {
		t.Errorf("tokens missing: %s", got)
	}
}

func TestRehydrate(t *testing.T) {
	m := &redact.Mask{}
	m.Add("Anna Müller")
	m.Add("Berlin Grundschule")

	in := strings.NewReader("Student TOKEN_0 attends TOKEN_1.")
	var out bytes.Buffer
	if err := m.Rehydrate().Handle(context.Background(), in, &out); err != nil {
		t.Fatal(err)
	}

	got := out.String()
	if strings.Contains(got, "TOKEN_0") || strings.Contains(got, "TOKEN_1") {
		t.Errorf("tokens still present: %s", got)
	}
	if !strings.Contains(got, "Anna Müller") || !strings.Contains(got, "Berlin Grundschule") {
		t.Errorf("real data missing: %s", got)
	}
}

func TestRoundtrip(t *testing.T) {
	m := &redact.Mask{}
	m.Add("Anna Müller")

	original := "Anna Müller scored 95 points."

	var anon bytes.Buffer
	m.Anonymize().Handle(context.Background(), strings.NewReader(original), &anon)

	var restored bytes.Buffer
	m.Rehydrate().Handle(context.Background(), &anon, &restored)

	if restored.String() != original {
		t.Errorf("roundtrip failed: got %q, want %q", restored.String(), original)
	}
}
