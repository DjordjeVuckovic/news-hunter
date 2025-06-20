package apis

import "fmt"

type DataMapping struct {
	Kind          string         `json:"kind" example:"DataMapping" yaml:"kind"`
	Version       string         `json:"version" example:"v1" yaml:"version"`
	Metadata      Metadata       `json:"metadata" yaml:"metadata"`
	Dataset       string         `json:"dataset" example:"kaggle" yaml:"dataset"`
	FieldMappings []FieldMapping `json:"fieldMappings" yaml:"fieldMappings"`
	DateFormat    string         `json:"dateFormat" example:"2006-01-02T15:04:05Z" yaml:"dateFormat"`
}

type Metadata struct {
	Name        string `json:"name" example:"Kaggle Dataset" yaml:"name"`
	Description string `json:"description" example:"Mapping for Kaggle dataset fields" yaml:"description"`
}

type FieldMapping struct {
	Source     string `json:"source" example:"id" yaml:"source"`
	SourceType string `json:"sourceType" example:"string" yaml:"sourceType"`
	Target     string `json:"target" example:"string" yaml:"target"`
	TargetType string `json:"targetType" example:"string" yaml:"targetType"`
	Required   bool   `json:"required" example:"true" yaml:"required"`
}

func (dm *DataMapping) Validate() error {
	if dm.Kind == "" {
		return fmt.Errorf("kind is required")
	}
	if dm.Version == "" {
		return fmt.Errorf("version is required")
	}
	if dm.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}
	if dm.Dataset == "" {
		return fmt.Errorf("dataset is required")
	}
	if len(dm.FieldMappings) == 0 {
		return fmt.Errorf("at least one field mapping is required")
	}
	for i, fm := range dm.FieldMappings {
		if fm.Source == "" {
			return fmt.Errorf("fieldMappings[%d] must have source defined", i)
		}
	}
	return nil
}

type MappingError struct {
	Message string `json:"message" example:"missing source field: id"`
}

func (e *MappingError) Error() string {
	return fmt.Sprintf("mapping error: %s", e.Message)
}
