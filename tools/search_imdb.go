package tools

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/gocolly/colly/v2"
)

type SearchIMDbInput struct {
	SearchTerm string `json:"search_term" jsonschema_description:"The search term to look for on IMDb."`
}

var SearchIMDbInputSchema = GenerateSchema[SearchIMDbInput]()

var SearchIMDbDefinition = ToolDefinition{
	Name:        "search_imdb",
	Description: "Search for a title on IMDb. Returns a JSON string with title, id, and description.",
	InputSchema: SearchIMDbInputSchema,
	Function:    SearchIMDb,
}

func SearchIMDb(input json.RawMessage) (string, error) {
	searchInput := SearchIMDbInput{}
	err := json.Unmarshal(input, &searchInput)
	if err != nil {
		return "", err
	}

	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"),
	)

	type IMDbResult struct {
		RawResult string `json:"rawResult"`
		ID        string `json:"id"`
	}

	var results []IMDbResult

	// Scrape each search result
	c.OnHTML(".ipc-metadata-list-summary-item__tc", func(e *colly.HTMLElement) {
		textContent := e.Text
		
		// Get the ID from the first child anchor tag
		href := e.ChildAttr("a", "href")
		
		// Extract ID from href like "/title/tt4955642/?ref_=fn_all_ttl_1"
		var id string
		if href != "" {
			parts := strings.Split(href, "/")
			if len(parts) > 2 {
				// Get the third part and remove query parameters
				idPart := strings.Split(parts[2], "?")[0]
				id = idPart
			}
		}

		results = append(results, IMDbResult{
			RawResult: textContent,
			ID:        id,
		})
	})

	// Construct IMDB search URL
	searchURL := fmt.Sprintf("https://www.imdb.com/find/?q=%s&ref_=nv_sr_sm", url.QueryEscape(searchInput.SearchTerm))
	
	err = c.Visit(searchURL)
	if err != nil {
		return "", fmt.Errorf("failed to scrape IMDB: %w", err)
	}

	// Convert results to JSON
	jsonData, err := json.Marshal(results)
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	return string(jsonData), nil
}
