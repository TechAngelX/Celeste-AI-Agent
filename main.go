package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"google.golang.org/genai"
)

type ChatRequest struct {
	Query string `json:"query"`
}

type ChatResponse struct {
	Message string `json:"message"`
}

func main() {
	// Check for API key secret
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	// Initialise Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal("Failed to create Gemini client:", err)
	}

	// Set up routes
	router := mux.NewRouter()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello! I'm Celeste, your AI shopping assistant!")
	}).Methods("GET")

	router.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		handleChat(w, r, client)
	}).Methods("POST")

	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Celeste is healthy!")
	}).Methods("GET")
	router.HandleFunc("/home", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./home.html")
	}).Methods("GET")

	port := ":8080"
	fmt.Printf("Celeste is starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, router))
}
func handleChat(w http.ResponseWriter, r *http.Request, client *genai.Client) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	prompt := fmt.Sprintf(`You are Celeste, a helpful and friendly shopping assistant for an online boutique.
    The customer said: "%s"

    Respond as Celeste in a helpful, conversational way. Keep it brief but friendly.`, req.Query)

	resp, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		log.Printf("Gemini error: %v", err)
		http.Error(w, "Sorry, I'm having trouble right now", http.StatusInternalServerError)
		return
	}

	response := ChatResponse{
		Message: resp.Text(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
