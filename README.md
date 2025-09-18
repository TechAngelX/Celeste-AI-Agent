cat > README.md << 'EOF'
# Introducing Céleste. The AI Shopping Assistant

An intelligent shopping assistant AI chatbot built for the **GKE Turns 10 Hackathon**. Céleste is an enhanced agentic AI-powered microservice that enhances the Google Online Boutique demo application with conversational shopping assistance using Google Gemini AI.

## Hackathon Challenge

This project was developed for the GKE Turns 10 Hackathon (Sept 2025), which challenges participants to integrate cutting-edge agentic AI capabilities with existing microservice applications orchestrated on Google Kubernetes Engine (GKE).

**Challenge Requirements:**
- Enhance Online Boutique microservice application with agentic AI
- Deploy and orchestrate on Google Kubernetes Engine (GKE)
- Integrate Google AI models (Gemini)
- Build new components that interact with existing APIs without modifying core application code

## Overview

Celeste integrates with the existing Online Boutique microservices architecture as a new AI-powered service. Rather than modifying the original application, Celeste operates as an intelligent interface that calls into the established APIs of the e-commerce platform, providing natural language shopping assistance to customers.

## Architecture

- **Base Application**: Google Online Boutique (microservices demo)
- **Language**: Go 1.21+
- **AI Model**: Google Gemini 2.5 Flash
- **Framework**: Gorilla Mux for HTTP routing
- **Deployment Platform**: Google Kubernetes Engine (GKE)
- **Integration Pattern**: REST API that interfaces with Online Boutique services

## Features

- Natural language product queries
- AI-powered conversational responses using Gemini
- RESTful API for easy integration with frontend
- Kubernetes-ready deployment for GKE
- Health check endpoints for service monitoring
- Lightweight microservice following Online Boutique patterns

## API Endpoints

### GET /
Returns a welcome message from Celeste.

### POST /chat
Processes natural language queries and returns AI-generated responses.

**Request Body:**
```json
{
  "query": "Can you help me find a dress for a wedding?"
}
