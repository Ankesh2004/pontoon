package main

import (
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/Ankesh2004/pontoon/internal/ai"
)

func main() {
	if err := godotenv.Load("../../.env"); err != nil {
		fmt.Println("Warning: Could not load .env file")
	}
	
	// Fallback to local .env
	godotenv.Load()

	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		panic("GROQ_API_KEY is missing")
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://pontoon:pontoon@localhost:5432/pontoon?sslmode=disable"
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	defer pool.Close()

	orchestrator := ai.NewOrchestrator(apiKey, pool)

	pipelineID := uuid.New()
	projectID := uuid.New()
	deploymentID := uuid.New()

	rawLogs := `
> pontoon@1.0.0 build
> vite build

vite v5.0.0 building for production...
transforming...
✓ 1500 modules transformed.
rendering chunks...
[vite:css] Unexpected character '@'
file: /app/src/index.css:10:1
error during build:
SyntaxError: Unexpected character '@'
    at ...
`

	fmt.Printf("Starting pipeline %s...\n", pipelineID)

	pCtx := &ai.PipelineContext{
		PipelineID:   pipelineID,
		ProjectID:    projectID,
		DeploymentID: deploymentID,
		RawLogs:      rawLogs,
	}

	err = orchestrator.RunPipeline(ctx, pCtx)
	if err != nil {
		fmt.Printf("Pipeline failed: %v\n", err)
	} else {
		fmt.Println("Pipeline succeeded (waiting approval)!")
	}

	fmt.Println("Final Context State:")
	if pCtx.ParsedError != nil {
		fmt.Printf("- Parsed Error: %s\n", *pCtx.ParsedError)
	}
	if pCtx.RootCause != nil {
		fmt.Printf("- Root Cause: %s\n", *pCtx.RootCause)
	}
	if pCtx.ProposedPatch != nil {
		fmt.Printf("- Proposed Patch:\n%s\n", *pCtx.ProposedPatch)
	}
	fmt.Printf("- Security Passed: %v\n", pCtx.SecurityPassed)
	if pCtx.ConfidenceScore != nil {
		fmt.Printf("- Confidence Score: %d\n", *pCtx.ConfidenceScore)
	}
}
