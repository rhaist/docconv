package docconv

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestNewLocalFile_fromReader(t *testing.T) {
	const content = "hello local file"
	lf, err := NewLocalFile(strings.NewReader(content))
	if err != nil {
		t.Fatalf("NewLocalFile() error = %v", err)
	}
	defer lf.Done()

	// Content must be readable from position 0.
	lf.Seek(0, io.SeekStart)
	got, err := io.ReadAll(lf)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}
	if string(got) != content {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestNewLocalFile_fromOsFile(t *testing.T) {
	// When given an *os.File, NewLocalFile should use it directly (no copy).
	f, err := os.CreateTemp(t.TempDir(), "docconv-test")
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("file content")
	f.Seek(0, io.SeekStart)

	lf, err := NewLocalFile(f)
	if err != nil {
		t.Fatalf("NewLocalFile() error = %v", err)
	}

	if lf.File != f {
		t.Error("expected LocalFile to wrap the original *os.File, got a different file")
	}
	// Done() should not remove the original file.
	name := lf.Name()
	lf.Done()
	if _, err := os.Stat(name); os.IsNotExist(err) {
		t.Error("Done() removed the original file; it should only remove temp files it created")
	}
}

func TestNewLocalFile_doneRemovesTempFile(t *testing.T) {
	lf, err := NewLocalFile(strings.NewReader("data"))
	if err != nil {
		t.Fatalf("NewLocalFile() error = %v", err)
	}

	name := lf.Name()
	if _, err := os.Stat(name); err != nil {
		t.Fatalf("temp file does not exist before Done(): %v", err)
	}

	lf.Done()

	if _, err := os.Stat(name); !os.IsNotExist(err) {
		t.Errorf("temp file still exists after Done()")
	}
}
