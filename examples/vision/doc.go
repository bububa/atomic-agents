// Package vision Basic Vision Example
// This example demonstrates how to use the Atomic Agents framework to analyze images with text, specifically focusing on extracting structured information from nutrition labels using GPT-4 Vision capabilities.
//
// ## Features
//
// - Image Analysis: Process nutrition label images using GPT-4 Vision
// - Structured Data Extraction: Convert visual information into structured Pydantic models
// - Multi-Image Processing: Analyze multiple nutrition labels simultaneously
// - Comprehensive Nutritional Data: Extract detailed nutritional information including:
//   - Basic nutritional facts (calories, fats, proteins, etc.)
//   - Serving size information
//   - Vitamin and mineral content
//   - Product details
//
// ## Components
//
//  1. Nutrition Label Schema (`NutritionLabel`)
//     Defines the structure for storing nutrition information, including:
//     - Macronutrients (fats, proteins, carbohydrates)
//     - Micronutrients (vitamins and minerals)
//     - Serving information
//     - Product details
//  2. Input/Output Schemas
//     - `NutritionAnalysisInput`: Handles input images and analysis instructions
//     - `NutritionAnalysisOutput`: Structures the extracted nutrition information
//  3. Nutrition Analyzer Agent
//     A specialized agent configured with:
//     - GPT-4 Vision capabilities
//     - Custom system prompts for nutrition label analysis
//     - Structured data validation
package vision
