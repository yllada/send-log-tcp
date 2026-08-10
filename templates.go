package main

import (
	"context"
)

// ================================================================================
// TEMPLATE SERVICE - Manages log templates (Enterprise Feature)
// Follows Single Responsibility Principle: handles only template CRUD operations
// ================================================================================

// LogTemplate represents a reusable log message template
type LogTemplate struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Message     string `json:"message"`
	Facility    uint8  `json:"facility"`
	Severity    uint8  `json:"severity"`
	Appname     string `json:"appname"`
	// MessageFormat is optional so templates stored before it existed still
	// decode; an empty value falls back to UseRFC5424.
	MessageFormat string `json:"messageFormat,omitempty"`
	UseRFC5424    bool   `json:"useRfc5424"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
}

// TemplateService handles log template operations
type TemplateService struct {
	ctx context.Context
}

// NewTemplateService creates a new TemplateService instance
func NewTemplateService() *TemplateService {
	return &TemplateService{}
}

// SetContext sets the Wails runtime context
func (t *TemplateService) SetContext(ctx context.Context) {
	t.ctx = ctx
}

// GetTemplates returns all saved log templates
func (t *TemplateService) GetTemplates() ([]LogTemplate, error) {
	return getTemplatesFromStorage()
}

// SaveTemplate saves a new or updates an existing log template
func (t *TemplateService) SaveTemplate(template LogTemplate) (LogTemplate, error) {
	return saveTemplateToStorage(template)
}

// DeleteTemplate deletes a log template by ID
func (t *TemplateService) DeleteTemplate(id string) error {
	return deleteTemplateFromStorage(id)
}

// ================================================================================
// CONFIG IMPORT/EXPORT (Enterprise Feature)
// ================================================================================

// ExportConfig exports all profiles and templates to a JSON string
func (t *TemplateService) ExportConfig() (string, error) {
	return exportConfigFromStorage()
}

// ImportConfig imports profiles and templates from a JSON string
func (t *TemplateService) ImportConfig(jsonData string, merge bool) error {
	return importConfigToStorage(jsonData, merge)
}
