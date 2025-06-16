package utils

import (
	"testing"
)

type TestStruct struct {
	Name    string            `json:"name"`
	Age     int               `json:"age"`
	Active  bool              `json:"active"`
	Score   float64           `json:"score"`
	Tags    []string          `json:"tags"`
	Config  map[string]string `json:"config"`
	Nested  NestedStruct      `json:"nested"`
	Ignored string            `json:"-"`
	NoTag   string
}

type NestedStruct struct {
	ID    int    `json:"id"`
	Value string `json:"value"`
}

func TestMapToStruct_Success(t *testing.T) {
	data := map[string]interface{}{
		"name":   "John Doe",
		"age":    30,
		"active": true,
		"score":  95.5,
		"tags":   []interface{}{"go", "programming"},
		"config": map[string]interface{}{
			"theme": "dark",
			"lang":  "en",
		},
		"nested": map[string]interface{}{
			"id":    123,
			"value": "test",
		},
	}

	var result TestStruct
	err := MapToStruct(data, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != "John Doe" {
		t.Errorf("Expected Name to be 'John Doe', got %s", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("Expected Age to be 30, got %d", result.Age)
	}
	if !result.Active {
		t.Errorf("Expected Active to be true")
	}
	if result.Score != 95.5 {
		t.Errorf("Expected Score to be 95.5, got %f", result.Score)
	}
	if len(result.Tags) != 2 || result.Tags[0] != "go" || result.Tags[1] != "programming" {
		t.Errorf("Expected Tags to be ['go', 'programming'], got %v", result.Tags)
	}
	if result.Config["theme"] != "dark" || result.Config["lang"] != "en" {
		t.Errorf("Expected Config to be correct, got %v", result.Config)
	}
	if result.Nested.ID != 123 || result.Nested.Value != "test" {
		t.Errorf("Expected Nested to be correct, got %v", result.Nested)
	}
}

func TestMapToStruct_NotPointer(t *testing.T) {
	data := map[string]interface{}{"name": "test"}
	var result TestStruct

	err := MapToStruct(data, result)

	if err == nil {
		t.Fatal("Expected error for non-pointer target")
	}
	if err.Error() != "target must be a pointer to a struct" {
		t.Errorf("Expected specific error message, got %s", err.Error())
	}
}

func TestMapToStruct_NotStruct(t *testing.T) {
	data := map[string]interface{}{"name": "test"}
	var result string

	err := MapToStruct(data, &result)

	if err == nil {
		t.Fatal("Expected error for non-struct target")
	}
	if err.Error() != "target must point to a struct" {
		t.Errorf("Expected specific error message, got %s", err.Error())
	}
}

func TestMapToStruct_TypeConversions(t *testing.T) {
	type ConversionStruct struct {
		StringFromInt   string  `json:"string_from_int"`
		BoolFromString  bool    `json:"bool_from_string"`
		IntFromString   int     `json:"int_from_string"`
		IntFromFloat    int     `json:"int_from_float"`
		FloatFromInt    float64 `json:"float_from_int"`
		FloatFromString float64 `json:"float_from_string"`
		UintFromString  uint    `json:"uint_from_string"`
	}

	data := map[string]interface{}{
		"string_from_int":   42,
		"bool_from_string":  "true",
		"int_from_string":   "123",
		"int_from_float":    45.7,
		"float_from_int":    100,
		"float_from_string": "3.14",
		"uint_from_string":  "456",
	}

	var result ConversionStruct
	err := MapToStruct(data, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.StringFromInt != "42" {
		t.Errorf("Expected StringFromInt to be '42', got %s", result.StringFromInt)
	}
	if !result.BoolFromString {
		t.Errorf("Expected BoolFromString to be true")
	}
	if result.IntFromString != 123 {
		t.Errorf("Expected IntFromString to be 123, got %d", result.IntFromString)
	}
	if result.IntFromFloat != 45 {
		t.Errorf("Expected IntFromFloat to be 45, got %d", result.IntFromFloat)
	}
	if result.FloatFromInt != 100.0 {
		t.Errorf("Expected FloatFromInt to be 100.0, got %f", result.FloatFromInt)
	}
	if result.FloatFromString != 3.14 {
		t.Errorf("Expected FloatFromString to be 3.14, got %f", result.FloatFromString)
	}
	if result.UintFromString != 456 {
		t.Errorf("Expected UintFromString to be 456, got %d", result.UintFromString)
	}
}

func TestMapToStruct_MissingFields(t *testing.T) {
	data := map[string]interface{}{
		"name": "partial",
	}

	var result TestStruct
	err := MapToStruct(data, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != "partial" {
		t.Errorf("Expected Name to be 'partial', got %s", result.Name)
	}
	if result.Age != 0 {
		t.Errorf("Expected Age to be 0 (default), got %d", result.Age)
	}
}

func TestMapToStruct_NilValues(t *testing.T) {
	data := map[string]interface{}{
		"name": nil,
		"age":  25,
	}

	var result TestStruct
	err := MapToStruct(data, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != "" {
		t.Errorf("Expected Name to be empty for nil value, got %s", result.Name)
	}
	if result.Age != 25 {
		t.Errorf("Expected Age to be 25, got %d", result.Age)
	}
}

func TestMapToStruct_InvalidConversions(t *testing.T) {
	type InvalidStruct struct {
		Number int `json:"number"`
	}

	data := map[string]interface{}{
		"number": "not_a_number",
	}

	var result InvalidStruct
	err := MapToStruct(data, &result)

	if err == nil {
		t.Fatal("Expected error for invalid conversion")
	}
}

func TestMapToStruct_JSONTagWithOptions(t *testing.T) {
	type TagStruct struct {
		Name string `json:"name,omitempty"`
		Age  int    `json:"age,required"`
	}

	data := map[string]interface{}{
		"name": "test",
		"age":  30,
	}

	var result TagStruct
	err := MapToStruct(data, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != "test" {
		t.Errorf("Expected Name to be 'test', got %s", result.Name)
	}
	if result.Age != 30 {
		t.Errorf("Expected Age to be 30, got %d", result.Age)
	}
}

func TestMapToStruct_EmptyMap(t *testing.T) {
	data := map[string]interface{}{}

	var result TestStruct
	err := MapToStruct(data, &result)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// All fields should have their zero values
	if result.Name != "" || result.Age != 0 || result.Active != false {
		t.Errorf("Expected zero values for all fields")
	}
}
