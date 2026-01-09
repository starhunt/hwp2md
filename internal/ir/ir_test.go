package ir

import (
	"encoding/json"
	"testing"
)

func TestNewDocument(t *testing.T) {
	doc := NewDocument()

	if doc.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", doc.Version)
	}
	if len(doc.Content) != 0 {
		t.Errorf("expected empty content, got %d blocks", len(doc.Content))
	}
}

func TestDocument_AddParagraph(t *testing.T) {
	doc := NewDocument()
	p := NewParagraph("Hello, World!")

	doc.AddParagraph(p)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 block, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != BlockTypeParagraph {
		t.Errorf("expected paragraph type, got %s", doc.Content[0].Type)
	}
	if doc.Content[0].Paragraph.Text != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %s", doc.Content[0].Paragraph.Text)
	}
}

func TestDocument_AddTable(t *testing.T) {
	doc := NewDocument()
	table := NewTable(2, 3)
	table.SetCell(0, 0, "Header 1")
	table.SetCell(0, 1, "Header 2")
	table.SetCell(0, 2, "Header 3")

	doc.AddTable(table)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 block, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != BlockTypeTable {
		t.Errorf("expected table type, got %s", doc.Content[0].Type)
	}
	if doc.Content[0].Table.Rows != 2 {
		t.Errorf("expected 2 rows, got %d", doc.Content[0].Table.Rows)
	}
}

func TestDocument_AddImage(t *testing.T) {
	doc := NewDocument()
	img := NewImage("img001")
	img.Alt = "Test image"

	doc.AddImage(img)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 block, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != BlockTypeImage {
		t.Errorf("expected image type, got %s", doc.Content[0].Type)
	}
	if doc.Content[0].Image.ID != "img001" {
		t.Errorf("expected 'img001', got %s", doc.Content[0].Image.ID)
	}
}

func TestDocument_AddList(t *testing.T) {
	doc := NewDocument()
	list := NewUnorderedList()
	list.AddItem("Item 1")
	list.AddItem("Item 2")

	doc.AddList(list)

	if len(doc.Content) != 1 {
		t.Fatalf("expected 1 block, got %d", len(doc.Content))
	}
	if doc.Content[0].Type != BlockTypeList {
		t.Errorf("expected list type, got %s", doc.Content[0].Type)
	}
	if len(doc.Content[0].List.Items) != 2 {
		t.Errorf("expected 2 items, got %d", len(doc.Content[0].List.Items))
	}
}

func TestDocument_JSONSerialization(t *testing.T) {
	doc := NewDocument()
	doc.Metadata.Title = "Test Document"
	doc.Metadata.Author = "Test Author"

	p := NewParagraph("Test paragraph")
	p.SetHeading(1)
	doc.AddParagraph(p)

	// Serialize
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Deserialize
	var restored Document
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// Verify
	if restored.Version != doc.Version {
		t.Errorf("version mismatch: expected %s, got %s", doc.Version, restored.Version)
	}
	if restored.Metadata.Title != doc.Metadata.Title {
		t.Errorf("title mismatch: expected %s, got %s", doc.Metadata.Title, restored.Metadata.Title)
	}
	if len(restored.Content) != len(doc.Content) {
		t.Errorf("content length mismatch: expected %d, got %d", len(doc.Content), len(restored.Content))
	}
}
