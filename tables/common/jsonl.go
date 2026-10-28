package common

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/turbot/tailpipe-plugin-sdk/artifact_loader"
	"github.com/turbot/tailpipe-plugin-sdk/mappers"
	"github.com/turbot/tailpipe-plugin-sdk/types"
)

// MaxLineBytes caps a single FDR record. Some events (e.g. PeFileWritten)
// exceed 1 MiB, far beyond bufio.Scanner's 64 KiB default.
const MaxLineBytes = 16 << 20

const JSONLinesLoaderIdentifier = "crowdstrike_jsonl_loader"

// JSONLinesLoader streams one row per line from a plain or gzipped JSON-lines
// artifact. Unlike the SDK's row loaders it doesn't stop at lines over 64 KiB,
// and read failures are sent as LineError rows so they count as row errors.
type JSONLinesLoader struct{}

func NewJSONLinesLoader() artifact_loader.Loader { return &JSONLinesLoader{} }

func (JSONLinesLoader) Identifier() string { return JSONLinesLoaderIdentifier }

func (JSONLinesLoader) Load(ctx context.Context, info *types.DownloadedArtifactInfo, dataChan chan *types.RowData) error {
	f, err := os.Open(info.LocalName)
	if err != nil {
		return fmt.Errorf("error opening %s: %w", info.LocalName, err)
	}
	var r io.Reader = f
	var gz *gzip.Reader
	if filepath.Ext(info.LocalName) == ".gz" {
		if gz, err = gzip.NewReader(f); err != nil {
			_ = f.Close()
			return fmt.Errorf("error creating gzip reader for %s: %w", info.LocalName, err)
		}
		r = gz
	}

	go func() {
		defer func() {
			if gz != nil {
				_ = gz.Close()
			}
			_ = f.Close() // read-only; a close error can't lose data
			close(dataChan)
		}()
		streamLines(ctx, r, MaxLineBytes, dataChan)
	}()
	return nil
}

// LineError is emitted in place of a line that couldn't be read; the mapper
// returns it so the SDK records a row error instead of dropping it silently.
type LineError struct {
	Line int
	Err  error
}

func (e LineError) Error() string { return fmt.Sprintf("line %d: %v", e.Line, e.Err) }

var errLineTooLong = fmt.Errorf("line exceeds %d bytes", MaxLineBytes)

func streamLines(ctx context.Context, r io.Reader, maxLine int, dataChan chan<- *types.RowData) {
	br := bufio.NewReaderSize(r, 1<<20)
	send := func(data any) bool {
		select {
		case dataChan <- &types.RowData{Data: data}:
			return true
		case <-ctx.Done():
			return false
		}
	}

	for lineNo := 1; ; lineNo++ {
		line, tooLong, err := readLine(br, maxLine)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				send(LineError{Line: lineNo, Err: err})
			}
			return
		}
		switch {
		case tooLong:
			if !send(LineError{Line: lineNo, Err: errLineTooLong}) {
				return
			}
		case strings.TrimSpace(line) != "":
			if !send(line) {
				return
			}
		}
	}
}

// readLine returns the next line without its terminator. Lines longer than
// maxLine are consumed and reported via tooLong so reading can continue.
func readLine(br *bufio.Reader, maxLine int) (line string, tooLong bool, err error) {
	var buf []byte
	for {
		frag, isPrefix, err := br.ReadLine()
		if err != nil {
			// A final unterminated line that exactly fills the buffer only
			// sees io.EOF after its last fragment; return it before the EOF.
			if errors.Is(err, io.EOF) && (len(buf) > 0 || tooLong) {
				return string(buf), tooLong, nil
			}
			return "", tooLong, err
		}
		if !tooLong {
			if len(buf)+len(frag) > maxLine {
				tooLong, buf = true, nil
			} else {
				buf = append(buf, frag...)
			}
		}
		if !isPrefix {
			return string(buf), tooLong, nil
		}
	}
}

// JSONLinesMapper decodes one JSON line into a table row via build.
type JSONLinesMapper[T any] struct {
	id    string
	build func(map[string]any) T
}

func NewJSONLinesMapper[T any](id string, build func(map[string]any) T) *JSONLinesMapper[T] {
	return &JSONLinesMapper[T]{id: id, build: build}
}

func (m *JSONLinesMapper[T]) Identifier() string { return m.id }

func (m *JSONLinesMapper[T]) Map(_ context.Context, a any, _ ...mappers.MapOption[T]) (T, error) {
	var zero T
	var data []byte
	switch v := a.(type) {
	case LineError:
		return zero, v
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return zero, fmt.Errorf("%s: expected string, got %T", m.id, a)
	}

	// Error messages deliberately omit line contents (FDR records carry PII).
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		return zero, fmt.Errorf("%s: invalid JSON: %w", m.id, err)
	}
	return m.build(doc), nil
}
