// Package Orchestration # Orchestration Agent Example
// This example demonstrates how to create an Orchestrator Agent that intelligently decides between using a search tool or a calculator tool based on user input.
//
// ## Features
// - Intelligent tool selection between search and calculator tools
// - Dynamic input/output schema handling
// - Real-time date context provider
// - Rich console output formatting
// - Final answer generation based on tool outputs
//
// ## Components
//
// ### Input/Output Schemas
//
// - **OrchestratorInputSchema**: Handles user input messages
// - **OrchestratorOutputSchema**: Specifies tool selection and parameters
// - **FinalAnswerSchema**: Formats the final response
//
// ### Tools
// These tools were installed using the Atomic Assembler CLI (See the main README [here](../../README.md) for more info)
// The agent orchestrates between two tools:
// - **SearxNG Search Tool**: For queries requiring factual information
// - **Calculator Tool**: For mathematical calculations
//
// ### Context Providers
//
// - **CurrentDateProvider**: Provides the current date in YYYY-MM-DD format
package orchestration
