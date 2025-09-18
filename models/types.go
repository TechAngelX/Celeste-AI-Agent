// models/types.go
package models

import (
	"context"
	"time"
)

// Core data types
type Product struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Picture     string   `json:"picture"`
	PriceUsd    PriceUsd `json:"priceUsd"`
	Categories  []string `json:"categories"`
}

type PriceUsd struct {
	CurrencyCode string `json:"currency_code"`
	Units        int64  `json:"units"`
	Nanos        int32  `json:"nanos"`
}

// Agent framework types (ADK-inspired)
type Agent interface {
	ID() string
	Initialize(ctx context.Context) error
	Process(ctx context.Context, input AgentMessage) (*AgentResponse, error)
	Shutdown(ctx context.Context) error
}

// MCP-inspired message format
type AgentMessage struct {
	ID        string                 `json:"id"`
	FromAgent string                 `json:"from_agent"`
	ToAgent   string                 `json:"to_agent"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Context   *UserContext           `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type AgentResponse struct {
	ID          string                 `json:"id"`
	FromAgent   string                 `json:"from_agent"`
	Type        string                 `json:"type"`
	Data        map[string]interface{} `json:"data"`
	NextActions []string               `json:"next_actions,omitempty"`
	Success     bool                   `json:"success"`
	Error       string                 `json:"error,omitempty"`
}

// User context for personalization
type UserContext struct {
	UserID      string            `json:"user_id"`
	Preferences map[string]string `json:"preferences"`
	History     []string          `json:"history"`
	CartItems   []string          `json:"cart_items"`
	Location    string            `json:"location,omitempty"`
}

// Enhanced chat response
type CelesteResponse struct {
	Message      string    `json:"message"`
	Products     []Product `json:"products,omitempty"`
	Actions      []string  `json:"actions,omitempty"`
	WorkflowID   string    `json:"workflow_id"`
	AgentPath    []string  `json:"agent_path"` // Shows which agents were involved
	Personalized bool      `json:"personalized"`
}
