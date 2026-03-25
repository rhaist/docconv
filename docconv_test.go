package docconv

import (
	"os/exec"
	"strings"
	"testing"
)

func TestConvertTrimsSpace(t *testing.T) {
	resp, err := Convert(
		strings.NewReader(" \n\n\nthe \n file\n\n"),
		"text/plain",
		false,
	)
	if err != nil {
		t.Fatalf("got error = %v, want nil", err)
	}
	if want := "the \n file"; resp.Body != want {
		t.Errorf("body = %v, want %v", resp.Body, want)
	}
}

func TestConvertUnknownMimeType(t *testing.T) {
	// Unknown MIME types should return an empty body with no error.
	resp, err := Convert(strings.NewReader("data"), "application/octet-stream", false)
	if err != nil {
		t.Fatalf("got error = %v, want nil", err)
	}
	if resp.Body != "" {
		t.Errorf("body = %q, want empty", resp.Body)
	}
}

func TestMimeTypeByExtension(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"doc.doc", "application/msword"},
		{"doc.docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"doc.pdf", "application/pdf"},
		{"doc.rtf", "application/rtf"},
		{"doc.html", "text/html"},
		{"doc.htm", "text/html"},
		{"doc.xml", "text/xml"},
		{"doc.txt", "text/plain"},
		{"doc.pptx", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
		{"doc.odt", "application/vnd.oasis.opendocument.text"},
		{"doc.pages", "application/vnd.apple.pages"},
		{"doc.png", "image/png"},
		{"doc.jpg", "image/jpeg"},
		{"doc.jpeg", "image/jpeg"},
		{"doc.tiff", "image/tiff"},
		{"DOC.PDF", "application/pdf"}, // case-insensitive
		{"noextension", "application/octet-stream"},
		{"unknown.xyz", "application/octet-stream"},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := MimeTypeByExtension(tt.filename)
			if got != tt.want {
				t.Errorf("MimeTypeByExtension(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestConvertXMLViaConvert(t *testing.T) {
	if _, err2 := exec.LookPath("tidy"); err2 != nil {
		t.Skip("tidy not installed")
	}

	resp, err := Convert(
		strings.NewReader(`<?xml version="1.0"?><root><item>hello</item></root>`),
		"text/xml",
		false,
	)
	if err != nil {
		t.Fatalf("Convert(text/xml) error = %v", err)
	}
	if !strings.Contains(resp.Body, "hello") {
		t.Errorf("body = %q, want it to contain %q", resp.Body, "hello")
	}
}
