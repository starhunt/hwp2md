package hwp5

import (
	"bytes"
	"compress/flate"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

// Record는 HWP 5.x 레코드 구조체
// 참조: HWP 5.0 명세서 2.3 레코드 구조
type Record struct {
	TagID uint16 // 레코드 종류 (10비트)
	Level uint16 // 논리적 계층 (10비트)
	Size  uint32 // 데이터 크기
	Data  []byte // 레코드 데이터
}

// RecordHeader는 4바이트 레코드 헤더
// 구조: [TagID:10비트][Level:10비트][Size:12비트]
type RecordHeader uint32

// ParseRecordHeader parses the 4-byte record header.
func ParseRecordHeader(data []byte) RecordHeader {
	return RecordHeader(binary.LittleEndian.Uint32(data))
}

// TagID returns the tag ID (10 bits).
func (h RecordHeader) TagID() uint16 {
	return uint16(h & 0x3FF)
}

// Level returns the nesting level (10 bits).
func (h RecordHeader) Level() uint16 {
	return uint16((h >> 10) & 0x3FF)
}

// Size returns the data size (12 bits).
// If size is 0xFFF (4095), the actual size follows in the next 4 bytes.
func (h RecordHeader) Size() uint16 {
	return uint16((h >> 20) & 0xFFF)
}

// RecordReader reads records from a stream.
type RecordReader struct {
	data   []byte
	offset int
}

// NewRecordReader creates a new record reader from raw stream data.
func NewRecordReader(data []byte) *RecordReader {
	return &RecordReader{
		data:   data,
		offset: 0,
	}
}

// Read reads the next record.
func (r *RecordReader) Read() (*Record, error) {
	if r.offset >= len(r.data) {
		return nil, io.EOF
	}

	// 최소 4바이트(헤더) 필요
	if r.offset+4 > len(r.data) {
		return nil, fmt.Errorf("incomplete record header at offset %d", r.offset)
	}

	header := ParseRecordHeader(r.data[r.offset : r.offset+4])
	r.offset += 4

	rec := &Record{
		TagID: header.TagID(),
		Level: header.Level(),
	}

	// 크기 결정
	size := uint32(header.Size())
	if size == 0xFFF {
		// 확장 크기: 다음 4바이트에서 실제 크기 읽기
		if r.offset+4 > len(r.data) {
			return nil, fmt.Errorf("incomplete extended size at offset %d", r.offset)
		}
		size = binary.LittleEndian.Uint32(r.data[r.offset : r.offset+4])
		r.offset += 4
	}
	rec.Size = size

	// 데이터 읽기
	if r.offset+int(size) > len(r.data) {
		return nil, fmt.Errorf("incomplete record data at offset %d: need %d bytes, have %d",
			r.offset, size, len(r.data)-r.offset)
	}
	rec.Data = r.data[r.offset : r.offset+int(size)]
	r.offset += int(size)

	return rec, nil
}

// ReadAll reads all records from the stream.
func (r *RecordReader) ReadAll() ([]*Record, error) {
	var records []*Record
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return records, err
		}
		records = append(records, rec)
	}
	return records, nil
}

// DecompressStream decompresses zlib/deflate-compressed stream data.
// HWP 5.x uses raw deflate (no zlib header) for stream compression.
func DecompressStream(data []byte) ([]byte, error) {
	// First try zlib (with header)
	if len(data) >= 2 {
		// Check for zlib header (0x78 0x9C, 0x78 0xDA, etc.)
		if data[0] == 0x78 {
			reader, err := zlib.NewReader(bytes.NewReader(data))
			if err == nil {
				defer reader.Close()
				decompressed, err := io.ReadAll(reader)
				if err == nil {
					return decompressed, nil
				}
			}
		}
	}

	// Try raw deflate (no header) - HWP 5.x uses this
	reader := flate.NewReader(bytes.NewReader(data))
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress (tried zlib and deflate): %w", err)
	}

	return decompressed, nil
}

// TagName returns the human-readable name for a tag ID.
func TagName(tagID uint16) string {
	names := map[uint16]string{
		TagDocumentProperties: "DOCUMENT_PROPERTIES",
		TagIDMappings:         "ID_MAPPINGS",
		TagBinData:            "BIN_DATA",
		TagFaceName:           "FACE_NAME",
		TagBorderFill:         "BORDER_FILL",
		TagCharShape:          "CHAR_SHAPE",
		TagTabDef:             "TAB_DEF",
		TagNumbering:          "NUMBERING",
		TagBullet:             "BULLET",
		TagParaShape:          "PARA_SHAPE",
		TagStyle:              "STYLE",
		TagDocData:            "DOC_DATA",
		TagParaHeader:         "PARA_HEADER",
		TagParaText:           "PARA_TEXT",
		TagParaCharShape:      "PARA_CHAR_SHAPE",
		TagParaLineSeg:        "PARA_LINE_SEG",
		TagParaRangeTag:       "PARA_RANGE_TAG",
		TagCtrlHeader:         "CTRL_HEADER",
		TagListHeader:         "LIST_HEADER",
		TagPageDef:            "PAGE_DEF",
		TagTable:              "TABLE",
		TagShapePicture:       "SHAPE_PICTURE",
		TagShapeComponent:     "SHAPE_COMPONENT",
	}

	if name, ok := names[tagID]; ok {
		return name
	}
	return fmt.Sprintf("UNKNOWN(0x%04X)", tagID)
}
