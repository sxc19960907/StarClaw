package daemon

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// MarketplaceItem represents a skill available in the marketplace.
type MarketplaceItem struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Author      string `json:"author,omitempty"`
	Version     string `json:"version,omitempty"`
	URL         string `json:"url,omitempty"`
}

// ListMarketplace reads the marketplace index from
// <starclawDir>/marketplace.json and returns the list of available
// items.  Returns an empty slice if the file does not exist or cannot
// be parsed.
func ListMarketplace(starclawDir string) []MarketplaceItem {
	if starclawDir == "" {
		return nil
	}

	path := filepath.Join(starclawDir, "marketplace.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	// Support both a single array and an object with an "items" key.
	var items []MarketplaceItem
	if err := json.Unmarshal(data, &items); err == nil {
		return items
	}

	var wrapper struct {
		Items []MarketplaceItem `json:"items"`
	}
	if err := json.Unmarshal(data, &wrapper); err == nil {
		return wrapper.Items
	}

	return nil
}
