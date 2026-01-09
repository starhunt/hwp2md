package ir

// Paragraph represents a text paragraph with style information.
type Paragraph struct {
	Text  string         `json:"text"`
	Runs  []Run          `json:"runs,omitempty"`
	Style ParagraphStyle `json:"style"`
}

// Run represents a styled text run within a paragraph.
type Run struct {
	Text  string    `json:"text"`
	Style TextStyle `json:"style,omitempty"`
}

// ParagraphStyle contains paragraph-level styling hints.
type ParagraphStyle struct {
	HeadingLevel int    `json:"heading_level,omitempty"` // 0 = normal, 1-6 = heading
	Alignment    string `json:"alignment,omitempty"`     // left, center, right, justify
	Indent       int    `json:"indent,omitempty"`        // indent level
	IsQuote      bool   `json:"is_quote,omitempty"`      // blockquote hint
}

// TextStyle contains character-level styling hints.
type TextStyle struct {
	Bold          bool   `json:"bold,omitempty"`
	Italic        bool   `json:"italic,omitempty"`
	Underline     bool   `json:"underline,omitempty"`
	Strikethrough bool   `json:"strikethrough,omitempty"`
	Superscript   bool   `json:"superscript,omitempty"`
	Subscript     bool   `json:"subscript,omitempty"`
	Code          bool   `json:"code,omitempty"`
	Link          string `json:"link,omitempty"` // hyperlink URL
}

// NewParagraph creates a new paragraph with the given text.
func NewParagraph(text string) *Paragraph {
	return &Paragraph{
		Text: text,
		Runs: make([]Run, 0),
	}
}

// AddRun adds a styled text run to the paragraph.
func (p *Paragraph) AddRun(text string, style TextStyle) {
	p.Runs = append(p.Runs, Run{
		Text:  text,
		Style: style,
	})
}

// SetHeading sets the heading level for the paragraph.
func (p *Paragraph) SetHeading(level int) {
	if level < 0 {
		level = 0
	}
	if level > 6 {
		level = 6
	}
	p.Style.HeadingLevel = level
}

// IsEmpty returns true if the paragraph has no text content.
func (p *Paragraph) IsEmpty() bool {
	return p.Text == "" && len(p.Runs) == 0
}
