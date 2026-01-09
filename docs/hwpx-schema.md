# HWPX 파일 포맷 스키마

## 1. 개요

HWPX는 **OWPML (Open Word Processor Markup Language)** 기반의 ZIP 압축 파일로, KS X 6101 표준을 따릅니다. DOCX와 유사한 구조를 가지며, XML 파일들의 패키지로 구성됩니다.

---

## 2. ZIP 내부 디렉토리 구조

```
document.hwpx (ZIP)
├── content.hpf              # 패키지 매니페스트 (진입점)
├── Contents/
│   ├── header.xml           # 문서 전역 정보 (스타일, 페이지 설정)
│   ├── section0.xml         # 섹션 0 본문 (문단, 표, 이미지)
│   ├── section1.xml         # 섹션 1 본문 (선택적)
│   └── ...
├── Meta/
│   └── meta.xml             # 문서 메타데이터 (작성자, 생성일)
├── BinData/
│   ├── bin0.png             # 임베디드 이미지
│   ├── bin1.jpg
│   └── ...
└── Settings/
    └── settings.xml         # 문서/앱 설정 (선택적)
```

### 텍스트 추출에 필요한 파일

| 파일 | 필수 | 용도 |
|------|------|------|
| `content.hpf` | O | 패키지 매니페스트, 파트 목록 |
| `Contents/section*.xml` | O | 본문 콘텐츠 (문단, 표, 이미지) |
| `Contents/header.xml` | △ | 스타일 정의, 번호 매기기 |
| `BinData/*` | △ | 이미지 추출 시 필요 |

---

## 3. 주요 XML 파일

### 3.1 content.hpf

패키지 매니페스트 파일로, 모든 파트 파일을 선언합니다.

```xml
<?xml version="1.0" encoding="UTF-8"?>
<pkg:package xmlns:pkg="http://www.hancom.co.kr/hwpml/2011/package">
  <pkg:parts>
    <pkg:part name="/Contents/header.xml" type="header"/>
    <pkg:part name="/Contents/section0.xml" type="section"/>
    <pkg:part name="/Contents/section1.xml" type="section"/>
    <pkg:part name="/BinData/bin0.png" type="binary"/>
  </pkg:parts>
</pkg:package>
```

**파서 구현:**
- `section*.xml` 파일 목록 추출
- BinData ID → 파일 경로 매핑

### 3.2 header.xml

문서 전역 설정을 포함합니다.

```xml
<hh:header xmlns:hh="http://www.hancom.co.kr/hwpml/2011/head">
  <hh:docInfo>
    <hh:title>문서 제목</hh:title>
  </hh:docInfo>
  <hh:styles>
    <hh:style styleId="0" name="본문" type="para"/>
    <hh:style styleId="1" name="제목 1" type="para" outlineLevel="1"/>
  </hh:styles>
  <hh:numberings>
    <hh:num numId="1" listType="bullet">
      <hh:lvl ilvl="0" text="•"/>
      <hh:lvl ilvl="1" text="◦"/>
    </hh:num>
  </hh:numberings>
</hh:header>
```

### 3.3 sectionN.xml

실제 본문 콘텐츠를 포함합니다.

```xml
<hs:section xmlns:hs="http://www.hancom.co.kr/hwpml/2011/section"
            xmlns:hp="http://www.hancom.co.kr/hwpml/2011/paragraph"
            xmlns:ht="http://www.hancom.co.kr/hwpml/2011/table">
  <hp:p paraId="1" styleIDRef="0">
    <hp:run>
      <hp:t>일반 텍스트입니다.</hp:t>
    </hp:run>
  </hp:p>
  <ht:tbl>
    <!-- 표 내용 -->
  </ht:tbl>
</hs:section>
```

---

## 4. XML 네임스페이스

| 접두사 | 네임스페이스 URI | 용도 |
|--------|------------------|------|
| `pkg` | `http://www.hancom.co.kr/hwpml/2011/package` | 패키지 매니페스트 |
| `hh` | `http://www.hancom.co.kr/hwpml/2011/head` | 헤더/스타일 |
| `hs` | `http://www.hancom.co.kr/hwpml/2011/section` | 섹션 |
| `hp` | `http://www.hancom.co.kr/hwpml/2011/paragraph` | 문단 |
| `ht` | `http://www.hancom.co.kr/hwpml/2011/table` | 표 |
| `hd` | `http://www.hancom.co.kr/hwpml/2011/drawing` | 이미지/도형 |
| `hl` | `http://www.hancom.co.kr/hwpml/2011/list` | 목록/번호 매기기 |

**참고:** 네임스페이스 URI는 버전에 따라 다를 수 있으므로, 파서는 **로컬 이름(local-name)**으로 매칭하는 것이 안전합니다.

---

## 5. 핵심 XML 요소

### 5.1 문단 (Paragraph)

```xml
<hp:p paraId="1" paraPrIDRef="2" styleIDRef="1" outlineLevel="1">
  <hp:run charPrIDRef="3">
    <hp:t>텍스트 내용</hp:t>
  </hp:run>
  <hp:run>
    <hp:tab/>
  </hp:run>
  <hp:run>
    <hp:br type="line"/>
  </hp:run>
</hp:p>
```

| 요소 | 설명 |
|------|------|
| `hp:p` | 문단 컨테이너 |
| `hp:run` | 텍스트 런 (동일 스타일 그룹) |
| `hp:t` | 실제 텍스트 내용 |
| `hp:tab` | 탭 문자 (`\t`) |
| `hp:br` | 줄바꿈 (`type="line"`) 또는 페이지 나눔 |

**속성:**

| 속성 | 요소 | 설명 |
|------|------|------|
| `paraId` | `hp:p` | 문단 고유 ID |
| `styleIDRef` | `hp:p` | 스타일 참조 ID |
| `outlineLevel` | `hp:p` | 개요 수준 (1-9, 제목용) |
| `charPrIDRef` | `hp:run` | 문자 속성 참조 ID |

### 5.2 표 (Table)

```xml
<ht:tbl tblPrIDRef="1">
  <ht:tr>
    <ht:tc gridSpan="1" rowSpan="1">
      <hp:p>
        <hp:run><hp:t>셀 1</hp:t></hp:run>
      </hp:p>
    </ht:tc>
    <ht:tc>
      <hp:p>
        <hp:run><hp:t>셀 2</hp:t></hp:run>
      </hp:p>
    </ht:tc>
  </ht:tr>
</ht:tbl>
```

| 요소 | 설명 |
|------|------|
| `ht:tbl` | 표 컨테이너 |
| `ht:tr` | 표 행 |
| `ht:tc` | 표 셀 (내부에 `hp:p` 포함) |

**속성:**

| 속성 | 요소 | 설명 |
|------|------|------|
| `gridSpan` | `ht:tc` | 열 병합 수 |
| `rowSpan` | `ht:tc` | 행 병합 수 |

### 5.3 이미지 (Image)

```xml
<hd:pic>
  <hd:shapePr width="100" height="80"/>
  <hd:img binItemIDRef="bin0" alt="이미지 설명"/>
</hd:pic>
```

| 요소 | 설명 |
|------|------|
| `hd:pic` | 그림 컨테이너 |
| `hd:img` | 이미지 참조 |

**속성:**

| 속성 | 요소 | 설명 |
|------|------|------|
| `binItemIDRef` | `hd:img` | BinData 파일 참조 ID |
| `alt` | `hd:img` | 대체 텍스트 |
| `width`, `height` | `hd:shapePr` | 크기 (픽셀) |

### 5.4 목록 (List)

목록은 문단 속성으로 표현됩니다:

```xml
<hp:p numId="1" ilvl="0">
  <hp:run><hp:t>첫 번째 항목</hp:t></hp:run>
</hp:p>
<hp:p numId="1" ilvl="1">
  <hp:run><hp:t>중첩 항목</hp:t></hp:run>
</hp:p>
```

**속성:**

| 속성 | 설명 |
|------|------|
| `numId` | 목록 정의 ID |
| `ilvl` | 들여쓰기 레벨 (0부터 시작) |

**목록 정의 (header.xml):**

```xml
<hl:num numId="1" listType="bullet">
  <hl:lvl ilvl="0" text="•"/>
  <hl:lvl ilvl="1" text="◦"/>
</hl:num>
<hl:num numId="2" listType="decimal">
  <hl:lvl ilvl="0" format="%1."/>
</hl:num>
```

---

## 6. 텍스트 추출 알고리즘

```
1. HWPX 파일을 ZIP으로 압축 해제
2. content.hpf 파싱하여 section*.xml 목록 획득
3. 각 section*.xml에 대해:
   a. XML 파싱
   b. 모든 블록 요소 순회:
      - hp:p (문단): 텍스트 추출
      - ht:tbl (표): 각 셀의 텍스트 추출
      - hd:pic (이미지): 이미지 참조 기록
   c. 문단 내 요소 처리:
      - hp:t: 텍스트 추가
      - hp:tab: '\t' 추가
      - hp:br (type="line"): '\n' 추가
4. 결과를 IR (Intermediate Representation)로 변환
```

---

## 7. 참고 자료

- [나무위키 - HWP](https://en.namu.wiki/w/HWP)
- [Just Solve - HWP](http://justsolve.archiveteam.org/wiki/HWP)
- [FireEye HWP Zero-Day Analysis (PDF)](https://media.kasperskycontenthub.com/wp-content/uploads/sites/43/2016/02/20081603/FireEye_HWP_ZeroDay.pdf)
- KS X 6101 - OWPML 표준
