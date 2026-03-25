package docconv

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/richardlehane/mscfb"
	"github.com/richardlehane/msoleps"
)

type docMetaResult struct {
	meta map[string]string
	err  error
}

type docBodyResult struct {
	body string
	err  error
}

// ConvertDoc converts an MS Word .doc to text.
func ConvertDoc(r io.Reader) (string, map[string]string, error) {
	f, err := NewLocalFile(r)
	if err != nil {
		return "", nil, fmt.Errorf("error creating local file: %v", err)
	}
	defer f.Done()

	// Meta data
	mc := make(chan docMetaResult, 1)
	go func() {
		meta := make(map[string]string)

		defer func() {
			if e := recover(); e != nil {
				mc <- docMetaResult{meta: meta, err: fmt.Errorf("panic reading doc metadata: %v", e)}
			}
		}()

		doc, err := mscfb.New(f)
		if err != nil {
			mc <- docMetaResult{meta: meta, err: fmt.Errorf("error reading doc metadata: %v", err)}
			return
		}

		props := msoleps.New()
		for entry, err := doc.Next(); err == nil; entry, err = doc.Next() {
			if msoleps.IsMSOLEPS(entry.Initial) {
				if err := props.Reset(doc); err != nil {
					mc <- docMetaResult{meta: meta, err: fmt.Errorf("error reading doc properties: %v", err)}
					return
				}

				for _, prop := range props.Property {
					meta[prop.Name] = prop.String()
				}
			}
		}

		const defaultTimeFormat = "2006-01-02 15:04:05.999999999 -0700 MST"

		// Convert parsed meta
		if tmp, ok := meta["LastSaveTime"]; ok {
			if t, err := time.Parse(defaultTimeFormat, tmp); err == nil {
				meta["ModifiedDate"] = fmt.Sprintf("%d", t.Unix())
			}
		}
		if tmp, ok := meta["CreateTime"]; ok {
			if t, err := time.Parse(defaultTimeFormat, tmp); err == nil {
				meta["CreatedDate"] = fmt.Sprintf("%d", t.Unix())
			}
		}

		mc <- docMetaResult{meta: meta}
	}()

	// Document body
	bc := make(chan docBodyResult, 1)
	go func() {
		var buf bytes.Buffer
		outputFile, err := os.CreateTemp("/tmp", "sajari-convert-")
		if err != nil {
			bc <- docBodyResult{err: fmt.Errorf("error creating temp file: %v", err)}
			return
		}
		defer os.Remove(outputFile.Name())
		defer outputFile.Close()

		if err = exec.Command("wvText", f.Name(), outputFile.Name()).Run(); err != nil {
			bc <- docBodyResult{err: fmt.Errorf("wvText error: %v", err)}
			return
		}

		if _, err = buf.ReadFrom(outputFile); err != nil {
			bc <- docBodyResult{err: fmt.Errorf("error reading wvText output: %v", err)}
			return
		}

		bc <- docBodyResult{body: buf.String()}
	}()

	br := <-bc
	mr := <-mc

	// If wvText failed or produced no output, fall back to DOCX parsing.
	// Some .doc files are actually DOCX-compatible (e.g. doc saved as docx).
	if br.err != nil || len(br.body) == 0 {
		f.Seek(0, 0)
		return ConvertDocx(f)
	}
	// Metadata errors are non-fatal: return body with whatever meta we have.
	_ = mr.err
	return br.body, mr.meta, nil
}
