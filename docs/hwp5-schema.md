# HWP 5.x 파일 포맷 스키마

## 1. 개요

HWP 5.x는 **OLE2 Compound File Binary (CFB)** 컨테이너 기반의 바이너리 문서 포맷입니다. Microsoft Office의 DOC, XLS 파일과 동일한 컨테이너 구조를 사용하며, 내부에 여러 스트림(Stream)과 저장소(Storage)를 포함합니다.

---

## 2. OLE2 내부 디렉토리 구조

```
document.hwp (OLE2 CFB)
├── FileHeader          # 파일 인식 정보 (고정 256바이트)
├── DocInfo             # 문서 정보 (압축 가능)
│   ├── ID_MAPPINGS     # ID 매핑 테이블
│   ├── BIN_DATA        # 바이너리 데이터 정보
│   ├── FACE_NAME       # 글꼴 이름
│   ├── CHAR_SHAPE      # 글자 모양
│   ├── PARA_SHAPE      # 문단 모양
│   └── ...
├── BodyText/           # 본문 저장소
│   ├── Section0        # 첫 번째 섹션 (압축 가능)
│   ├── Section1        # 두 번째 섹션
│   └── ...
├── BinData/            # 바이너리 데이터 저장소
│   ├── BIN0001.jpg     # 이미지 등
│   ├── BIN0002.png
│   └── ...
├── PrvText             # 미리보기 텍스트
├── PrvImage            # 미리보기 이미지
└── Scripts/            # 스크립트 저장소 (선택적)
```

### 텍스트 추출에 필요한 스트림

| 스트림 | 필수 | 용도 |
|--------|------|------|
| `FileHeader` | O | 파일 버전, 압축/암호화 플래그 확인 |
| `DocInfo` | O | ID 매핑 테이블, 글꼴/스타일 정보 |
| `BodyText/Section*` | O | 본문 콘텐츠 (문단, 표, 이미지 참조) |
| `BinData/*` | △ | 이미지 추출 시 필요 |

---

## 3. FileHeader 구조

FileHeader는 고정 256바이트 크기이며, 압축/암호화되지 않습니다.

```
┌──────────────────────────────────────────────────────────────────┐
│ Offset │ Size │ Description                                      │
├────────┼──────┼──────────────────────────────────────────────────┤
│ 0      │ 32   │ 시그니처: "HWP Document File" (null-padded)       │
│ 32     │ 4    │ 파일 버전 (예: 5.1.0.0 → 0x050100)               │
│ 36     │ 4    │ 속성 플래그                                       │
│ 40     │ 216  │ 예약 영역                                         │
└──────────────────────────────────────────────────────────────────┘
```

### 속성 플래그 (비트 필드)

| 비트 | 설명 |
|------|------|
| 0 | 압축 여부 |
| 1 | 암호화 여부 |
| 2 | 배포용 문서 |
| 3 | 스크립트 저장 |
| 4 | DRM 보안 |
| 5 | XMLTemplate 저장 |
| 6 | 문서 이력 관리 |
| 7 | 전자 서명 |
| 8 | 공인 인증서 암호화 |
| 9 | 전자 서명 예비 |
| 10 | 공인 인증서 DRM |
| 11 | CCL 문서 |

---

## 4. 레코드 구조

DocInfo와 BodyText/Section 스트림은 **레코드(Record)** 단위로 데이터가 저장됩니다.

### 레코드 헤더 (4바이트)

```
┌─────────────────────────────────────────────────────────────────┐
│  Tag ID (10비트)  │  Level (10비트)  │  Size (12비트)            │
│  [31:22]          │  [21:12]         │  [11:0]                   │
└─────────────────────────────────────────────────────────────────┘
```

- **Tag ID**: 레코드 종류 식별자 (0-1023)
- **Level**: 논리적 계층 표현 (0부터 시작)
- **Size**: 데이터 길이 (바이트)
  - 4095 이상이면 `0xFFF`로 설정하고 다음 4바이트에 실제 크기

### 주요 레코드 태그

| 태그 | 값 (hex) | 값 (dec) | 설명 |
|------|----------|----------|------|
| DOCUMENT_PROPERTIES | 0x10 | 16 | 문서 속성 |
| ID_MAPPINGS | 0x11 | 17 | ID 매핑 테이블 |
| BIN_DATA | 0x12 | 18 | 바이너리 데이터 정보 |
| FACE_NAME | 0x13 | 19 | 글꼴 이름 |
| BORDER_FILL | 0x14 | 20 | 테두리/배경 |
| CHAR_SHAPE | 0x15 | 21 | 글자 모양 |
| TAB_DEF | 0x16 | 22 | 탭 정의 |
| NUMBERING | 0x17 | 23 | 문단 번호 |
| BULLET | 0x18 | 24 | 글머리표 |
| PARA_SHAPE | 0x19 | 25 | 문단 모양 |
| STYLE | 0x1A | 26 | 스타일 |
| **PARA_HEADER** | 0x42 | 66 | 문단 헤더 |
| **PARA_TEXT** | 0x43 | 67 | 문단 텍스트 |
| PARA_CHAR_SHAPE | 0x44 | 68 | 문단 글자 모양 |
| PARA_LINE_SEG | 0x45 | 69 | 문단 레이아웃 |
| **CTRL_HEADER** | 0x47 | 71 | 컨트롤 헤더 |
| **TABLE** | 0x4D | 77 | 표 |
| **LIST_HEADER** | 0x48 | 72 | 리스트 헤더 (셀) |

---

## 5. 핵심 레코드 상세

### 5.1 PARA_HEADER (문단 헤더)

문단의 시작을 나타냅니다.

| 오프셋 | 크기 | 설명 |
|--------|------|------|
| 0 | 4 | 문단 텍스트 문자 수 |
| 4 | 2 | 문단 모양 ID (ParaShapeID) |
| 6 | 1 | 스타일 ID (StyleID) |
| 7 | 1 | 분할 유형 (DivisionType) |
| 8 | 2 | 글자 모양 적용 수 |
| 10 | 2 | 범위 태그 수 |
| 12 | 2 | 줄 정렬 수 |
| 14 | 4 | 인스턴스 ID |

### 5.2 PARA_TEXT (문단 텍스트)

UTF-16LE로 인코딩된 텍스트와 인라인 컨트롤 문자를 포함합니다.

#### 컨트롤 문자

| 값 (hex) | 설명 | 추가 데이터 |
|----------|------|-------------|
| 0x0001 | Extended (예약) | +14바이트 |
| 0x0002 | 구역/단 정의 | +14바이트 |
| 0x0003 | 필드 시작 | +14바이트 |
| 0x0004-0x0009 | Inline 컨트롤 | +14바이트 |
| 0x000A | 줄바꿈 (Line Break) | 없음 |
| 0x000B | 그리기 개체/표 | +14바이트 |
| 0x000D | 문단 끝 (Paragraph End) | 없음 |
| 0x0010 | 하이픈 | 없음 |
| 0x001E | NBSP (고정 너비) | 없음 |
| 0x001F | NBSP | 없음 |

#### 텍스트 추출 알고리즘

```go
for i := 0; i < len(data); i += 2 {
    char := uint16(data[i]) | uint16(data[i+1])<<8

    switch {
    case char == 0x000A:
        // 줄바꿈 추가
    case char == 0x000D:
        // 문단 끝 (무시)
    case char >= 0x0001 && char <= 0x001F:
        // Extended/Inline 컨트롤: 14바이트 건너뛰기
        if isExtendedOrInline(char) {
            i += 14
        }
    default:
        // 일반 문자 추가
    }
}
```

### 5.3 CTRL_HEADER (컨트롤 헤더)

표, 이미지, 각주 등의 컨트롤을 나타냅니다.

| 오프셋 | 크기 | 설명 |
|--------|------|------|
| 0 | 4 | 컨트롤 ID (4문자, 역순) |
| 4 | ... | 컨트롤별 추가 데이터 |

#### 컨트롤 ID

| ID | 역순 표기 | 설명 |
|----|-----------|------|
| `tbl ` | ` lbt` | 표 (Table) |
| `gso ` | ` osg` | 그리기 개체 (GSO) |
| `secd` | `dces` | 구역 정의 (Section Def) |
| `cold` | `dloc` | 단 정의 (Column Def) |
| `head` | `daeh` | 머리말 (Header) |
| `foot` | `toof` | 꼬리말 (Footer) |

### 5.4 TABLE (표)

표의 구조 정보를 포함합니다.

| 오프셋 | 크기 | 설명 |
|--------|------|------|
| 0 | 4 | 속성 플래그 |
| 4 | 2 | 행 수 (Rows) |
| 6 | 2 | 열 수 (Cols) |
| 8 | 2 | 셀 간격 |
| 10 | 2 | 왼쪽 여백 |
| 12 | 2 | 오른쪽 여백 |
| 14 | 2 | 위 여백 |
| 16 | 2 | 아래 여백 |

### 5.5 LIST_HEADER (셀)

표 셀 또는 리스트 항목을 나타냅니다.

| 오프셋 | 크기 | 설명 |
|--------|------|------|
| 0 | 2 | 문단 수 (ParaCount) |
| 2 | 4 | 속성 |
| ... | ... | 추가 데이터 |

---

## 6. 테이블 파싱 구조

HWP5에서 테이블은 다음과 같은 레코드 계층으로 표현됩니다:

```
PARA_HEADER (Level=0)         # 본문 문단
├── PARA_TEXT (Level=1)       # 문단 텍스트 (컨트롤 코드 포함)
├── CTRL_HEADER " lbt" (Level=1)  # 테이블 시작
│   ├── TABLE (Level=2)           # 테이블 속성 (행/열 수)
│   ├── LIST_HEADER (Level=2)     # 셀 1
│   │   ├── PARA_HEADER (Level=2)
│   │   └── PARA_TEXT (Level=3)
│   ├── LIST_HEADER (Level=2)     # 셀 2
│   │   ├── PARA_HEADER (Level=2)
│   │   └── PARA_TEXT (Level=3)
│   └── ...
PARA_HEADER (Level=0)         # 다음 본문 문단
```

### 테이블 파싱 알고리즘

1. `CTRL_HEADER`에서 컨트롤 ID가 ` lbt` (또는 `tbl `)인 경우 테이블 시작
2. `TABLE` 레코드에서 행/열 수 추출
3. `LIST_HEADER`마다 하나의 셀로 인식
4. 각 셀 내의 `PARA_HEADER`/`PARA_TEXT`에서 텍스트 추출
5. 셀을 행 우선으로 2D 배열에 배치

---

## 7. 압축 및 암호화

### 압축

- **알고리즘**: zlib (DEFLATE)
- **적용 대상**: DocInfo, BodyText/Section*, BinData
- **참고**: 일부 HWP 파일은 zlib 헤더 없이 raw deflate 사용

```go
// 압축 해제 시도 순서
1. zlib (헤더 0x78 확인)
2. raw deflate (flate.NewReader)
```

### 암호화

- 일반 비밀번호 암호화
- 공인인증서 기반 암호화
- DRM

현재 hwp2md는 암호화된 문서를 지원하지 않습니다.

---

## 8. IR 변환 매핑

| HWP5 레코드 | IR 타입 |
|-------------|---------|
| PARA_HEADER + PARA_TEXT | `ir.Paragraph` |
| CTRL_HEADER(" lbt") + TABLE | `ir.TableBlock` |
| LIST_HEADER | `ir.TableCell` |
| BinData 이미지 | `ir.ImageBlock` |

---

## 9. 구현 파일

| 파일 | 용도 |
|------|------|
| `internal/parser/hwp5/constants.go` | 상수 정의 (태그 ID, 컨트롤 코드) |
| `internal/parser/hwp5/header.go` | FileHeader 파싱 |
| `internal/parser/hwp5/record.go` | 레코드 파싱, 압축 해제 |
| `internal/parser/hwp5/text.go` | UTF-16LE 텍스트 추출 |
| `internal/parser/hwp5/docinfo.go` | DocInfo 스트림 파싱 |
| `internal/parser/hwp5/section.go` | Section 파싱 (문단, 테이블) |
| `internal/parser/hwp5/parser.go` | 메인 파서, IR 변환 |

---

## 10. 참고 자료

- [한글 문서 파일 형식 5.0 명세서](https://cdn.hancom.com/link/docs/한글문서파일형식_5.0_revision1.3.pdf)
- [한컴 기술 블로그](https://tech.hancom.com/python-hwp-parsing-1/)
- [pyhwp](https://github.com/mete0r/pyhwp) - Python HWP 파서
- [hwp.js](https://github.com/hahnlee/hwp.js) - TypeScript HWP 파서
- [mscfb](https://github.com/richardlehane/mscfb) - Go OLE2 라이브러리
