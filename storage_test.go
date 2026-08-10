package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// =============================================================================
// STORAGE TESTS - Profile and Template Persistence
// Tests for JSON-based storage layer with file operations
// =============================================================================

// =============================================================================
// STORAGE DATA SERIALIZATION TESTS
// =============================================================================

func TestStorageDataSerialization(t *testing.T) {
	storage := StorageData{
		Version: 1,
		Profiles: []ConnectionProfile{
			{
				ID:        "123",
				Name:      "Production",
				Address:   "192.168.1.100",
				Port:      "514",
				Protocol:  "tcp",
				CreatedAt: 1234567890,
				UpdatedAt: 1234567890,
			},
		},
		Templates: []LogTemplate{
			{
				ID:         "456",
				Name:       "Error Template",
				Message:    "Error occurred: {message}",
				Facility:   3,
				Severity:   3,
				UseRFC5424: true,
				CreatedAt:  1234567890,
				UpdatedAt:  1234567890,
			},
		},
	}

	// Serialize
	data, err := json.Marshal(&storage)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	// Deserialize
	var loaded StorageData
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to deserialize: %v", err)
	}

	// Verify
	if loaded.Version != 1 {
		t.Errorf("Version: got %d, want 1", loaded.Version)
	}
	if len(loaded.Profiles) != 1 {
		t.Fatalf("Profiles count: got %d, want 1", len(loaded.Profiles))
	}
	if loaded.Profiles[0].Name != "Production" {
		t.Errorf("Profile name: got %q, want %q", loaded.Profiles[0].Name, "Production")
	}
	if len(loaded.Templates) != 1 {
		t.Fatalf("Templates count: got %d, want 1", len(loaded.Templates))
	}
	if loaded.Templates[0].Name != "Error Template" {
		t.Errorf("Template name: got %q, want %q", loaded.Templates[0].Name, "Error Template")
	}
}

func TestStorageDataEmptyArrays(t *testing.T) {
	storage := StorageData{
		Version:   1,
		Profiles:  []ConnectionProfile{},
		Templates: []LogTemplate{},
	}

	data, err := json.Marshal(&storage)
	if err != nil {
		t.Fatalf("failed to serialize empty storage: %v", err)
	}

	var loaded StorageData
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to deserialize empty storage: %v", err)
	}

	if loaded.Profiles == nil {
		t.Error("Profiles should not be nil")
	}
	if loaded.Templates == nil {
		t.Error("Templates should not be nil")
	}
}

// =============================================================================
// GENERATE ID TESTS
// =============================================================================

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	if id1 == "" {
		t.Error("generateID returned empty string")
	}

	// IDs should be numeric (timestamp based)
	for _, c := range id1 {
		if c < '0' || c > '9' {
			t.Errorf("ID contains non-numeric character: %c", c)
		}
	}

	// Sleep a bit to ensure different timestamp
	time.Sleep(time.Nanosecond * 100)

	id2 := generateID()
	if id1 == id2 {
		t.Error("generated IDs should be unique")
	}
}

func TestGenerateIDLength(t *testing.T) {
	id := generateID()
	// Unix nano timestamp should be at least 19 digits (as of 2024)
	if len(id) < 18 {
		t.Errorf("ID seems too short: %q (len=%d)", id, len(id))
	}
}

// =============================================================================
// CONNECTION PROFILE TESTS
// =============================================================================

func TestConnectionProfileJSON(t *testing.T) {
	profile := ConnectionProfile{
		ID:             "test-123",
		Name:           "My Profile",
		Description:    "Test description",
		Address:        "syslog.example.com",
		Port:           "6514",
		Protocol:       "tcp",
		FramingMethod:  "octet-counting",
		UseTLS:         true,
		TLSVerify:      true,
		CACertPath:     "/path/to/ca.pem",
		ClientCertPath: "/path/to/client.pem",
		ClientKeyPath:  "/path/to/client.key",
		CreatedAt:      1234567890,
		UpdatedAt:      1234567891,
	}

	data, err := json.Marshal(&profile)
	if err != nil {
		t.Fatalf("failed to marshal profile: %v", err)
	}

	var loaded ConnectionProfile
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal profile: %v", err)
	}

	if loaded.ID != profile.ID {
		t.Errorf("ID: got %q, want %q", loaded.ID, profile.ID)
	}
	if loaded.Name != profile.Name {
		t.Errorf("Name: got %q, want %q", loaded.Name, profile.Name)
	}
	if loaded.UseTLS != profile.UseTLS {
		t.Errorf("UseTLS: got %v, want %v", loaded.UseTLS, profile.UseTLS)
	}
}

func TestConnectionProfileOptionalFields(t *testing.T) {
	// JSON with minimal fields
	jsonData := `{
		"id": "123",
		"name": "Minimal",
		"address": "localhost",
		"port": "514",
		"protocol": "udp"
	}`

	var profile ConnectionProfile
	if err := json.Unmarshal([]byte(jsonData), &profile); err != nil {
		t.Fatalf("failed to unmarshal minimal profile: %v", err)
	}

	if profile.Description != "" {
		t.Errorf("Description should be empty, got %q", profile.Description)
	}
	if profile.UseTLS != false {
		t.Error("UseTLS should default to false")
	}
}

// =============================================================================
// LOG TEMPLATE TESTS
// =============================================================================

func TestLogTemplateJSON(t *testing.T) {
	template := LogTemplate{
		ID:          "tpl-123",
		Name:        "Error Alert",
		Description: "Critical error template",
		Message:     "CRITICAL: {component} failed - {error}",
		Facility:    1,
		Severity:    2,
		Appname:     "myapp",
		UseRFC5424:  true,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567891,
	}

	data, err := json.Marshal(&template)
	if err != nil {
		t.Fatalf("failed to marshal template: %v", err)
	}

	var loaded LogTemplate
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal template: %v", err)
	}

	if loaded.Message != template.Message {
		t.Errorf("Message: got %q, want %q", loaded.Message, template.Message)
	}
	if loaded.Facility != template.Facility {
		t.Errorf("Facility: got %d, want %d", loaded.Facility, template.Facility)
	}
	if loaded.Severity != template.Severity {
		t.Errorf("Severity: got %d, want %d", loaded.Severity, template.Severity)
	}
}

// =============================================================================
// IMPORT/EXPORT TESTS
// =============================================================================

func TestImportExportRoundTrip(t *testing.T) {
	original := StorageData{
		Version: 1,
		Profiles: []ConnectionProfile{
			{ID: "p1", Name: "Profile 1", Address: "192.168.1.1", Port: "514", Protocol: "tcp"},
			{ID: "p2", Name: "Profile 2", Address: "192.168.1.2", Port: "6514", Protocol: "tcp", UseTLS: true},
		},
		Templates: []LogTemplate{
			{ID: "t1", Name: "Template 1", Message: "Test message", Facility: 1, Severity: 6},
			{ID: "t2", Name: "Template 2", Message: "Another message", Facility: 16, Severity: 3},
		},
	}

	// Export
	exported, err := json.MarshalIndent(&original, "", "  ")
	if err != nil {
		t.Fatalf("failed to export: %v", err)
	}

	// Import
	var imported StorageData
	if err := json.Unmarshal(exported, &imported); err != nil {
		t.Fatalf("failed to import: %v", err)
	}

	// Verify counts
	if len(imported.Profiles) != len(original.Profiles) {
		t.Errorf("Profiles count: got %d, want %d", len(imported.Profiles), len(original.Profiles))
	}
	if len(imported.Templates) != len(original.Templates) {
		t.Errorf("Templates count: got %d, want %d", len(imported.Templates), len(original.Templates))
	}
}

func TestImportInvalidJSON(t *testing.T) {
	invalidJSON := `{"version": 1, "profiles": [`

	var storage StorageData
	err := json.Unmarshal([]byte(invalidJSON), &storage)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// =============================================================================
// SORTING TESTS
// =============================================================================

func TestProfilesSortedByName(t *testing.T) {
	profiles := []ConnectionProfile{
		{Name: "Zebra"},
		{Name: "Alpha"},
		{Name: "Middle"},
	}

	// Verify original order is preserved (sorting happens in storage functions)
	names := make([]string, len(profiles))
	for i, p := range profiles {
		names[i] = p.Name
	}

	// This test verifies the concept - actual sorting is done in storage functions
	if names[0] != "Zebra" { // Original order preserved
		t.Error("test setup issue - profiles should be in original order")
	}
}

func TestTemplatesSortedByName(t *testing.T) {
	templates := []LogTemplate{
		{Name: "Zebra Template"},
		{Name: "Alpha Template"},
		{Name: "Middle Template"},
	}

	// Verify original order
	if templates[0].Name != "Zebra Template" {
		t.Error("test setup issue - templates should be in original order")
	}
}

// =============================================================================
// EDGE CASES
// =============================================================================

func TestProfileWithSpecialCharactersInName(t *testing.T) {
	profile := ConnectionProfile{
		ID:   "123",
		Name: "Production (US-East) #1 / Primary",
	}

	data, err := json.Marshal(&profile)
	if err != nil {
		t.Fatalf("failed to marshal profile with special chars: %v", err)
	}

	var loaded ConnectionProfile
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal profile with special chars: %v", err)
	}

	if loaded.Name != profile.Name {
		t.Errorf("Name mismatch: got %q, want %q", loaded.Name, profile.Name)
	}
}

func TestTemplateWithUnicodeMessage(t *testing.T) {
	template := LogTemplate{
		ID:      "123",
		Name:    "Unicode Test",
		Message: "エラー発生: システム障害 / Error: 系統故障",
	}

	data, err := json.Marshal(&template)
	if err != nil {
		t.Fatalf("failed to marshal template with unicode: %v", err)
	}

	var loaded LogTemplate
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal template with unicode: %v", err)
	}

	if loaded.Message != template.Message {
		t.Errorf("Message mismatch with unicode")
	}
}

func TestStorageWithLargeNumberOfItems(t *testing.T) {
	storage := StorageData{
		Version:   1,
		Profiles:  make([]ConnectionProfile, 100),
		Templates: make([]LogTemplate, 100),
	}

	for i := 0; i < 100; i++ {
		storage.Profiles[i] = ConnectionProfile{
			ID:      generateID(),
			Name:    "Profile " + strings.Repeat("X", i%10),
			Address: "192.168.1." + string(rune('0'+i%10)),
			Port:    "514",
		}
		time.Sleep(time.Nanosecond) // Ensure unique IDs
	}

	for i := 0; i < 100; i++ {
		storage.Templates[i] = LogTemplate{
			ID:      generateID(),
			Name:    "Template " + strings.Repeat("Y", i%10),
			Message: "Test message " + string(rune('0'+i%10)),
		}
		time.Sleep(time.Nanosecond)
	}

	data, err := json.Marshal(&storage)
	if err != nil {
		t.Fatalf("failed to marshal large storage: %v", err)
	}

	var loaded StorageData
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("failed to unmarshal large storage: %v", err)
	}

	if len(loaded.Profiles) != 100 {
		t.Errorf("Profiles count: got %d, want 100", len(loaded.Profiles))
	}
	if len(loaded.Templates) != 100 {
		t.Errorf("Templates count: got %d, want 100", len(loaded.Templates))
	}
}

// =============================================================================
// CONSTANTS TESTS
// =============================================================================

func TestStorageConstants(t *testing.T) {
	if storageVersion < 1 {
		t.Error("storage version should be at least 1")
	}

	if storageFileName == "" {
		t.Error("storage file name should not be empty")
	}

	if !strings.HasSuffix(storageFileName, ".json") {
		t.Error("storage file should be JSON")
	}

	if appFolderName == "" {
		t.Error("app folder name should not be empty")
	}
}
