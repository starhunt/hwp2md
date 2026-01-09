package upstage

import (
	"os"
	"testing"
)

func TestNew_NoAPIKey(t *testing.T) {
	// Temporarily unset API key
	oldKey := os.Getenv("UPSTAGE_API_KEY")
	os.Unsetenv("UPSTAGE_API_KEY")
	defer func() {
		if oldKey != "" {
			os.Setenv("UPSTAGE_API_KEY", oldKey)
		}
	}()

	_, err := New(Config{})
	if err == nil {
		t.Error("expected error when API key is not set")
	}
}

func TestNew_WithAPIKey(t *testing.T) {
	p, err := New(Config{
		APIKey: "test-key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.Name() != ProviderName {
		t.Errorf("expected provider name %q, got %q", ProviderName, p.Name())
	}
}

func TestNew_DefaultValues(t *testing.T) {
	p, err := New(Config{
		APIKey: "test-key",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.baseURL != DefaultBaseURL {
		t.Errorf("expected baseURL %q, got %q", DefaultBaseURL, p.baseURL)
	}

	if p.model != DefaultModel {
		t.Errorf("expected model %q, got %q", DefaultModel, p.model)
	}
}

func TestNew_CustomValues(t *testing.T) {
	customURL := "https://custom.api.upstage.ai/v1/parse"
	customModel := "document-parse-nightly"

	p, err := New(Config{
		APIKey:  "test-key",
		BaseURL: customURL,
		Model:   customModel,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.baseURL != customURL {
		t.Errorf("expected baseURL %q, got %q", customURL, p.baseURL)
	}

	if p.model != customModel {
		t.Errorf("expected model %q, got %q", customModel, p.model)
	}
}

func TestParseList_Ordered(t *testing.T) {
	p, _ := New(Config{APIKey: "test-key"})

	markdown := "1. First item\n2. Second item\n3. Third item"
	list := p.parseList(markdown, "")

	if list == nil {
		t.Fatal("expected list, got nil")
	}

	if !list.Ordered {
		t.Error("expected ordered list")
	}

	if len(list.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(list.Items))
	}
}

func TestParseList_Unordered(t *testing.T) {
	p, _ := New(Config{APIKey: "test-key"})

	markdown := "- First item\n- Second item\n- Third item"
	list := p.parseList(markdown, "")

	if list == nil {
		t.Fatal("expected list, got nil")
	}

	if list.Ordered {
		t.Error("expected unordered list")
	}

	if len(list.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(list.Items))
	}
}

func TestParseList_Empty(t *testing.T) {
	p, _ := New(Config{APIKey: "test-key"})

	list := p.parseList("", "")
	if list != nil {
		t.Error("expected nil for empty input")
	}
}

func TestParseHTMLTable_Empty(t *testing.T) {
	p, _ := New(Config{APIKey: "test-key"})

	table := p.parseHTMLTable("")
	if table != nil {
		t.Error("expected nil for empty HTML")
	}
}

func TestProvider_Name(t *testing.T) {
	p, _ := New(Config{APIKey: "test-key"})

	if p.Name() != "upstage" {
		t.Errorf("expected name 'upstage', got %q", p.Name())
	}
}
