# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`docconv` (module: `code.sajari.com/docconv/v2`) is a Go library that converts documents (PDF, DOC, DOCX, RTF, ODT, HTML, XML, PPTX, Apple Pages, images) to plain text. It wraps external CLI tools and provides a structured `Response{Body, Meta, MSecs, Error}`.

The `docd` subdirectory is an HTTP service (port 8888) built on top of the library.

## System Dependencies

Tests and some converters require system packages. On Debian/Ubuntu:
```bash
sudo apt-get install wv unrtf tidy poppler-utils
```
OCR support additionally requires `tesseract`.

## Build & Test

```bash
# Build
go build -v ./...

# Test (requires system deps above)
go test -v -race ./...

# Single package
go test -v -race ./docx_test/...

# With OCR support (requires gosseract/tesseract)
go build -tags ocr -v ./...
go test -tags ocr -v -race ./...

# Build the docd HTTP service
go build -v ./docd
```

## Architecture

**Dispatch pattern:** `Convert(r io.Reader, mimeType string, readability bool)` in `docconv.go` routes to format-specific converters based on MIME type. `ConvertPath` detects MIME via file extension.

**Format converters** (each returns `(string, map[string]string, error)`):
- Pure Go: `docx.go`, `odt.go`, `pages.go`, `pptx.go`, `xml.go`
- Wraps external CLI: `doc.go` (`wvText`), `pdf.go` (`pdftotext`/`pdfinfo`), `rtf.go` (`unrtf`), `html.go` (`tidy`)
- OCR variants: `pdf_ocr.go`, `image_ocr.go` — only compiled with `-tags ocr`
- `image.go`: base image handling without OCR

**Key utilities:**
- `local.go` — `LocalFile` ensures input is on disk (creates temp files for streams); converters that shell out need a real file path
- `limit.go` — wraps readers with 20MB cap
- `tidy.go` — wraps the `tidy` CLI for HTML/XML sanitization

**iWork/Pages format:** Uses a custom snappy decompressor (`snappy/`) and protobuf definitions (`iWork/`) to parse Apple's binary format.

**`docd` service:** HTTP handlers in `docd/convert.go` accept multipart form, path, or streaming input and return JSON. Routing via `gorilla/mux`.

**`client/`:** HTTP client package for talking to a remote `docd` instance.

## Adding a New Format

1. Create `<format>.go` with a `Convert<Format>(r io.Reader) (string, map[string]string, error)` function
2. Add a MIME type case in the `Convert` switch in `docconv.go`
3. Add the extension mapping in `MimeTypeByExtension`
4. Add test data in `<format>_test/testdata/` and a `_test.go` file
