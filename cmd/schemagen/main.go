package main

import (
	"flag"
	"fmt"
	"github.com/DjordjeVuckovic/news-hunter/pkg/apis/datamapping"
	"github.com/DjordjeVuckovic/news-hunter/pkg/schema"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var (
		outputDir = flag.String("output", "api", "Output directory for generated schemas")
		// format    = flag.String("format", "json", "Output format: json, yaml")
	)
	flag.Parse()

	if err := os.MkdirAll(*outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	generator := schema.NewGenerator()

	// Generate schema for DataMapping
	schemaJSON, err := generator.GenerateJSONSchema(datamapping.DataMapper{})
	if err != nil {
		log.Fatalf("Failed to generate schema for DataMapping: %v", err)
	}

	// Write JSON schema
	jsonFile := filepath.Join(*outputDir, "datamapping-v1.json")
	if err := os.WriteFile(jsonFile, []byte(schemaJSON), 0644); err != nil {
		log.Fatalf("Failed to write JSON schema: %v", err)
	}

	fmt.Printf("Generated JSON schema: %s\n", jsonFile)

	// Generate YAML example
	yamlExample := generateYAMLExample()
	yamlFile := filepath.Join(*outputDir, "datamapping-example.yaml")
	if err := os.WriteFile(yamlFile, []byte(yamlExample), 0644); err != nil {
		log.Fatalf("Failed to write YAML example: %v", err)
	}

	fmt.Printf("Generated YAML example: %s\n", yamlFile)
}

func generateYAMLExample() string {
	return `# DataMapping Example Configuration
# This file demonstrates the structure for defining field mappings

kind: DataMapping
version: v1
metadata:
  name: "Kaggle News Dataset"
  description: "Field mapping configuration for Kaggle news dataset import"
dataset: "kaggle"
dateFormat: "2006-01-02T15:04:05Z"
fieldMappings:
  - source: "title"
    sourceType: "string"
    target: "Title"
    targetType: "string"
    required: true
  - source: "content" 
    sourceType: "string"
    target: "Content"
    targetType: "string"
    required: true
  - source: "author"
    sourceType: "string" 
    target: "Author"
    targetType: "string"
    required: false
  - source: "published_date"
    sourceType: "datetime"
    target: "CreatedAt"
    targetType: "datetime"
    required: false
  - source: "language"
    sourceType: "string"
    target: "Language" 
    targetType: "string"
    required: false
`
}
