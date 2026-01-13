// Package hwp5 provides a parser for HWP 5.x binary documents.
package hwp5

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/richardlehane/mscfb"
	"github.com/roboco-io/hwp2md/internal/ir"
	"github.com/roboco-io/hwp2md/internal/parser"
)

// Parser parses HWP 5.x binary documents.
type Parser struct {
	path    string
	file    *os.File
	doc     *mscfb.Reader
	options parser.Options

	// Parsed data
	header     *FileHeader
	docInfo    *DocInfo
	sections   []string // Section stream names
	binDataDir string   // BinData storage path
}

// New creates a new HWP5 parser for the given file path.
func New(path string, opts parser.Options) (*Parser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("HWP 파일을 열 수 없습니다: %w", err)
	}

	doc, err := mscfb.New(f)
	if err != nil {
		f.Close()
		return nil, fmt.Errorf("OLE2 문서 파싱 실패: %w", err)
	}

	p := &Parser{
		path:    path,
		file:    f,
		doc:     doc,
		options: opts,
	}

	// Parse FileHeader
	if err := p.parseFileHeader(); err != nil {
		p.Close()
		return nil, err
	}

	// Find sections
	p.findSections()

	return p, nil
}

// Parse implements the Parser interface.
func (p *Parser) Parse() (*ir.Document, error) {
	doc := ir.NewDocument()

	// Parse DocInfo stream
	if err := p.parseDocInfo(); err != nil {
		return nil, fmt.Errorf("DocInfo 파싱 실패: %w", err)
	}

	// Set metadata
	doc.Metadata = p.buildMetadata()

	// Parse each section
	for _, sectionName := range p.sections {
		if err := p.parseSection(doc, sectionName); err != nil {
			return nil, fmt.Errorf("섹션 %s 파싱 실패: %w", sectionName, err)
		}
	}

	return doc, nil
}

// Close releases resources.
func (p *Parser) Close() error {
	if p.file != nil {
		return p.file.Close()
	}
	return nil
}

// parseFileHeader reads and parses the FileHeader stream.
func (p *Parser) parseFileHeader() error {
	data, err := p.readStream(StreamFileHeader)
	if err != nil {
		return fmt.Errorf("FileHeader 스트림을 읽을 수 없습니다: %w", err)
	}

	header, err := ParseFileHeader(data)
	if err != nil {
		return err
	}

	p.header = header

	// 암호화된 문서 확인
	if header.IsEncrypted() {
		return fmt.Errorf("암호화된 HWP 문서는 지원하지 않습니다")
	}

	// DRM 문서 확인
	if header.HasDRM() {
		return fmt.Errorf("DRM 보호된 HWP 문서는 지원하지 않습니다")
	}

	return nil
}

// findSections finds all Section streams in BodyText storage.
func (p *Parser) findSections() {
	for _, entry := range p.doc.File {
		name := entry.Name
		path := entry.Path
		fullPath := strings.Join(path, "/")

		// BodyText/SectionN 형식 찾기
		if strings.HasPrefix(fullPath, StreamBodyText) || strings.HasPrefix(name, "Section") {
			if strings.HasPrefix(name, "Section") {
				if fullPath != "" {
					p.sections = append(p.sections, fullPath+"/"+name)
				} else {
					p.sections = append(p.sections, name)
				}
			}
		}

		// BinData 저장소 위치 기억
		if name == StreamBinData {
			p.binDataDir = fullPath
		}
	}

	// 섹션 이름순 정렬
	sort.Strings(p.sections)

	// 섹션이 없으면 직접 찾기
	if len(p.sections) == 0 {
		for _, entry := range p.doc.File {
			if strings.HasPrefix(entry.Name, "Section") {
				p.sections = append(p.sections, entry.Name)
			}
		}
		sort.Strings(p.sections)
	}
}

// parseDocInfo reads and parses the DocInfo stream.
func (p *Parser) parseDocInfo() error {
	data, err := p.readStream(StreamDocInfo)
	if err != nil {
		return err
	}

	// 압축 해제 (필요시)
	if p.header.IsCompressed() {
		decompressed, err := DecompressStream(data)
		if err != nil {
			return fmt.Errorf("DocInfo 압축 해제 실패: %w", err)
		}
		data = decompressed
	}

	docInfo, err := ParseDocInfo(data)
	if err != nil {
		return err
	}

	p.docInfo = docInfo
	return nil
}

// parseSection parses a single section and adds content to the IR document.
func (p *Parser) parseSection(doc *ir.Document, sectionPath string) error {
	data, err := p.readStreamByPath(sectionPath)
	if err != nil {
		return err
	}

	// 압축 해제 (필요시)
	if p.header.IsCompressed() {
		decompressed, err := DecompressStream(data)
		if err != nil {
			return fmt.Errorf("섹션 압축 해제 실패: %w", err)
		}
		data = decompressed
	}

	sectionParser := NewSectionParser(p.docInfo)
	section, err := sectionParser.Parse(data)
	if err != nil {
		return err
	}

	// IR로 변환
	p.convertSectionToIR(doc, section)

	return nil
}

// convertSectionToIR converts a parsed section to IR blocks.
func (p *Parser) convertSectionToIR(doc *ir.Document, section *Section) {
	// 문단 변환
	for _, para := range section.Paragraphs {
		if para.Text == "" {
			continue
		}

		irPara := ir.NewParagraph(strings.TrimSpace(para.Text))
		doc.AddParagraph(irPara)
	}

	// 표 변환
	for _, table := range section.Tables {
		if table == nil || table.Rows == 0 || table.Cols == 0 {
			continue
		}

		irTable := ir.NewTable(table.Rows, table.Cols)

		for rowIdx, row := range table.Cells {
			for colIdx, cell := range row {
				if cell == nil {
					continue
				}

				// 셀 텍스트 추출
				var cellText strings.Builder
				for i, para := range cell.Paragraphs {
					if i > 0 {
						cellText.WriteString("\n")
					}
					cellText.WriteString(para.Text)
				}

				if rowIdx < len(irTable.Cells) && colIdx < len(irTable.Cells[rowIdx]) {
					irTable.Cells[rowIdx][colIdx].Text = strings.TrimSpace(cellText.String())
					irTable.Cells[rowIdx][colIdx].RowSpan = cell.RowSpan
					irTable.Cells[rowIdx][colIdx].ColSpan = cell.ColSpan
				}
			}
		}

		// 첫 행을 헤더로 설정
		if table.Rows > 1 {
			irTable.SetHeaderRow()
		}

		doc.AddTable(irTable)
	}

	// 이미지 변환 (옵션이 활성화된 경우)
	if p.options.ExtractImages {
		for _, img := range section.Images {
			irImg := ir.NewImage("")
			irImg.ID = fmt.Sprintf("BIN%04X", img.BinDataID)
			irImg.Width = int(img.Width / 7200) // EMU to pixels (approximate)
			irImg.Height = int(img.Height / 7200)

			// BinData에서 이미지 데이터 추출
			if p.docInfo != nil {
				for _, binData := range p.docInfo.BinDataList {
					if binData.BinDataID == img.BinDataID {
						irImg.Path = binData.GetBinDataPath()
						irImg.Format = strings.ToLower(binData.Extension)

						// 실제 이미지 데이터 추출
						if p.options.ImageDir != "" {
							p.extractImage(irImg, binData)
						}
						break
					}
				}
			}

			doc.AddImage(irImg)
		}
	}
}

// extractImage extracts image data from BinData storage.
func (p *Parser) extractImage(irImg *ir.ImageBlock, binData *BinDataInfo) {
	binPath := fmt.Sprintf("BinData/BIN%04X.%s", binData.BinDataID, binData.Extension)

	data, err := p.readStreamByPath(binPath)
	if err != nil {
		// 압축된 이름으로 재시도
		binPath = fmt.Sprintf("BinData/BIN%04X", binData.BinDataID)
		data, err = p.readStreamByPath(binPath)
		if err != nil {
			return
		}
	}

	// 압축 해제 (BinData는 개별 압축될 수 있음)
	if p.header.IsCompressed() {
		if decompressed, err := DecompressStream(data); err == nil {
			data = decompressed
		}
	}

	irImg.Data = data

	// 파일로 저장
	if p.options.ImageDir != "" {
		outPath := filepath.Join(p.options.ImageDir, irImg.Path)
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err == nil {
			_ = os.WriteFile(outPath, data, 0644)
			irImg.Path = outPath
		}
	}
}

// buildMetadata builds IR metadata from DocInfo.
func (p *Parser) buildMetadata() ir.Metadata {
	meta := ir.Metadata{}

	// 파일명에서 제목 추출
	base := filepath.Base(p.path)
	meta.Title = strings.TrimSuffix(base, filepath.Ext(base))

	// 버전 정보
	if p.header != nil {
		meta.Creator = fmt.Sprintf("HWP %s", p.header.Version.String())
	}

	return meta
}

// readStream reads a stream by name from the root.
func (p *Parser) readStream(name string) ([]byte, error) {
	for _, entry := range p.doc.File {
		if entry.Name == name {
			return io.ReadAll(entry)
		}
	}
	return nil, fmt.Errorf("스트림을 찾을 수 없습니다: %s", name)
}

// readStreamByPath reads a stream by full path.
func (p *Parser) readStreamByPath(streamPath string) ([]byte, error) {
	// 경로 정규화
	streamPath = strings.TrimPrefix(streamPath, "/")
	parts := strings.Split(streamPath, "/")

	for _, entry := range p.doc.File {
		entryPath := strings.Join(append(entry.Path, entry.Name), "/")
		if entryPath == streamPath || entry.Name == streamPath {
			return io.ReadAll(entry)
		}

		// 부분 매칭 (BodyText/Section0 형식)
		if len(parts) > 0 && entry.Name == parts[len(parts)-1] {
			if len(parts) == 1 || (len(entry.Path) > 0 && entry.Path[len(entry.Path)-1] == parts[len(parts)-2]) {
				return io.ReadAll(entry)
			}
		}
	}

	return nil, fmt.Errorf("스트림을 찾을 수 없습니다: %s", streamPath)
}

// GetHeader returns the parsed file header.
func (p *Parser) GetHeader() *FileHeader {
	return p.header
}

// GetDocInfo returns the parsed document info.
func (p *Parser) GetDocInfo() *DocInfo {
	return p.docInfo
}

// IsCompressed returns true if the document is compressed.
func (p *Parser) IsCompressed() bool {
	return p.header != nil && p.header.IsCompressed()
}

// GetVersion returns the HWP version string.
func (p *Parser) GetVersion() string {
	if p.header != nil {
		return p.header.Version.String()
	}
	return ""
}

// ExtractImages extracts all images to the specified directory.
func (p *Parser) ExtractImages(dir string) ([]ir.ImageBlock, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("이미지 디렉토리 생성 실패: %w", err)
	}

	var images []ir.ImageBlock

	if p.docInfo == nil {
		return images, nil
	}

	for _, binData := range p.docInfo.BinDataList {
		if binData.BinDataID == 0 {
			continue
		}

		binPath := fmt.Sprintf("BinData/BIN%04X.%s", binData.BinDataID, binData.Extension)
		data, err := p.readStreamByPath(binPath)
		if err != nil {
			continue
		}

		// 압축 해제
		if p.header.IsCompressed() {
			if decompressed, err := DecompressStream(data); err == nil {
				data = decompressed
			}
		}

		filename := binData.GetBinDataPath()
		outPath := filepath.Join(dir, filename)

		if err := os.WriteFile(outPath, data, 0644); err != nil {
			continue
		}

		img := ir.ImageBlock{
			ID:     fmt.Sprintf("BIN%04X", binData.BinDataID),
			Path:   outPath,
			Format: strings.ToLower(binData.Extension),
		}
		images = append(images, img)
	}

	return images, nil
}
