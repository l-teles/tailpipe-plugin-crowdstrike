package common

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/turbot/tailpipe-plugin-sdk/types"
)

// loadAll runs the loader over a file and returns every emitted row.
func loadAll(t *testing.T, path string) []any {
	t.Helper()
	ch := make(chan *types.RowData)
	if err := NewJSONLinesLoader().Load(context.Background(), &types.DownloadedArtifactInfo{LocalName: path}, ch); err != nil {
		t.Fatalf("load: %v", err)
	}
	var out []any
	for r := range ch {
		out = append(out, r.Data)
	}
	return out
}

func writeFile(t *testing.T, name string, content []byte, gz bool) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), name)
	if gz {
		var buf bytes.Buffer
		w := gzip.NewWriter(&buf)
		if _, err := w.Write(content); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		content = buf.Bytes()
	}
	if err := os.WriteFile(p, content, 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestJSONLinesLoader_StreamsGzipAndPlainFiles(t *testing.T) {
	t.Parallel()

	// A >64 KiB line is the case the SDK's row loaders silently truncate at.
	big := `{"big":"` + strings.Repeat("x", 200_000) + `"}`
	content := []byte("{\"n\":\"1\"}\r\n\n   \n" + big + "\n{\"n\":\"3\"}") // no trailing newline

	for _, tc := range []struct {
		name string
		gz   bool
	}{{"sample.txt.gz", true}, {"sample.jsonl", false}} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			rows := loadAll(t, writeFile(t, tc.name, content, tc.gz))
			if len(rows) != 3 {
				t.Fatalf("got %d rows, want 3: %v", len(rows), rows)
			}
			if rows[0] != `{"n":"1"}` || rows[1] != big || rows[2] != `{"n":"3"}` {
				t.Errorf("unexpected rows: %.40q", rows)
			}
		})
	}
}

func TestStreamLines_ReportsOverlongLineAndContinues(t *testing.T) {
	t.Parallel()

	in := strings.NewReader("{\"a\":1}\n" + strings.Repeat("y", 100) + "\n{\"a\":3}\n")
	ch := make(chan *types.RowData, 10)
	streamLines(context.Background(), in, 50, ch)
	close(ch)

	var rows []any
	for r := range ch {
		rows = append(rows, r.Data)
	}
	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3", len(rows))
	}
	le, ok := rows[1].(LineError)
	if !ok || le.Line != 2 || !errors.Is(le.Err, errLineTooLong) {
		t.Errorf("row 2: got %#v, want LineError for line 2", rows[1])
	}
	if rows[2] != `{"a":3}` {
		t.Errorf("row 3: got %v", rows[2])
	}
}

func TestJSONLinesLoader_ReportsCorruptGzip(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, _ = w.Write([]byte(strings.Repeat("{\"a\":\"b\"}\n", 10_000)))
	_ = w.Close()
	truncated := buf.Bytes()[:buf.Len()/2]

	rows := loadAll(t, writeFile(t, "cut.gz", truncated, false))
	if len(rows) == 0 {
		t.Fatal("no rows")
	}
	if _, ok := rows[len(rows)-1].(LineError); !ok {
		t.Errorf("last row: got %T, want LineError for the read failure", rows[len(rows)-1])
	}
}

func TestJSONLinesMapper(t *testing.T) {
	t.Parallel()

	m := NewJSONLinesMapper("test_mapper", func(doc map[string]any) *string { return StringFromMap(doc, "k") })

	got, err := m.Map(context.Background(), `{"k":"v"}`)
	if err != nil || got == nil || *got != "v" {
		t.Errorf("valid line: got %v, %v", got, err)
	}

	secret := "hunter2-hostname"
	if _, err := m.Map(context.Background(), `{"k":"`+secret); err == nil {
		t.Error("malformed line: want error")
	} else if strings.Contains(err.Error(), secret) {
		t.Errorf("error leaks line contents: %v", err)
	}

	le := LineError{Line: 7, Err: errLineTooLong}
	if _, err := m.Map(context.Background(), le); !errors.As(err, &le) {
		t.Errorf("LineError: got %v", err)
	}
}

func TestReadLine_KeepsFinalLineFillingBuffer(t *testing.T) {
	t.Parallel()

	// 32 bytes with no trailing newline fills a 16-byte reader exactly twice.
	want := strings.Repeat("z", 32)
	br := bufio.NewReaderSize(strings.NewReader(want), 16)

	line, tooLong, err := readLine(br, 100)
	if err != nil || tooLong || line != want {
		t.Fatalf("got %q, %v, %v; want the whole final line", line, tooLong, err)
	}
	if _, _, err := readLine(br, 100); !errors.Is(err, io.EOF) {
		t.Errorf("second read: got %v, want io.EOF", err)
	}
}
