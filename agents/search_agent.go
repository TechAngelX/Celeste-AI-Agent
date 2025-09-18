package agents

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"celeste/models"
	"google.golang.org/genai"
)

type SearchAgent struct {
	id           string
	geminiClient *genai.Client
	catalog      []models.Product
}

func NewSearchAgent(geminiClient *genai.Client) *SearchAgent {
	return &SearchAgent{
		id:           "search_agent",
		geminiClient: geminiClient,
	}
}

func (sa *SearchAgent) ID() string {
	return sa.id
}

func (sa *SearchAgent) Initialize(ctx context.Context) error {
	data, err := os.ReadFile("data/itemCatalogue.json")
	if err != nil {
		return err
	}

	var catalog struct {
		Products []models.Product `json:"products"`
	}

	if err := json.Unmarshal(data, &catalog); err != nil {
		return err
	}

	sa.catalog = catalog.Products
	return nil
}

func (sa *SearchAgent) Process(ctx context.Context, input models.AgentMessage) (*models.AgentResponse, error) {
	query, ok := input.Data["query"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid query format")
	}

	intent, err := sa.analyzeIntent(ctx, query)
	if err != nil {
		intent = "general_search"
	}

	products := sa.searchProducts(query)

	return &models.AgentResponse{
		ID:        input.ID,
		FromAgent: sa.id,
		Type:      "search_results",
		Data: map[string]interface{}{
			"products": products,
			"intent":   intent,
			"query":    query,
		},
		NextActions: []string{"check_inventory", "get_recommendations"},
		Success:     true,
	}, nil
}

func (sa *SearchAgent) analyzeIntent(ctx context.Context, query string) (string, error) {
	prompt := fmt.Sprintf(`Analyze this query and return one classification:
Query: "%s"

Classifications: product_search, style_advice, price_inquiry, size_help, occasion_shopping, comparison, general_help

Return only the classification.`, query)

	resp, err := sa.geminiClient.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), nil)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Text()), nil
}

func (sa *SearchAgent) searchProducts(query string) []models.Product {
	queryLower := strings.ToLower(query)
	var matches []models.Product

	synonyms := map[string][]string{
		"shoes": {"boots", "trainers", "footwear"},
		"top":   {"shirt", "tank", "clothing"},
		"bag":   {"tote", "handbag", "accessory"},
		"dress": {"gown", "outfit", "clothing"},
	}

	for _, product := range sa.catalog {
		score := 0

		if strings.Contains(strings.ToLower(product.Name), queryLower) {
			score += 5
		}
		if strings.Contains(strings.ToLower(product.Description), queryLower) {
			score += 3
		}

		for _, category := range product.Categories {
			if strings.Contains(queryLower, category) {
				score += 2
			}
		}

		for term, syns := range synonyms {
			if strings.Contains(queryLower, term) {
				for _, syn := range syns {
					if strings.Contains(strings.ToLower(product.Name), syn) {
						score += 2
					}
				}
			}
		}

		if score > 0 {
			matches = append(matches, product)
		}

		if len(matches) >= 4 {
			break
		}
	}

	return matches
}

func (sa *SearchAgent) Shutdown(ctx context.Context) error {
	return nil
}
