package log

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"runtime"
	"sync"
	"time"
)

type groupOrAttrs struct {
	group string      // group name if non-empty
	attrs []slog.Attr // attrs if non-empty
}

type TextHandler struct {
	opts slog.HandlerOptions
	goas []groupOrAttrs
	mu   *sync.Mutex
	out  io.Writer
}

func (s *TextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= s.opts.Level.Level()
}

// append log bytes after Enable
func (s *TextHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := make([]byte, 0, 1024)
	// append level
	buf = s.appendAttr(buf, slog.Any(slog.LevelKey, r.Level), false)
	buf = append(buf, '\t')
	// append time
	if !r.Time.IsZero() {
		buf = s.appendAttr(buf, slog.Time(slog.TimeKey, r.Time), false)
		buf = append(buf, '\t')
	}
	// append msg
	buf = s.appendAttr(buf, slog.String(slog.MessageKey, r.Message), false)
	buf = append(buf, '\t')
	// append source
	if s.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		buf = s.appendAttr(buf, slog.String(slog.SourceKey, fmt.Sprintf("%s:%d", f.File, f.Line)), false)
	}
	if r.NumAttrs() > 0 {
		buf = append(buf, '\t')
	}
	// TODO: output the Attrs and groups from WithAttrs and WithGroup.
	r.Attrs(func(a slog.Attr) bool {
		buf = s.appendAttr(buf, a, true)
		return true
	})
	buf = append(buf, '\n')
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.out.Write(buf)
	return err
}

// return appended bytes with attrs
func (s *TextHandler) appendAttr(buf []byte, a slog.Attr, isKeyPrint bool) []byte {
	a.Value = a.Value.Resolve()
	// Ignore empty Attrs.
	if a.Equal(slog.Attr{}) {
		return buf
	}
	if isKeyPrint {
		buf = fmt.Appendf(buf, "%s: ", a.Key)
	}
	switch a.Value.Kind() {
	case slog.KindTime:
		buf = append(buf, a.Value.Time().Format(time.RFC3339)...)
	default:
		buf = append(buf, a.Value.String()...)
	}
	return buf
}

func (s *TextHandler) withGroupOrAttrs(goa groupOrAttrs) *TextHandler {
	h2 := *s
	h2.goas = make([]groupOrAttrs, len(s.goas)+1)
	copy(h2.goas, s.goas)
	h2.goas[len(h2.goas)-1] = goa
	return &h2
}

func (s *TextHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return s
	}
	return s.withGroupOrAttrs(groupOrAttrs{group: name})
}

func (s *TextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return s
	}
	return s.withGroupOrAttrs(groupOrAttrs{attrs: attrs})
}

func NewTextHandler(out io.Writer, opts *slog.HandlerOptions) *TextHandler {
	h := &TextHandler{
		mu:  &sync.Mutex{},
		out: out,
	}
	if opts != nil {
		h.opts = *opts
	}
	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}
	return h
}
