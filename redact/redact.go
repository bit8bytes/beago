// Package redact provides pipe handlers for anonymizing and rehydrating personal data
// before and after LLM processing.
package redact

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/bit8bytes/beago/pipe"
)

// Mask holds a bidirectional mapping between real values and anonymous tokens.
type Mask struct {
	mu      sync.RWMutex
	forward []string // real, token, real, token, ...
	inverse []string // token, real, token, real, ...
	count   int
}

// Add registers a real value and auto-generates a token for it.
// An optional label is used as the token prefix (e.g. "PERSON" → "PERSON_0").
// If no label is provided, "TOKEN" is used.
func (m *Mask) Add(real string, label ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	prefix := "TOKEN"
	if len(label) > 0 && label[0] != "" {
		prefix = label[0]
	}
	token := fmt.Sprintf("%s_%d", prefix, m.count)
	m.count++
	m.forward = append(m.forward, real, token)
	m.inverse = append(m.inverse, token, real)
}

// Anonymize returns a Handler that replaces real values with tokens.
func (m *Mask) Anonymize() pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		m.mu.RLock()
		replacer := strings.NewReplacer(m.forward...)
		m.mu.RUnlock()
		_, err = io.WriteString(w, replacer.Replace(string(data)))
		return err
	})
}

// Rehydrate returns a Handler that replaces tokens back with real values.
func (m *Mask) Rehydrate() pipe.Handler {
	return pipe.HandlerFunc(func(ctx context.Context, r io.Reader, w io.Writer) error {
		data, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		m.mu.RLock()
		replacer := strings.NewReplacer(m.inverse...)
		m.mu.RUnlock()
		_, err = io.WriteString(w, replacer.Replace(string(data)))
		return err
	})
}
