package parser

import (
	"bytes"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected Format
	}{
		{
			name:     "hwpx extension",
			path:     "document.hwpx",
			expected: FormatHWPX,
		},
		{
			name:     "HWPX uppercase",
			path:     "DOCUMENT.HWPX",
			expected: FormatHWPX,
		},
		{
			name:     "hwp extension",
			path:     "document.hwp",
			expected: FormatHWP,
		},
		{
			name:     "hwp5 extension",
			path:     "document.hwp5",
			expected: FormatHWP,
		},
		{
			name:     "unknown extension",
			path:     "document.docx",
			expected: FormatUnknown,
		},
		{
			name:     "no extension",
			path:     "document",
			expected: FormatUnknown,
		},
		{
			name:     "path with directory",
			path:     "/path/to/document.hwpx",
			expected: FormatHWPX,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := DetectFormat(tc.path)
			if got != tc.expected {
				t.Errorf("DetectFormat(%q) = %v, want %v", tc.path, got, tc.expected)
			}
		})
	}
}

func TestFormat_String(t *testing.T) {
	tests := []struct {
		format   Format
		expected string
	}{
		{FormatHWPX, "hwpx"},
		{FormatHWP, "hwp"},
		{FormatUnknown, "unknown"},
		{Format(999), "unknown"},
	}

	for _, tc := range tests {
		got := tc.format.String()
		if got != tc.expected {
			t.Errorf("Format(%d).String() = %q, want %q", int(tc.format), got, tc.expected)
		}
	}
}

func TestDetectFormatFromReader(t *testing.T) {
	// ZIP signature (PK\x03\x04)
	zipHeader := []byte{0x50, 0x4B, 0x03, 0x04, 0x00, 0x00, 0x00, 0x00}

	// HWP5 signature (HWP Document File)
	hwp5Header := append([]byte("HWP Document File"), make([]byte, 15)...)

	// Unknown format
	unknownHeader := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07}

	tests := []struct {
		name     string
		data     []byte
		expected Format
	}{
		{
			name:     "zip/hwpx format",
			data:     zipHeader,
			expected: FormatHWPX,
		},
		{
			name:     "hwp5 format",
			data:     hwp5Header,
			expected: FormatHWP,
		},
		{
			name:     "unknown format",
			data:     unknownHeader,
			expected: FormatUnknown,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reader := bytes.NewReader(tc.data)
			got, err := DetectFormatFromReader(reader)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.expected {
				t.Errorf("DetectFormatFromReader() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestDetectFormatFromReader_ShortData(t *testing.T) {
	// Data shorter than header size
	shortData := []byte{0x50, 0x4B}
	reader := bytes.NewReader(shortData)

	_, err := DetectFormatFromReader(reader)
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.ExtractImages != false {
		t.Error("expected ExtractImages to be false by default")
	}
	if opts.ImageDir != "" {
		t.Error("expected ImageDir to be empty by default")
	}
}
