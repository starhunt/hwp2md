// Package ir defines the Intermediate Representation for HWP documents.
// IR is the output of Stage 1 (Parser) and input for Stage 2 (LLM Formatter).
package ir

// Document represents the intermediate representation of an HWP document.
type Document struct {
	Version     string   `json:"version"`
	Metadata    Metadata `json:"metadata"`
	Content     []Block  `json:"content"`
	RawMarkdown string   `json:"raw_markdown,omitempty"` // Pre-rendered markdown from external parser (e.g., Upstage)
}

// Metadata contains document metadata.
type Metadata struct {
	Title       string `json:"title,omitempty"`
	Author      string `json:"author,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Keywords    string `json:"keywords,omitempty"`
	Description string `json:"description,omitempty"`
	Creator     string `json:"creator,omitempty"`
	Created     string `json:"created,omitempty"`
	Modified    string `json:"modified,omitempty"`
}

// BlockType represents the type of content block.
type BlockType string

const (
	BlockTypeParagraph BlockType = "paragraph"
	BlockTypeTable     BlockType = "table"
	BlockTypeImage     BlockType = "image"
	BlockTypeList      BlockType = "list"
)

// Block represents a content block in the document.
type Block struct {
	Type      BlockType   `json:"type"`
	Paragraph *Paragraph  `json:"paragraph,omitempty"`
	Table     *TableBlock `json:"table,omitempty"`
	Image     *ImageBlock `json:"image,omitempty"`
	List      *ListBlock  `json:"list,omitempty"`
}

// NewDocument creates a new IR document with the current version.
func NewDocument() *Document {
	return &Document{
		Version: "1.0",
		Content: make([]Block, 0),
	}
}

// AddParagraph adds a paragraph block to the document.
func (d *Document) AddParagraph(p *Paragraph) {
	d.Content = append(d.Content, Block{
		Type:      BlockTypeParagraph,
		Paragraph: p,
	})
}

// AddTable adds a table block to the document.
func (d *Document) AddTable(t *TableBlock) {
	d.Content = append(d.Content, Block{
		Type:  BlockTypeTable,
		Table: t,
	})
}

// AddImage adds an image block to the document.
func (d *Document) AddImage(img *ImageBlock) {
	d.Content = append(d.Content, Block{
		Type:  BlockTypeImage,
		Image: img,
	})
}

// AddList adds a list block to the document.
func (d *Document) AddList(l *ListBlock) {
	d.Content = append(d.Content, Block{
		Type: BlockTypeList,
		List: l,
	})
}
