package docconv

import (
	"strings"
	"testing"
)

func TestCleanHTML(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		all         bool
		wantContain []string
		wantAbsent  []string
	}{
		{
			name:        "strips script tags",
			html:        `<html><body><p>visible</p><script>alert('xss')</script></body></html>`,
			all:         false,
			wantContain: []string{"visible"},
			wantAbsent:  []string{"alert", "xss", "script"},
		},
		{
			name:        "strips style tags",
			html:        `<html><body><p>text</p><style>.foo{color:red}</style></body></html>`,
			all:         false,
			wantContain: []string{"text"},
			wantAbsent:  []string{"color", "style"},
		},
		{
			name:        "strips unknown tags like fb:like",
			html:        `<html><body><p>content</p><fb:like href="x">junk</fb:like></body></html>`,
			all:         false,
			wantContain: []string{"content"},
			wantAbsent:  []string{"junk", "fb:like"},
		},
		{
			name:        "keeps accepted structural tags",
			html:        `<html><body><div><p>para</p></div></body></html>`,
			all:         false,
			wantContain: []string{"<div>", "<p>", "para", "</p>", "</div>"},
		},
		{
			name:        "all=true includes html element content",
			html:        `<html><head><title>T</title></head><body><p>B</p></body></html>`,
			all:         true,
			wantContain: []string{"T", "B"},
		},
		{
			name:        "all=false skips head content",
			html:        `<html><head><title>T</title></head><body><p>B</p></body></html>`,
			all:         false,
			wantContain: []string{"B"},
			wantAbsent:  []string{"T"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanHTML(strings.NewReader(tt.html), tt.all)
			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("cleanHTML() output missing %q\ngot: %s", want, got)
				}
			}
			for _, absent := range tt.wantAbsent {
				if strings.Contains(got, absent) {
					t.Errorf("cleanHTML() output should not contain %q\ngot: %s", absent, got)
				}
			}
		})
	}
}

func TestAcceptedHTMLTag(t *testing.T) {
	accepted := []string{"div", "p", "br", "span", "body", "h1", "h2", "table", "tr", "td"}
	for _, tag := range accepted {
		if !acceptedHTMLTag(tag) {
			t.Errorf("acceptedHTMLTag(%q) = false, want true", tag)
		}
	}

	rejected := []string{"script", "style", "fb:like", "unknown", "iframe", "object"}
	for _, tag := range rejected {
		if acceptedHTMLTag(tag) {
			t.Errorf("acceptedHTMLTag(%q) = true, want false", tag)
		}
	}
}
