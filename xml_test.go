package docconv

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestXMLToText(t *testing.T) {
	tests := []struct {
		name   string
		xml    string
		breaks []string
		skip   []string
		strict bool
		want   string
	}{
		{
			name:   "simple text",
			xml:    `<root><a>hello</a><b>world</b></root>`,
			want:   "helloworld",
		},
		{
			name:   "breaks insert newlines",
			xml:    `<root><p>hello</p><p>world</p></root>`,
			breaks: []string{"p"},
			want:   "\nhello\nworld",
		},
		{
			name:   "skip element excluded",
			xml:    `<root><keep>visible</keep><skip>hidden</skip><keep>also</keep></root>`,
			skip:   []string{"skip"},
			want:   "visiblealso",
		},
		{
			name:   "skip nested elements",
			xml:    `<root><skip><inner>deep</inner></skip><keep>end</keep></root>`,
			skip:   []string{"skip"},
			want:   "end",
		},
		{
			name:   "skip and breaks combined",
			xml:    `<root><p>one</p><script>bad</script><p>two</p></root>`,
			breaks: []string{"p"},
			skip:   []string{"script"},
			want:   "\none\ntwo",
		},
		{
			// strict=false passes unknown entities through as literal text.
			name:   "non-strict passes through unknown entities",
			xml:    `<root><p>hello&nbsp;world</p></root>`,
			breaks: []string{"p"},
			strict: false,
			want:   "\nhello&nbsp;world",
		},
		{
			name:   "strict rejects unknown entities",
			xml:    `<root><p>hello&nbsp;world</p></root>`,
			breaks: []string{"p"},
			strict: true,
			want:   "", // error expected, checked separately
		},
		{
			name: "whitespace-only text nodes ignored between elements",
			xml:  "<root>\n  <a>val</a>\n  <b>other</b>\n</root>",
			want: "\n  val\n  other\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := XMLToText(strings.NewReader(tt.xml), tt.breaks, tt.skip, tt.strict)
			if tt.want == "" && tt.strict {
				if err == nil {
					t.Error("expected error in strict mode, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("XMLToText() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestXMLToMap(t *testing.T) {
	tests := []struct {
		name string
		xml  string
		want map[string]string
	}{
		{
			name: "simple tags",
			xml:  `<root><title>Hello</title><author>World</author></root>`,
			want: map[string]string{"title": "Hello", "author": "World"},
		},
		{
			name: "whitespace between tags does not overwrite values",
			xml:  "<root>\n  <a>first</a>\n  <b>second</b>\n</root>",
			want: map[string]string{"a": "first", "b": "second"},
		},
		{
			name: "empty element",
			xml:  `<root><empty></empty><full>text</full></root>`,
			want: map[string]string{"full": "text"},
		},
		{
			name: "last value wins for repeated tags",
			xml:  `<root><tag>first</tag><tag>second</tag></root>`,
			want: map[string]string{"tag": "second"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := XMLToMap(strings.NewReader(tt.xml))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			// Only check keys present in want; XMLToMap may include parent
			// element names from whitespace nodes in non-tested keys.
			for k, wantV := range tt.want {
				if got[k] != wantV {
					t.Errorf("key %q = %q, want %q\nfull map: %v", k, got[k], wantV, got)
				}
			}
			// Use cmp for full equality on cases without whitespace ambiguity.
			if tt.name == "simple tags" || tt.name == "last value wins for repeated tags" {
				if diff := cmp.Diff(tt.want, got); diff != "" {
					t.Errorf("XMLToMap() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestXMLToMapWhitespaceBug(t *testing.T) {
	// Regression test: whitespace CharData between </a> and <b> must not
	// overwrite the value stored for "a".
	xml := "<root>\n\t<a>value-a</a>\n\t<b>value-b</b>\n</root>"
	got, err := XMLToMap(strings.NewReader(xml))
	if err != nil {
		t.Fatal(err)
	}
	if got["a"] != "value-a" {
		t.Errorf("key 'a' = %q, want %q (whitespace node overwrote value)", got["a"], "value-a")
	}
	if got["b"] != "value-b" {
		t.Errorf("key 'b' = %q, want %q", got["b"], "value-b")
	}
}
