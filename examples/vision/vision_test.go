package vision

import (
	"context"
	"fmt"
	"os"

	"github.com/bububa/atomic-agents/agents"
	"github.com/bububa/atomic-agents/components"
	"github.com/bububa/atomic-agents/components/systemprompt/cot"
	"github.com/bububa/atomic-agents/examples"
	"github.com/bububa/atomic-agents/schema"
	"github.com/bububa/instructor-go"
)

// NutritionLabel Represents the complete nutritional information from a food label
type NutritionLabel struct {
	schema.Base
	// Calories per serving
	Calories int `json:"calories" jsonschema:"title=calories,description=Calories per serving"`
	// TotalFat Total fat in grams
	TotalFat int `json:"total_fat" jsonschema:"title=total_fat,description=Total fat in grams"`
	// SaturatedFat Saturated fat in grams
	SaturatedFat int `json:"saturated_fat" jsonschema:"title=saturated_fat,description=Saturated fat in grams"`
	// TransFat Trans fat in grams
	TransFat int `json:"trans_fat" jsonschema:"title=trans_fat,description=Trans fat in grams"`
	// Cholesterol in milligrams
	Cholesterol int `json:"cholesterol" jsonschema:"title=cholesterol,description=Cholesterol in milligrams"`
	// Sodium in milligrams
	Sodium int `json:"sodium" jsonschema:"title=sodium,description=Sodium in milligrams"`
	// TotalCarbohydrates Total carbohydrates in grams
	TotalCarbohydrates int `json:"total_carbohydrates" jsonschema:"title=total_carbohydrates,description=Total carbohydrates in grams"`
	// DietaryFiber Dietary fiber in grams
	DietaryFiber int `json:"dietary_fiber" jsonschema:"title=dietary_fiber,description=Dietary fiber in grams"`
	// TotalSugars Total sugars in grams
	TotalSugars int `json:"total_sugars" jsonschema:"title=total_sugars,description=Total sugars in grams"`
	// AddedSugars Added sugars in grams
	AddedSugars int `json:"added_sugars" jsonschema:"title=added_sugars,description=Added sugars in grams"`
	// Protein in grams
	Protein int `json:"protein" jsonschema:"title=protein,description=Protein in grams"`
	// VitaminD Vitamin D in micrograms
	VitaminD int `json:"vitamin_d" jsonschema:"title=vitamin_d,description=Vitamin D in micrograms"`
	// Calcium in milligrams
	Calcium int `json:"calcium" jsonschema:"title=calcium,description=Calcium in milligrams"`
	// Iron in milligrams
	Iron int `json:"iron" jsonschema:"title=iron,description=Iron in milligrams"`
	// Potassium in milligrams
	Potassium int `json:"potassium" jsonschema:"title=potassium,description=Potassium in milligrams"`
	// ServingSize The size of a single serving of this product
	ServingSize int `json:"serving_size" jsonschema:"title=serving_size,description=Serving size"`
	// ServingsPerContainer Number of servings contained in the package
	ServingsPerContainer int `json:"servings_per_container" jsonschema:"title=servings_per_container,description=Number of servings contained in the package"`
	// ProductName The full name or description of the type of the food/drink. e.g: 'Coca Cola Light', 'Pepsi Max', 'Smoked Bacon', 'Chianti Wine'
	ProductName string `json:"product_name" jsonschema:"title=product_name,description=The full name or description of the type of the food/drink. e.g: 'Coca Cola Light', 'Pepsi Max', 'Smoked Bacon', 'Chianti Wine'"`
}

// Input schema for nutrition label analysis
type Input struct {
	schema.Base
	// InstructionText The instruction for analyzing the nutrition label
	InstructionText string `json:"instruction_text" jsonschema:"title=instruction_text,description=The instruction for analyzing the nutrition label"`
}

// Output schema containing extracted nutrition information
type Output struct {
	schema.Base
	// AnalyzedLabels List of nutrition labels extracted from the provided images
	AnalyzedLabels []NutritionLabel `json:"analyzed_labels" jsonschema:"title=analyzed_labels,description=List of nutrition labels extracted from the provided images"`
}

func Example_vision() {
	ctx := context.Background()
	fmt.Println("Analyzing nutrition labels...")
	mem := components.NewMemory(10)
	systemPromptGenerator := cot.New(
		cot.WithBackground([]string{
			"You are a specialized nutrition label analyzer.",
			"You excel at extracting precise nutritional information from food label images.",
			"You understand various serving size formats and measurement units.",
			"You can process multiple nutrition labels simultaneously.",
		}),
		cot.WithSteps([]string{
			"For each nutrition label image:",
			"1. Locate and identify the nutrition facts panel",
			"2. Extract all serving information and nutritional values",
			"3. Validate measurements and units for accuracy",
			"4. Compile the nutrition facts into structured data",
		}),
		cot.WithOutputInstructs([]string{
			"For each analyzed nutrition label:",
			"1. Record complete serving size information",
			"2. Extract all nutrient values with correct units",
			"3. Ensure all measurements are properly converted",
			"4. Include all extracted labels in the final result",
		}),
	)
	agent := agents.NewAgent[Input, Output](
		agents.WithClient(examples.NewInstructor(instructor.ProviderOpenAI)),
		agents.WithMemory(mem),
		agents.WithModel(os.Getenv("OPENAI_VISION_MODEL")),
		agents.WithSystemPromptGenerator(systemPromptGenerator),
		agents.WithTemperature(0.5),
		agents.WithMaxTokens(1000),
	)

	attachement := schema.Attachement{
		ImageURLs: []string{
			"https://img1.baidu.com/it/u=3796584098,3946123104&fm=253&fmt=auto&app=138&f=JPEG?w=800&h=1424",
			"https://nutrition-ai.oss-ap-northeast-1.aliyuncs.com/log/img/10239186114263491872/13785780926862670528.jpg",
		},
	}
	input := &Input{
		InstructionText: "Please analyze these nutrition labels and extract all nutritional information.",
	}
	input.SetAttachement(&attachement)
	output := new(Output)
	apiResp := new(components.LLMResponse)
	if err := agent.Run(ctx, input, output, apiResp); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Analysis completed successfully")
	// Display the results
	for idx, label := range output.AnalyzedLabels {
		fmt.Printf("\nNutrition Label %d:\n", idx)
		fmt.Printf("Product Name: %s\n", label.ProductName)
		fmt.Printf("Serving Size: %d\n", label.ServingSize)
		fmt.Printf("Servings Per Container: %d\n", label.ServingsPerContainer)
		fmt.Printf("Calories: %d\n", label.Calories)
		fmt.Printf("Total Fat: %dg\n", label.TotalFat)
		fmt.Printf("Saturated Fat: %dg\n", label.SaturatedFat)
		fmt.Printf("Trans Fat: %dg\n", label.TransFat)
		fmt.Printf("Cholesterol: %dmg\n", label.Cholesterol)
		fmt.Printf("Sodium: %dmg\n", label.Sodium)
		fmt.Printf("Total Carbohydrates: %dg\n", label.TotalCarbohydrates)
		fmt.Printf("Dietary Fiber: %dg\n", label.DietaryFiber)
		fmt.Printf("Total Sugars: %dg\n", label.TotalSugars)
		fmt.Printf("Added Sugars: %dg\n", label.AddedSugars)
		fmt.Printf("Protein: %dg\n", label.Protein)
		fmt.Printf("Vitamin D: %dmcg\n", label.VitaminD)
		fmt.Printf("Calcium: %dmg\n", label.Calcium)
		fmt.Printf("Iron: %dmg\n", label.Iron)
		fmt.Printf("Potassium: %dmg\n", label.Potassium)
	}
	// Output:
	//
}
