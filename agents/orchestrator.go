package agents

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"celeste/models"
	"google.golang.org/genai"
)

type AgentOrchestrator struct {
	geminiClient *genai.Client
	agents       map[string]models.Agent
	messageBus   chan models.AgentMessage
	contextStore map[string]*models.UserContext
	mutex        sync.RWMutex
}

func NewAgentOrchestrator(geminiClient *genai.Client) *AgentOrchestrator {
	return &AgentOrchestrator{
		geminiClient: geminiClient,
		agents:       make(map[string]models.Agent),
		messageBus:   make(chan models.AgentMessage, 100),
		contextStore: make(map[string]*models.UserContext),
	}
}

func (ao *AgentOrchestrator) RegisterAgent(agent models.Agent) error {
	ao.mutex.Lock()
	defer ao.mutex.Unlock()

	if err := agent.Initialize(context.Background()); err != nil {
		return fmt.Errorf("failed to initialize agent %s: %v", agent.ID(), err)
	}

	ao.agents[agent.ID()] = agent
	log.Printf("Agent registered and initialized: %s", agent.ID())
	return nil
}

func (ao *AgentOrchestrator) Initialize() error {
	inventoryAgent := NewInventoryAgent(ao.geminiClient)
	searchAgent := NewSearchAgent(ao.geminiClient)
	recommendationAgent := NewRecommendationAgent(ao.geminiClient)

	agents := []models.Agent{inventoryAgent, searchAgent, recommendationAgent}

	for _, agent := range agents {
		if err := ao.RegisterAgent(agent); err != nil {
			return err
		}
	}

	go ao.processMessages()

	log.Printf("Agent orchestrator initialized with %d agents", len(ao.agents))
	return nil
}

func (ao *AgentOrchestrator) UpdateUserContext(userID string, context *models.UserContext) {
	ao.mutex.Lock()
	defer ao.mutex.Unlock()
	ao.contextStore[userID] = context
}

func (ao *AgentOrchestrator) GetUserContext(userID string) *models.UserContext {
	ao.mutex.RLock()
	defer ao.mutex.RUnlock()
	return ao.contextStore[userID]
}

func (ao *AgentOrchestrator) sendToAgent(message models.AgentMessage) {
	select {
	case ao.messageBus <- message:
	case <-time.After(5 * time.Second):
		log.Printf("Message timeout for agent %s", message.ToAgent)
	}
}

func (ao *AgentOrchestrator) processMessages() {
	for message := range ao.messageBus {
		if agent, exists := ao.agents[message.ToAgent]; exists {
			go func(a models.Agent, msg models.AgentMessage) {
				_, err := a.Process(context.Background(), msg)
				if err != nil {
					log.Printf("Agent %s processing error: %v", a.ID(), err)
				}
			}(agent, message)
		}
	}
}

func (ao *AgentOrchestrator) ProcessUserRequest(ctx context.Context, userID, query string) (*models.CelesteResponse, error) {
	userContext := ao.GetUserContext(userID)
	if userContext == nil {
		userContext = &models.UserContext{
			UserID:      userID,
			Preferences: make(map[string]string),
			History:     []string{},
			CartItems:   []string{},
		}
		ao.UpdateUserContext(userID, userContext)
	}

	userContext.History = append(userContext.History, query)
	if len(userContext.History) > 10 {
		userContext.History = userContext.History[1:]
	}
	ao.UpdateUserContext(userID, userContext)

	workflowID := fmt.Sprintf("workflow_%s_%d", userID, time.Now().Unix())
	agentPath := []string{}

	searchMsg := models.AgentMessage{
		ID:        fmt.Sprintf("%s_search", workflowID),
		FromAgent: "orchestrator",
		ToAgent:   "search_agent",
		Type:      "product_search",
		Data:      map[string]interface{}{"query": query},
		Context:   userContext,
		Timestamp: time.Now(),
	}

	searchAgent := ao.agents["search_agent"]
	searchResponse, err := searchAgent.Process(ctx, searchMsg)
	if err != nil {
		return nil, err
	}
	agentPath = append(agentPath, "search_agent")

	inventoryMsg := models.AgentMessage{
		ID:        fmt.Sprintf("%s_inventory", workflowID),
		FromAgent: "orchestrator",
		ToAgent:   "inventory_agent",
		Type:      "check_inventory",
		Data:      searchResponse.Data,
		Context:   userContext,
		Timestamp: time.Now(),
	}

	inventoryAgent := ao.agents["inventory_agent"]
	inventoryResponse, err := inventoryAgent.Process(ctx, inventoryMsg)
	if err != nil {
		log.Printf("Inventory check failed: %v", err)
		inventoryResponse = &models.AgentResponse{Data: make(map[string]interface{})}
	} else {
		agentPath = append(agentPath, "inventory_agent")
	}

	recMsg := models.AgentMessage{
		ID:        fmt.Sprintf("%s_recommendations", workflowID),
		FromAgent: "orchestrator",
		ToAgent:   "recommendation_agent",
		Type:      "personalized_recommendations",
		Data: map[string]interface{}{
			"search_results": searchResponse.Data,
			"inventory_info": inventoryResponse.Data,
		},
		Context:   userContext,
		Timestamp: time.Now(),
	}

	recAgent := ao.agents["recommendation_agent"]
	recResponse, err := recAgent.Process(ctx, recMsg)
	if err != nil {
		log.Printf("Recommendation generation failed: %v", err)
		recResponse = &models.AgentResponse{Data: make(map[string]interface{})}
	} else {
		agentPath = append(agentPath, "recommendation_agent")
	}

	return ao.synthesizeResponse(query, searchResponse, inventoryResponse, recResponse, workflowID, agentPath)
}

func (ao *AgentOrchestrator) synthesizeResponse(query string, searchResp, invResp, recResp *models.AgentResponse, workflowID string, agentPath []string) (*models.CelesteResponse, error) {
	var products []models.Product
	if searchData, ok := searchResp.Data["products"].([]models.Product); ok {
		products = searchData
	}

	actions := []string{"Browse similar items", "Add to wishlist", "Get size guidance"}
	if recActions, ok := recResp.Data["actions"].([]string); ok {
		actions = recActions
	}

	message := ao.generateAgentCoordinatedResponse(query, searchResp, invResp, recResp)

	return &models.CelesteResponse{
		Message:      message,
		Products:     products,
		Actions:      actions,
		WorkflowID:   workflowID,
		AgentPath:    agentPath,
		Personalized: len(agentPath) > 1,
	}, nil
}

func (ao *AgentOrchestrator) generateAgentCoordinatedResponse(query string, searchResp, invResp, recResp *models.AgentResponse) string {
	prompt := fmt.Sprintf(`You are Céleste, coordinating multiple AI agents to provide intelligent shopping assistance.

Customer query: "%s"

Agent coordination results:
- Search Agent: Found products and analyzed intent
- Inventory Agent: Checked stock levels and availability  
- Recommendation Agent: Generated personalized suggestions

Provide a response that demonstrates this multi-agent coordination while being helpful and conversational.`, query)

	resp, err := ao.geminiClient.Models.GenerateContent(
		context.Background(),
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		return "I've coordinated multiple agents to find the best options for you!"
	}

	return resp.Text()
}

func (ao *AgentOrchestrator) ListAgents() []string {
	ao.mutex.RLock()
	defer ao.mutex.RUnlock()

	var agents []string
	for id := range ao.agents {
		agents = append(agents, id)
	}
	return agents
}
