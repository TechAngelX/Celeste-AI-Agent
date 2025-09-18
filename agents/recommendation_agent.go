package agents

import (
	"context"
	"fmt"
	"strings"

	"celeste/models"
	"google.golang.org/genai"
)

type RecommendationAgent struct {
	id           string
	geminiClient *genai.Client
}

func NewRecommendationAgent(geminiClient *genai.Client) *RecommendationAgent {
	return &RecommendationAgent{
		id:           "recommendation_agent",
		geminiClient: geminiClient,
	}
}

func (ra *RecommendationAgent) ID() string {
	return ra.id
}

func (ra *RecommendationAgent) Initialize(ctx context.Context) error {
	return nil
}

func (ra *RecommendationAgent) Process(ctx context.Context, input models.AgentMessage) (*models.AgentResponse, error) {
	searchResults, ok := input.Data["search_results"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no search results for recommendations")
	}

	inventoryInfo, _ := input.Data["inventory_info"].(map[string]interface{})
	userContext := input.Context

	recommendations := ra.generatePersonalizedRecommendations(ctx, searchResults, inventoryInfo, userContext)
	actions := ra.generateNextActions(searchResults, userContext)

	return &models.AgentResponse{
		ID:        input.ID,
		FromAgent: ra.id,
		Type:      "personalized_recommendations",
		Data: map[string]interface{}{
			"recommendations":       recommendations,
			"actions":               actions,
			"personalization_score": ra.calculatePersonalizationScore(userContext),
		},
		NextActions: actions,
		Success:     true,
	}, nil
}

func (ra *RecommendationAgent) generatePersonalizedRecommendations(ctx context.Context, searchResults, inventoryInfo map[string]interface{}, userContext *models.UserContext) []string {
	intent, _ := searchResults["intent"].(string)

	baseRecommendations := ra.getBaseRecommendations(intent)

	if userContext != nil && len(userContext.History) > 0 {
		return ra.personalizeRecommendations(baseRecommendations, userContext)
	}

	return baseRecommendations
}

func (ra *RecommendationAgent) getBaseRecommendations(intent string) []string {
	recommendations := map[string][]string{
		"product_search":    {"View similar items", "Add to cart", "Compare prices"},
		"style_advice":      {"Get style guide", "See outfit suggestions", "Book style consultation"},
		"occasion_shopping": {"Complete the look", "See accessories", "Size guidance"},
		"price_inquiry":     {"Price alerts", "Compare alternatives", "See deals"},
		"size_help":         {"Size chart", "Virtual fitting", "Exchange policy"},
	}

	if recs, exists := recommendations[intent]; exists {
		return recs
	}

	return []string{"Browse categories", "Get recommendations", "Contact support"}
}

func (ra *RecommendationAgent) personalizeRecommendations(baseRecs []string, userContext *models.UserContext) []string {
	personalized := make([]string, 0, len(baseRecs))

	for _, rec := range baseRecs {
		if ra.isRelevantToUser(rec, userContext) {
			personalized = append(personalized, rec)
		}
	}

	if len(personalized) == 0 {
		return baseRecs
	}

	return personalized
}

func (ra *RecommendationAgent) isRelevantToUser(recommendation string, userContext *models.UserContext) bool {
	if len(userContext.History) == 0 {
		return true
	}

	recentQueries := strings.Join(userContext.History, " ")

	relevanceMap := map[string][]string{
		"style":    {"style", "outfit", "look", "fashion"},
		"price":    {"price", "cost", "cheap", "expensive", "deal"},
		"size":     {"size", "fit", "large", "small", "medium"},
		"occasion": {"wedding", "party", "work", "casual", "formal"},
	}

	for category, keywords := range relevanceMap {
		if strings.Contains(recommendation, category) {
			for _, keyword := range keywords {
				if strings.Contains(strings.ToLower(recentQueries), keyword) {
					return true
				}
			}
		}
	}

	return true
}

func (ra *RecommendationAgent) generateNextActions(searchResults map[string]interface{}, userContext *models.UserContext) []string {
	actions := []string{}

	products, ok := searchResults["products"].([]models.Product)
	if ok && len(products) > 0 {
		actions = append(actions, "Add to cart", "Save to wishlist")
	}

	if userContext != nil && len(userContext.History) > 2 {
		actions = append(actions, "View browsing history", "Get personalized suggestions")
	}

	actions = append(actions, "Continue shopping", "Get styling advice")

	return actions
}

func (ra *RecommendationAgent) calculatePersonalizationScore(userContext *models.UserContext) float64 {
	if userContext == nil {
		return 0.0
	}

	score := 0.0

	if len(userContext.History) > 0 {
		score += 0.3
	}
	if len(userContext.Preferences) > 0 {
		score += 0.4
	}
	if len(userContext.CartItems) > 0 {
		score += 0.3
	}

	return score
}

func (ra *RecommendationAgent) Shutdown(ctx context.Context) error {
	return nil
}
