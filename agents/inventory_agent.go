package agents

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"celeste/models"
	"google.golang.org/genai"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type InventoryAgent struct {
	id                 string
	geminiClient       *genai.Client
	catalogServiceAddr string
}

func NewInventoryAgent(geminiClient *genai.Client) *InventoryAgent {
	return &InventoryAgent{
		id:                 "inventory_agent",
		geminiClient:       geminiClient,
		catalogServiceAddr: "productcatalogservice:3550",
	}
}

func (ia *InventoryAgent) ID() string {
	return ia.id
}

func (ia *InventoryAgent) Initialize(ctx context.Context) error {
	return nil
}

func (ia *InventoryAgent) Process(ctx context.Context, input models.AgentMessage) (*models.AgentResponse, error) {
	products, ok := input.Data["products"].([]models.Product)
	if !ok {
		return nil, fmt.Errorf("no products to check inventory")
	}

	inventoryStatus := ia.checkInventoryStatus(ctx, products)
	recommendations := ia.generateInventoryRecommendations(inventoryStatus)

	return &models.AgentResponse{
		ID:        input.ID,
		FromAgent: ia.id,
		Type:      "inventory_status",
		Data: map[string]interface{}{
			"inventory_status": inventoryStatus,
			"recommendations":  recommendations,
			"checked_at":       time.Now().Format(time.RFC3339),
		},
		NextActions: []string{"generate_recommendations", "update_user_context"},
		Success:     true,
	}, nil
}

func (ia *InventoryAgent) checkInventoryStatus(ctx context.Context, products []models.Product) map[string]interface{} {
	conn, err := grpc.Dial(ia.catalogServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return ia.simulateInventoryCheck(products)
	}
	defer conn.Close()

	return ia.simulateInventoryCheck(products)
}

func (ia *InventoryAgent) simulateInventoryCheck(products []models.Product) map[string]interface{} {
	status := make(map[string]interface{})
	rand.Seed(time.Now().UnixNano())

	for _, product := range products {
		stockLevel := rand.Intn(50) + 1
		demand := []string{"low", "medium", "high"}[rand.Intn(3)]

		status[product.ID] = map[string]interface{}{
			"stock_level": stockLevel,
			"demand":      demand,
			"trending":    stockLevel < 10,
			"seasonal":    ia.isSeasonalItem(product),
		}
	}

	return status
}

func (ia *InventoryAgent) isSeasonalItem(product models.Product) bool {
	seasonal_keywords := []string{"summer", "winter", "spring", "fall", "holiday"}
	for _, keyword := range seasonal_keywords {
		if contains(product.Categories, keyword) ||
			contains([]string{product.Name, product.Description}, keyword) {
			return true
		}
	}
	return false
}

func (ia *InventoryAgent) generateInventoryRecommendations(status map[string]interface{}) []string {
	recommendations := []string{}

	for productID, info := range status {
		if infoMap, ok := info.(map[string]interface{}); ok {
			if stockLevel, ok := infoMap["stock_level"].(int); ok && stockLevel < 10 {
				recommendations = append(recommendations, fmt.Sprintf("Low stock alert for %s", productID))
			}
			if trending, ok := infoMap["trending"].(bool); ok && trending {
				recommendations = append(recommendations, fmt.Sprintf("Trending item: %s", productID))
			}
		}
	}

	return recommendations
}

func (ia *InventoryAgent) Shutdown(ctx context.Context) error {
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
