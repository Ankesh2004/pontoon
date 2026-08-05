package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
)

const groqModel = "llama-3.3-70b-versatile"

// --- Helper for Groq calls ---

func callGroqJSON(ctx context.Context, client *openai.Client, systemPrompt, userPrompt string, result interface{}) error {
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: groqModel,
			Messages: []openai.ChatCompletionMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userPrompt},
			},
			ResponseFormat: &openai.ChatCompletionResponseFormat{
				Type: openai.ChatCompletionResponseFormatTypeJSONObject,
			},
		},
	)
	if err != nil {
		return err
	}

	content := resp.Choices[0].Message.Content
	return json.Unmarshal([]byte(content), result)
}

// --- Inspector Agent ---

type InspectorAgent struct {
	client *openai.Client
}

func NewInspectorAgent(client *openai.Client) *InspectorAgent {
	return &InspectorAgent{client: client}
}

func (a *InspectorAgent) Name() string { return "Inspector" }

func (a *InspectorAgent) Execute(ctx context.Context, pCtx *PipelineContext) error {
	systemPrompt := `You are the Inspector Agent in a CI/CD pipeline. 
Your job is to extract only the critical error trace from the provided raw logs.
Output strictly in JSON format with fields:
- "parsed_error": string containing the extracted error trace.
- "confidence_score": integer 0-100 indicating your confidence in finding the error.`

	userPrompt := fmt.Sprintf("Analyze these raw build logs:\n%s", pCtx.RawLogs)

	var output struct {
		ParsedError     string `json:"parsed_error"`
		ConfidenceScore int    `json:"confidence_score"`
	}

	if err := callGroqJSON(ctx, a.client, systemPrompt, userPrompt, &output); err != nil {
		return err
	}

	pCtx.ParsedError = &output.ParsedError
	pCtx.ConfidenceScore = &output.ConfidenceScore
	return nil
}

// --- Diagnostician Agent ---

type DiagnosticianAgent struct {
	client *openai.Client
}

func NewDiagnosticianAgent(client *openai.Client) *DiagnosticianAgent {
	return &DiagnosticianAgent{client: client}
}

func (a *DiagnosticianAgent) Name() string { return "Diagnostician" }

func (a *DiagnosticianAgent) Execute(ctx context.Context, pCtx *PipelineContext) error {
	systemPrompt := `You are the Diagnostician Agent.
Your job is to perform Root Cause Analysis (RCA) on the provided build error.
Output strictly in JSON format with fields:
- "root_cause": string containing your root cause analysis.
- "confidence_score": integer 0-100 indicating your confidence.`

	// Use previously extracted state
	parsedErr := ""
	if pCtx.ParsedError != nil {
		parsedErr = *pCtx.ParsedError
	}
	userPrompt := fmt.Sprintf("Analyze this error trace and provide the root cause:\n%s", parsedErr)

	var output struct {
		RootCause       string `json:"root_cause"`
		ConfidenceScore int    `json:"confidence_score"`
	}

	if err := callGroqJSON(ctx, a.client, systemPrompt, userPrompt, &output); err != nil {
		return err
	}

	pCtx.RootCause = &output.RootCause
	
	// Update aggregated confidence (average)
	if pCtx.ConfidenceScore != nil {
		avg := (*pCtx.ConfidenceScore + output.ConfidenceScore) / 2
		pCtx.ConfidenceScore = &avg
	} else {
		pCtx.ConfidenceScore = &output.ConfidenceScore
	}

	return nil
}

// --- Architect Agent ---

type ArchitectAgent struct {
	client *openai.Client
}

func NewArchitectAgent(client *openai.Client) *ArchitectAgent {
	return &ArchitectAgent{client: client}
}

func (a *ArchitectAgent) Name() string { return "Architect" }

func (a *ArchitectAgent) Execute(ctx context.Context, pCtx *PipelineContext) error {
	systemPrompt := `You are the Architect Agent.
Your job is to generate the exact unified diff or patched file to fix the problem based on the root cause.
Output strictly in JSON format with fields:
- "proposed_patch": string containing the exact unified diff or fix instructions.
- "confidence_score": integer 0-100 indicating your confidence in the fix.`

	rootCause := ""
	if pCtx.RootCause != nil {
		rootCause = *pCtx.RootCause
	}
	userPrompt := fmt.Sprintf("Generate a fix for this root cause:\n%s", rootCause)

	var output struct {
		ProposedPatch   string `json:"proposed_patch"`
		ConfidenceScore int    `json:"confidence_score"`
	}

	if err := callGroqJSON(ctx, a.client, systemPrompt, userPrompt, &output); err != nil {
		return err
	}

	// Make sure diffs format nicely in JSON
	cleanedPatch := strings.TrimSpace(output.ProposedPatch)
	pCtx.ProposedPatch = &cleanedPatch
	
	// Update aggregated confidence (average)
	if pCtx.ConfidenceScore != nil {
		avg := (*pCtx.ConfidenceScore + output.ConfidenceScore) / 2
		pCtx.ConfidenceScore = &avg
	} else {
		pCtx.ConfidenceScore = &output.ConfidenceScore
	}

	return nil
}

// --- SecOps Agent ---

type SecOpsAgent struct {
	client *openai.Client
}

func NewSecOpsAgent(client *openai.Client) *SecOpsAgent {
	return &SecOpsAgent{client: client}
}

func (a *SecOpsAgent) Name() string { return "SecOps" }

func (a *SecOpsAgent) Execute(ctx context.Context, pCtx *PipelineContext) error {
	systemPrompt := `You are the SecOps Agent.
Your job is to analyze the proposed patch for security vulnerabilities (e.g., untrusted packages, root execution, backdoors).
Output strictly in JSON format with fields:
- "safe": boolean true if safe, false if risky.
- "risk_summary": string summarizing any risks found, or "No risks found".
- "confidence_score": integer 0-100 indicating your confidence.`

	patch := ""
	if pCtx.ProposedPatch != nil {
		patch = *pCtx.ProposedPatch
	}
	userPrompt := fmt.Sprintf("Analyze this patch for security risks:\n%s", patch)

	var output struct {
		Safe            bool   `json:"safe"`
		RiskSummary     string `json:"risk_summary"`
		ConfidenceScore int    `json:"confidence_score"`
	}

	if err := callGroqJSON(ctx, a.client, systemPrompt, userPrompt, &output); err != nil {
		return err
	}

	pCtx.SecurityPassed = output.Safe
	
	// Update aggregated confidence (average)
	if pCtx.ConfidenceScore != nil {
		avg := (*pCtx.ConfidenceScore + output.ConfidenceScore) / 2
		pCtx.ConfidenceScore = &avg
	} else {
		pCtx.ConfidenceScore = &output.ConfidenceScore
	}

	// If secops flags it as unsafe, we can just fail the confidence to force a pipeline rejection
	if !output.Safe {
		zero := 0
		pCtx.ConfidenceScore = &zero
	}

	return nil
}
