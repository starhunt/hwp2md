package hwp5

import (
	"testing"
)

func TestParseFileHeader(t *testing.T) {
	// 유효한 HWP 5.0 파일 헤더 생성
	data := make([]byte, FileHeaderSize)

	// 시그니처: "HWP Document File"
	copy(data[0:32], []byte(Signature))

	// 버전: 5.0.3.0 (little-endian)
	data[32] = 0 // Revision
	data[33] = 3 // Build
	data[34] = 0 // Minor
	data[35] = 5 // Major

	// 플래그: 압축됨
	data[36] = 0x01
	data[37] = 0x00
	data[38] = 0x00
	data[39] = 0x00

	header, err := ParseFileHeader(data)
	if err != nil {
		t.Fatalf("ParseFileHeader failed: %v", err)
	}

	// 버전 확인
	if header.Version.Major != 5 {
		t.Errorf("Expected major version 5, got %d", header.Version.Major)
	}
	if header.Version.Minor != 0 {
		t.Errorf("Expected minor version 0, got %d", header.Version.Minor)
	}
	if header.Version.Build != 3 {
		t.Errorf("Expected build 3, got %d", header.Version.Build)
	}

	// 버전 문자열 확인
	if header.Version.String() != "5.0.3.0" {
		t.Errorf("Expected version string '5.0.3.0', got %s", header.Version.String())
	}

	// 압축 플래그 확인
	if !header.IsCompressed() {
		t.Error("Expected IsCompressed() to be true")
	}

	if header.IsEncrypted() {
		t.Error("Expected IsEncrypted() to be false")
	}
}

func TestParseFileHeader_InvalidSignature(t *testing.T) {
	data := make([]byte, FileHeaderSize)
	copy(data[0:32], []byte("Invalid Signature"))

	_, err := ParseFileHeader(data)
	if err == nil {
		t.Error("Expected error for invalid signature")
	}
}

func TestParseFileHeader_TooSmall(t *testing.T) {
	data := make([]byte, 100) // FileHeaderSize보다 작음

	_, err := ParseFileHeader(data)
	if err == nil {
		t.Error("Expected error for small header")
	}
}

func TestRecordHeader(t *testing.T) {
	// 레코드 헤더 테스트: TagID=0x42 (PARA_HEADER), Level=0, Size=22
	// 구조: [TagID:10비트][Level:10비트][Size:12비트]
	// 0x42 = 66, Level=0, Size=22
	// Binary: 00000000010001010000000001000010
	// Rearranged for little-endian uint32:
	// Size(12) | Level(10) | TagID(10)
	// 22 << 20 | 0 << 10 | 66 = 0x01600042

	headerVal := uint32(22<<20 | 0<<10 | 66)
	data := make([]byte, 4)
	data[0] = byte(headerVal)
	data[1] = byte(headerVal >> 8)
	data[2] = byte(headerVal >> 16)
	data[3] = byte(headerVal >> 24)

	header := ParseRecordHeader(data)

	if header.TagID() != 66 {
		t.Errorf("Expected TagID 66 (0x42), got %d", header.TagID())
	}

	if header.Level() != 0 {
		t.Errorf("Expected Level 0, got %d", header.Level())
	}

	if header.Size() != 22 {
		t.Errorf("Expected Size 22, got %d", header.Size())
	}
}

func TestRecordReader(t *testing.T) {
	// 간단한 레코드 데이터 생성
	// Record 1: TagID=0x10, Level=0, Size=4, Data=[0x01,0x00,0x00,0x00]
	data := make([]byte, 8)

	// 헤더: TagID=16, Level=0, Size=4
	headerVal := uint32(4<<20 | 0<<10 | 16)
	data[0] = byte(headerVal)
	data[1] = byte(headerVal >> 8)
	data[2] = byte(headerVal >> 16)
	data[3] = byte(headerVal >> 24)

	// 데이터: 4바이트
	data[4] = 0x01
	data[5] = 0x00
	data[6] = 0x00
	data[7] = 0x00

	reader := NewRecordReader(data)
	rec, err := reader.Read()
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if rec.TagID != 16 {
		t.Errorf("Expected TagID 16, got %d", rec.TagID)
	}

	if rec.Level != 0 {
		t.Errorf("Expected Level 0, got %d", rec.Level)
	}

	if rec.Size != 4 {
		t.Errorf("Expected Size 4, got %d", rec.Size)
	}

	if len(rec.Data) != 4 {
		t.Errorf("Expected Data length 4, got %d", len(rec.Data))
	}
}

func TestTextExtractor(t *testing.T) {
	te := NewTextExtractor()

	// UTF-16LE 인코딩된 "한글" 텍스트
	// '한' = U+D55C = 0xD55C (little-endian: 0x5C 0xD5)
	// '글' = U+AE00 = 0xAE00 (little-endian: 0x00 0xAE)
	data := []byte{0x5C, 0xD5, 0x00, 0xAE}

	text := te.ExtractText(data)
	if text != "한글" {
		t.Errorf("Expected '한글', got '%s'", text)
	}
}

func TestTextExtractor_WithTab(t *testing.T) {
	te := NewTextExtractor()

	// "A" + TAB + "B" in UTF-16LE
	// 'A' = 0x0041, TAB = 0x0009, 'B' = 0x0042
	data := []byte{
		0x41, 0x00, // A
		0x09, 0x00, // TAB
		0x42, 0x00, // B
	}

	text := te.ExtractText(data)
	if text != "A\tB" {
		t.Errorf("Expected 'A\\tB', got '%s'", text)
	}
}

func TestTextExtractor_WithLineBreak(t *testing.T) {
	te := NewTextExtractor()

	// "A" + LINE_BREAK + "B" in UTF-16LE
	// 'A' = 0x0041, LINE = 0x0000, 'B' = 0x0042
	data := []byte{
		0x41, 0x00, // A
		0x00, 0x00, // LINE (soft break)
		0x42, 0x00, // B
	}

	text := te.ExtractText(data)
	if text != "A\nB" {
		t.Errorf("Expected 'A\\nB', got '%s'", text)
	}
}

func TestDecodeUTF16LE(t *testing.T) {
	// "Test" in UTF-16LE
	data := []byte{0x54, 0x00, 0x65, 0x00, 0x73, 0x00, 0x74, 0x00}
	result := DecodeUTF16LE(data)
	if result != "Test" {
		t.Errorf("Expected 'Test', got '%s'", result)
	}
}

func TestDecodeUTF16LE_Korean(t *testing.T) {
	// "테스트" in UTF-16LE
	// '테' = U+D14C, '스' = U+C2A4, '트' = U+D2B8
	data := []byte{
		0x4C, 0xD1, // 테
		0xA4, 0xC2, // 스
		0xB8, 0xD2, // 트
	}
	result := DecodeUTF16LE(data)
	if result != "테스트" {
		t.Errorf("Expected '테스트', got '%s'", result)
	}
}

func TestTagName(t *testing.T) {
	tests := []struct {
		tagID    uint16
		expected string
	}{
		{TagDocumentProperties, "DOCUMENT_PROPERTIES"},
		{TagParaHeader, "PARA_HEADER"},
		{TagParaText, "PARA_TEXT"},
		{TagTable, "TABLE"},
		{0xFFFF, "UNKNOWN(0xFFFF)"},
	}

	for _, tt := range tests {
		result := TagName(tt.tagID)
		if result != tt.expected {
			t.Errorf("TagName(%d) = %s, expected %s", tt.tagID, result, tt.expected)
		}
	}
}

func TestDecompressStream(t *testing.T) {
	// zlib 압축된 "Hello" 데이터
	// 실제 zlib 압축 데이터 (deflate)
	compressed := []byte{
		0x78, 0x9C, // zlib header
		0xF3, 0x48, 0xCD, 0xC9, 0xC9, 0x07, 0x00, // compressed "Hello"
		0x05, 0x8C, 0x01, 0xF5, // checksum
	}

	decompressed, err := DecompressStream(compressed)
	if err != nil {
		t.Fatalf("DecompressStream failed: %v", err)
	}

	if string(decompressed) != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", string(decompressed))
	}
}

func TestCharShape_Attributes(t *testing.T) {
	cs := &CharShape{
		Attributes: 0x03, // Bold + Italic
		Height:     1000, // 10pt
	}

	if !cs.IsBold() {
		t.Error("Expected IsBold() to be true")
	}

	if !cs.IsItalic() {
		t.Error("Expected IsItalic() to be true")
	}

	fontSize := cs.GetFontSizePt()
	if fontSize != 10.0 {
		t.Errorf("Expected font size 10.0pt, got %f", fontSize)
	}
}
