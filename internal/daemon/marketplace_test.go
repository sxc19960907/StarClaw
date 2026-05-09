package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestListMarketplace_NoFile(t *testing.T) {
	dir := t.TempDir()
	items := ListMarketplace(dir)
	if items != nil {
		t.Errorf("expected nil for nonexistent file, got %v", items)
	}
}

func TestListMarketplace_EmptyDir(t *testing.T) {
	items := ListMarketplace("")
	if items != nil {
		t.Errorf("expected nil for empty dir, got %v", items)
	}
}

func TestListMarketplace_ArrayFormat(t *testing.T) {
	dir := t.TempDir()

	items := []MarketplaceItem{
		{Name: "skill-a", Description: "Skill A description", Author: "author1", Version: "1.0.0"},
		{Name: "skill-b", Description: "Skill B description", URL: "https://example.com/b"},
	}

	data, err := json.Marshal(items)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "marketplace.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	result := ListMarketplace(dir)
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}

	if result[0].Name != "skill-a" {
		t.Errorf("expected first item name 'skill-a', got %q", result[0].Name)
	}
	if result[0].Description != "Skill A description" {
		t.Errorf("expected description 'Skill A description', got %q", result[0].Description)
	}
	if result[0].Author != "author1" {
		t.Errorf("expected author 'author1', got %q", result[0].Author)
	}
	if result[0].Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got %q", result[0].Version)
	}
	if result[1].Name != "skill-b" {
		t.Errorf("expected second item name 'skill-b', got %q", result[1].Name)
	}
	if result[1].URL != "https://example.com/b" {
		t.Errorf("expected URL 'https://example.com/b', got %q", result[1].URL)
	}
}

func TestListMarketplace_ObjectWithItemsFormat(t *testing.T) {
	dir := t.TempDir()

	wrapper := struct {
		Items []MarketplaceItem `json:"items"`
	}{
		Items: []MarketplaceItem{
			{Name: "alpha", Description: "Alpha skill"},
			{Name: "beta", Description: "Beta skill"},
		},
	}

	data, err := json.Marshal(wrapper)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "marketplace.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	result := ListMarketplace(dir)
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
	if result[1].Name != "beta" {
		t.Errorf("expected second item name 'beta', got %q", result[1].Name)
	}
}

func TestListMarketplace_InvalidJSON(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "marketplace.json"), []byte("{invalid json}"), 0644); err != nil {
		t.Fatal(err)
	}

	result := ListMarketplace(dir)
	if result != nil {
		t.Errorf("expected nil for invalid JSON, got %v", result)
	}
}

func TestListMarketplace_EmptyItems(t *testing.T) {
	dir := t.TempDir()

	data, err := json.Marshal([]MarketplaceItem{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "marketplace.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	result := ListMarketplace(dir)
	if result == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(result) != 0 {
		t.Errorf("expected 0 items, got %d", len(result))
	}
}

func TestListMarketplace_FullItem(t *testing.T) {
	dir := t.TempDir()

	fullItem := MarketplaceItem{
		Name:        "complete-skill",
		Description: "A fully described skill",
		Author:      "starclaw",
		Version:     "2.1.3",
		URL:         "https://starclaw.dev/skills/complete-skill",
	}

	data, err := json.Marshal([]MarketplaceItem{fullItem})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "marketplace.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	result := ListMarketplace(dir)
	if len(result) != 1 {
		t.Fatalf("expected 1 item, got %d", len(result))
	}
	if result[0] != fullItem {
		t.Errorf("item mismatch: got %+v, want %+v", result[0], fullItem)
	}
}
