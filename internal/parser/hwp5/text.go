package hwp5

import (
	"encoding/binary"
	"strings"
	"unicode/utf16"
)

// TextExtractor extracts text from ParaText records.
type TextExtractor struct {
	controlChars map[uint16]bool
}

// NewTextExtractor creates a new text extractor.
func NewTextExtractor() *TextExtractor {
	return &TextExtractor{
		controlChars: map[uint16]bool{
			CharLine:           true,
			CharPara:           true,
			CharTab:            true,
			CharDrawingObj:     true,
			CharInlineStart:    true,
			CharFieldStart:     true,
			CharFieldEnd:       true,
			CharBookmark:       true,
			CharTitleMark:      true,
			CharHyphen:         true,
			CharNBSP:           true,
			CharFixedWidthNBSP: true,
			CharExtChar:        true,
		},
	}
}

// ExtractText extracts text from ParaText record data.
// ParaText는 UTF-16LE로 인코딩된 텍스트와 컨트롤 문자를 포함.
func (te *TextExtractor) ExtractText(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	var sb strings.Builder
	i := 0

	for i+1 < len(data) {
		char := binary.LittleEndian.Uint16(data[i : i+2])
		i += 2

		switch char {
		case CharLine:
			// 줄 나눔 - soft break
			sb.WriteString("\n")
		case CharPara:
			// 문단 나눔 - 이미 문단 단위로 처리하므로 무시
		case CharTab:
			sb.WriteString("\t")
		case CharDrawingObj:
			// 그리기 개체/표 - 14바이트 추가 정보 건너뛰기 (총 16바이트 - 이미 읽은 2바이트)
			if i+14 <= len(data) {
				i += 14
			}
		case CharInlineStart:
			// 인라인 개체 시작 - 14바이트 추가 정보 건너뛰기
			if i+14 <= len(data) {
				i += 14
			}
		case CharFieldStart:
			// 필드 시작 - 14바이트 추가 정보 건너뛰기
			if i+14 <= len(data) {
				i += 14
			}
		case CharFieldEnd:
			// 필드 끝
		case CharBookmark:
			// 책갈피 - 14바이트 추가 정보 건너뛰기
			if i+14 <= len(data) {
				i += 14
			}
		case CharTitleMark:
			// 제목 표시 - 14바이트 추가 정보 건너뛰기
			if i+14 <= len(data) {
				i += 14
			}
		case CharHyphen:
			sb.WriteString("-")
		case CharNBSP:
			sb.WriteString(" ")
		case CharFixedWidthNBSP:
			sb.WriteString(" ")
		case CharExtChar:
			// 확장 문자 - 14바이트 추가 정보 건너뛰기
			if i+14 <= len(data) {
				i += 14
			}
		default:
			// 일반 문자 (UTF-16LE)
			// 0x0001 ~ 0x001F 범위의 제어 문자는 무시 (HWP 특수 컨트롤 코드)
			// hwp.js 기준: Extended(1-3, 11-18, 21-23), Inline(4-9, 19-20)은 16바이트
			// Char(0, 10, 13)은 2바이트
			if char >= 0x0001 && char <= 0x001F {
				// Extended 타입 (1-3, 11-18, 21-23): 14바이트 추가 데이터
				// Inline 타입 (4-9, 19-20): 14바이트 추가 데이터
				// Char 타입 (10=줄바꿈, 13=문단끝): 추가 데이터 없음
				isExtended := (char >= 1 && char <= 3) || (char >= 11 && char <= 18) || (char >= 21 && char <= 23)
				isInline := (char >= 4 && char <= 9) || (char >= 19 && char <= 20)

				if isExtended || isInline {
					// 14바이트 추가 정보 건너뛰기
					if i+14 <= len(data) {
						i += 14
					}
				}
				// Char 타입 (0, 10, 13)은 추가 데이터 없음
			} else if char >= 0x0020 {
				sb.WriteRune(rune(char))
			}
		}
	}

	return sb.String()
}

// DecodeUTF16LE decodes UTF-16LE bytes to string.
func DecodeUTF16LE(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	// Convert bytes to uint16 slice
	u16s := make([]uint16, len(data)/2)
	for i := 0; i < len(u16s); i++ {
		u16s[i] = binary.LittleEndian.Uint16(data[i*2:])
	}

	// Decode UTF-16 to runes
	runes := utf16.Decode(u16s)

	// Remove null terminators
	for len(runes) > 0 && runes[len(runes)-1] == 0 {
		runes = runes[:len(runes)-1]
	}

	return string(runes)
}

// ControlInfo represents information about an inline control character.
type ControlInfo struct {
	Type   uint16 // 컨트롤 타입
	ID     uint32 // 컨트롤 ID
	Offset int    // 텍스트 내 위치
}

// ExtractTextWithControls extracts text and control information.
func (te *TextExtractor) ExtractTextWithControls(data []byte) (string, []ControlInfo) {
	if len(data) < 2 {
		return "", nil
	}

	var sb strings.Builder
	var controls []ControlInfo
	i := 0
	textOffset := 0

	for i+1 < len(data) {
		char := binary.LittleEndian.Uint16(data[i : i+2])
		i += 2

		switch char {
		case CharLine:
			sb.WriteString("\n")
			textOffset++
		case CharPara:
			// 무시
		case CharTab:
			sb.WriteString("\t")
			textOffset++
		case CharDrawingObj, CharInlineStart, CharFieldStart, CharBookmark, CharTitleMark, CharExtChar:
			// 컨트롤 문자 - 14바이트 추가 정보 (총 16바이트)
			if i+14 <= len(data) {
				ctrlID := binary.LittleEndian.Uint32(data[i : i+4])
				controls = append(controls, ControlInfo{
					Type:   char,
					ID:     ctrlID,
					Offset: textOffset,
				})
				i += 14
			}
		case CharFieldEnd:
			// 필드 끝
		case CharHyphen:
			sb.WriteString("-")
			textOffset++
		case CharNBSP, CharFixedWidthNBSP:
			sb.WriteString(" ")
			textOffset++
		default:
			// 일반 문자 (UTF-16LE)
			// Extended(1-3, 11-18, 21-23), Inline(4-9, 19-20)은 16바이트
			if char >= 0x0001 && char <= 0x001F {
				isExtended := (char >= 1 && char <= 3) || (char >= 11 && char <= 18) || (char >= 21 && char <= 23)
				isInline := (char >= 4 && char <= 9) || (char >= 19 && char <= 20)

				if isExtended || isInline {
					if i+14 <= len(data) {
						i += 14
					}
				}
			} else if char >= 0x0020 {
				sb.WriteRune(rune(char))
				textOffset++
			}
		}
	}

	return sb.String(), controls
}
