# HWP 5.x 파일 포맷 조사 결과

## 개요

HWP 5.x는 한글과컴퓨터의 한글 워드프로세서에서 사용하는 바이너리 문서 포맷이다. 한글 2002부터 한글 2022까지 사용되었으며, 2014년부터는 XML 기반의 HWPX 포맷이 병행 지원된다.

### 구현 상태

| 기능 | 상태 | 설명 |
|------|------|------|
| OLE2 파싱 | ✅ 완료 | mscfb 라이브러리 사용 |
| FileHeader | ✅ 완료 | 버전, 압축, 암호화 플래그 파싱 |
| DocInfo | ✅ 완료 | ID 매핑, 글꼴, 스타일 정보 |
| BodyText/Section | ✅ 완료 | 문단, 텍스트 추출 |
| 테이블 | ✅ 완료 | 행/열, 셀 병합, 셀 내용 |
| 이미지 | ⚠️ 부분 | BinData 참조만 지원 |
| 압축 해제 | ✅ 완료 | zlib/deflate 지원 |
| 암호화 | ❌ 미지원 | 암호화 문서 차단 |
| DRM | ❌ 미지원 | DRM 문서 차단 |

## 공식 문서

### 한글 문서 파일 형식 5.0 명세서
- **URL**: https://cdn.hancom.com/link/docs/한글문서파일형식_5.0_revision1.3.pdf
- **버전**: Revision 1.3
- **발행**: 한글과컴퓨터

### 한컴 기술 블로그
- https://tech.hancom.com/python-hwp-parsing-1/
- https://tech.hancom.com/한-글-문서-파일-형식-hwp-포맷-구조-살펴보기/

## 파일 구조

### OLE2 Compound File Binary (CFB) 컨테이너

HWP 5.x 파일은 Microsoft의 OLE2 Compound File 형식을 기반으로 한다. 이는 파일 내부에 여러 스트림(Stream)과 저장소(Storage)를 포함하는 계층적 구조이다.

```
HWP 파일
├── FileHeader          # 파일 인식 정보 (고정 256바이트)
├── DocInfo             # 문서 정보 (압축/암호화 가능)
├── BodyText/           # 본문 저장소
│   ├── Section0        # 첫 번째 섹션
│   ├── Section1        # 두 번째 섹션
│   └── ...
├── BinData/            # 바이너리 데이터 저장소
│   ├── BIN0001.jpg     # 이미지 등
│   └── ...
├── PrvText             # 미리보기 텍스트
├── PrvImage            # 미리보기 이미지
└── Scripts/            # 스크립트 저장소 (선택)
```

### FileHeader 구조

| 오프셋 | 크기 | 설명 |
|--------|------|------|
| 0 | 32 | 시그니처: "HWP Document File" |
| 32 | 4 | 파일 버전 (예: 5.0.3.0) |
| 36 | 4 | 속성 플래그 |
| ... | ... | 예약 영역 |

**속성 플래그**:
- 비트 0: 압축 여부
- 비트 1: 암호화 여부
- 비트 2: 배포용 문서
- 비트 3: 스크립트 저장
- 비트 4: DRM 보안
- 비트 5: XMLTemplate 저장
- 비트 6: 문서 이력 관리
- 비트 7: 전자 서명
- 비트 8: 공인 인증서 암호화
- 비트 9: 전자 서명 예비
- 비트 10: 공인 인증서 DRM
- 비트 11: CCL 문서

### 레코드 구조

DocInfo와 BodyText/Section 스트림은 레코드(Record) 단위로 데이터가 저장된다.

**레코드 헤더 (4바이트)**:
```
┌─────────────────────────────────────────────────────────────┐
│  Tag ID (10비트)  │  Level (10비트)  │  Size (12비트)        │
└─────────────────────────────────────────────────────────────┘
```

- **Tag ID**: 레코드 종류 식별자
- **Level**: 논리적 계층 표현 (0부터 시작)
- **Size**: 데이터 길이 (바이트), 4095 이상이면 0xFFF로 설정하고 다음 4바이트에 실제 크기

**주요 레코드 태그**:
| 태그 | 값 | 설명 |
|------|-----|------|
| HWPTAG_DOCUMENT_PROPERTIES | 0x0010 | 문서 속성 |
| HWPTAG_ID_MAPPINGS | 0x0011 | ID 매핑 테이블 |
| HWPTAG_BIN_DATA | 0x0012 | 바이너리 데이터 정보 |
| HWPTAG_FACE_NAME | 0x0013 | 글꼴 이름 |
| HWPTAG_BORDER_FILL | 0x0014 | 테두리/배경 |
| HWPTAG_CHAR_SHAPE | 0x0015 | 글자 모양 |
| HWPTAG_TAB_DEF | 0x0016 | 탭 정의 |
| HWPTAG_NUMBERING | 0x0017 | 문단 번호 |
| HWPTAG_BULLET | 0x0018 | 글머리표 |
| HWPTAG_PARA_SHAPE | 0x0019 | 문단 모양 |
| HWPTAG_STYLE | 0x001A | 스타일 |
| HWPTAG_PARA_HEADER | 0x0042 | 문단 헤더 |
| HWPTAG_PARA_TEXT | 0x0043 | 문단 텍스트 |
| HWPTAG_PARA_CHAR_SHAPE | 0x0044 | 문단 글자 모양 |
| HWPTAG_PARA_LINE_SEG | 0x0045 | 문단 레이아웃 |
| HWPTAG_CTRL_HEADER | 0x0046 | 컨트롤 헤더 |
| HWPTAG_TABLE | 0x0050 | 표 |
| HWPTAG_LIST_HEADER | 0x0051 | 리스트 헤더 |

### 압축 및 암호화

- **압축**: zlib (DEFLATE) 알고리즘 사용
- **암호화**: 다양한 암호화 방식 지원
  - 일반 비밀번호 암호화
  - 공인인증서 기반 암호화
  - DRM

## 오픈소스 구현체

### Go 언어

현재 Go로 작성된 완전한 HWP 5.x 파서는 찾지 못함. 구현 시 참고할 수 있는 프로젝트들:

### Rust

| 프로젝트 | URL | 특징 |
|----------|-----|------|
| **OpenHWP** | https://github.com/openhwp/openhwp | HWP 5.0 읽기, HWPX 읽기/쓰기, IR 변환 지원 |
| **unhwp** | https://lib.rs/crates/unhwp | HWP 5.0+, HWPX, HWP 3.x 지원, Markdown 출력 |

### Python

| 프로젝트 | URL | 특징 |
|----------|-----|------|
| **pyhwp** | https://github.com/mete0r/pyhwp | 가장 성숙한 구현체, ODT/TXT 변환, HWP Binary Specification 1.1 기반 |
| **hwpers** | https://github.com/Indosaram/hwpers | HWP 5.0 완전 지원 |
| **hwp-extract** | https://github.com/volexity/hwp-extract | 암호화된 HWP 지원, 메타데이터 추출 |

### JavaScript/TypeScript

| 프로젝트 | URL | 특징 |
|----------|-----|------|
| **hwp.js** | https://github.com/hahnlee/hwp.js | TypeScript 기반, 웹 뷰어, cfb-js로 OLE 파싱 |
| **hwp-parser** | https://github.com/BOB-APT-Solution/hwp-parser | JS 코드 추출 등 보안 분석용 |

## 구현 전략 권장사항

### 1. OLE2 파싱

Go에서 OLE2 Compound File을 파싱하기 위한 라이브러리:
- 직접 구현 또는 기존 라이브러리 활용 필요
- 참고: https://github.com/richardlehane/mscfb (Go OLE2 라이브러리)

### 2. 레코드 파싱

1. FileHeader 읽기 → 버전/속성 확인
2. DocInfo 스트림 압축 해제 → 레코드 순회 → ID 매핑 테이블 구축
3. BodyText/SectionN 스트림 압축 해제 → 레코드 순회 → IR 변환

### 3. IR 변환

현재 HWPX 파서와 동일한 IR 구조(`internal/ir/`)로 변환:
- `HWPTAG_PARA_HEADER` + `HWPTAG_PARA_TEXT` → `ir.Paragraph`
- `HWPTAG_TABLE` + 셀 데이터 → `ir.Table`
- BinData의 이미지 → `ir.Image`

### 4. 참고 구현체

가장 참고하기 좋은 구현체:
1. **pyhwp** (Python): 가장 완성도 높고 문서화 잘 됨
2. **OpenHWP** (Rust): 최신 구현, IR 변환 아키텍처 참고
3. **hwp.js** (TypeScript): 웹 기반 구현 참고

## 구현 완료 항목

1. [x] Go OLE2 라이브러리 선정 → `github.com/richardlehane/mscfb` 사용
2. [x] FileHeader 파싱 구현 → `internal/parser/hwp5/header.go`
3. [x] DocInfo 스트림 파싱 (레코드 구조) → `internal/parser/hwp5/docinfo.go`
4. [x] BodyText/Section 파싱 → `internal/parser/hwp5/section.go`
5. [x] IR 변환 로직 구현 → `internal/parser/hwp5/parser.go`
6. [x] 테이블 파싱 구현 → 셀 병합, 셀 내용 포함
7. [x] 테스트 케이스 작성 → `internal/parser/hwp5/*_test.go`, `tests/e2e_test.go`

## 향후 개선 사항

1. [ ] BinData 이미지 추출 개선
2. [ ] 다양한 HWP 버전 테스트 (5.0.0.0 ~ 5.1.1.0)
3. [ ] 복잡한 테이블 구조 지원 (중첩 테이블 등)

## 참고 자료

- 한글 문서 파일 형식 5.0: https://cdn.hancom.com/link/docs/한글문서파일형식_5.0_revision1.3.pdf
- 한컴 기술 블로그: https://tech.hancom.com/python-hwp-parsing-1/
- pyhwp 문서: https://pyhwp.readthedocs.io/
- OpenHWP: https://github.com/openhwp/openhwp
- unhwp: https://lib.rs/crates/unhwp
- hwp.js: https://github.com/hahnlee/hwp.js
- mscfb (Go OLE2): https://github.com/richardlehane/mscfb
