package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/genai"

	"celeste/agents"
)

type ChatRequest struct {
	Query  string `json:"query"`
	UserID string `json:"user_id,omitempty"`
}

type CelesteService struct {
	orchestrator *agents.AgentOrchestrator
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		fmt.Printf("Open this URL in your browser: %s\n", url)
		return
	}
	if err != nil {
		fmt.Printf("Could not open browser: %v\nOpen this URL manually: %s\n", err, url)
	}
}

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal("Failed to create Gemini client:", err)
	}

	orchestrator := agents.NewAgentOrchestrator(client)
	if err := orchestrator.Initialize(); err != nil {
		log.Fatal("Failed to initialize agent orchestrator:", err)
	}

	service := &CelesteService{
		orchestrator: orchestrator,
	}

	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Celeste Multi-Agent AI Shopping Assistant")
	}).Methods("GET")

	router.HandleFunc("/chat", service.handleChat).Methods("POST")

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		agentList := service.orchestrator.ListAgents()
		response := map[string]interface{}{
			"status":      "healthy",
			"agents":      agentList,
			"agent_count": len(agentList),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}).Methods("GET")

	router.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("./home.html"); err == nil {
			http.ServeFile(w, r, "./home.html")
		} else {
			http.ServeFile(w, r, "./web/home.html")
		}
	}).Methods("GET")

	router.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("./home.html"); err == nil {
			http.ServeFile(w, r, "./home.html")
		} else {
			http.ServeFile(w, r, "./web/home.html")
		}
	}).Methods("GET")

	router.HandleFunc("/api-comparison", func(w http.ResponseWriter, r *http.Request) {
		if _, err := os.Stat("./api-comparison.html"); err == nil {
			http.ServeFile(w, r, "./api-comparison.html")
		} else {
			http.ServeFile(w, r, "./web/api-comparison.html")
		}
	}).Methods("GET")

	router.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		agentList := service.orchestrator.ListAgents()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"agents": agentList,
			"count":  len(agentList),
		})
	}).Methods("GET")
	port := ":8080"
	fmt.Printf("Celeste Multi-Agent System starting on port %s\n", port)

	go func() {
		time.Sleep(2 * time.Second)
		openBrowser("http://localhost:8080/home")
	}()

	log.Fatal(http.ListenAndServe(port, router))
}

func (s *CelesteService) handleChat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	userID := req.UserID
	if userID == "" {
		userID = "anonymous_user"
	}

	ctx := r.Context()
	response, err := s.orchestrator.ProcessUserRequest(ctx, userID, req.Query)
	if err != nil {
		log.Printf("Orchestrator error: %v", err)
		http.Error(w, "Agent processing failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
