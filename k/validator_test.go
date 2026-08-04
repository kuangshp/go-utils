package k

import (
	"encoding/json"
	"testing"
)

func TestIsNonEmptyJSONArray(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "object array", value: `[{"code":"1"}]`, want: true},
		{name: "mixed array", value: `[1,"two",null]`, want: true},
		{name: "empty array", value: `[]`, want: false},
		{name: "null", value: `null`, want: false},
		{name: "object", value: `{"code":"1"}`, want: false},
		{name: "invalid JSON", value: `[1,`, want: false},
		{name: "empty string", value: ``, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNonEmptyJSONArray(tt.value); got != tt.want {
				t.Fatalf("IsNonEmptyJSONArray(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsValidJSONArray(t *testing.T) {
	type optionData struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	type userData struct {
		ID     int  `json:"id"`
		Active bool `json:"active"`
	}

	tests := []struct {
		name     string
		validate func() bool
		want     bool
	}{
		{
			name: "valid options",
			validate: func() bool {
				value := `[{"code":"enabled","name":"Enabled"},{"code":"disabled","name":"Disabled"}]`
				return IsValidJSONArray[optionData](value, func(item optionData) bool {
					return item.Code != "" && item.Name != ""
				})
			},
			want: true,
		},
		{
			name: "invalid option",
			validate: func() bool {
				value := `[{"code":"enabled","name":""}]`
				return IsValidJSONArray[optionData](value, func(item optionData) bool {
					return item.Code != "" && item.Name != ""
				})
			},
			want: false,
		},
		{
			name: "valid users",
			validate: func() bool {
				value := `[{"id":1,"active":true},{"id":2,"active":true}]`
				return IsValidJSONArray[userData](value, func(item userData) bool {
					return item.ID > 0 && item.Active
				})
			},
			want: true,
		},
		{
			name: "invalid user",
			validate: func() bool {
				value := `[{"id":0,"active":true}]`
				return IsValidJSONArray[userData](value, func(item userData) bool {
					return item.ID > 0 && item.Active
				})
			},
			want: false,
		},
		{
			name: "nil validator",
			validate: func() bool {
				return IsValidJSONArray[json.RawMessage](`[1,2]`, nil)
			},
			want: true,
		},
		{
			name: "empty array",
			validate: func() bool {
				return IsValidJSONArray[optionData](`[]`, func(optionData) bool { return true })
			},
			want: false,
		},
		{
			name: "invalid JSON",
			validate: func() bool {
				return IsValidJSONArray[optionData](`[`, func(optionData) bool { return true })
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.validate(); got != tt.want {
				t.Fatalf("IsValidJSONArray() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNonEmptyJSONObject(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "non-empty object", value: `{"id":1}`, want: true},
		{name: "nested object", value: `{"user":{"id":1}}`, want: true},
		{name: "empty object", value: `{}`, want: false},
		{name: "null", value: `null`, want: false},
		{name: "array", value: `[{"id":1}]`, want: false},
		{name: "invalid JSON", value: `{"id":`, want: false},
		{name: "empty string", value: ``, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNonEmptyJSONObject(tt.value); got != tt.want {
				t.Fatalf("IsNonEmptyJSONObject(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestIsValidJSONObject(t *testing.T) {
	type userData struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	validateUser := func(item userData) bool {
		return item.ID > 0 && item.Name != ""
	}

	tests := []struct {
		name     string
		value    string
		validate func(userData) bool
		want     bool
	}{
		{
			name:     "valid object",
			value:    `{"id":1,"name":"Alice"}`,
			validate: validateUser,
			want:     true,
		},
		{
			name:     "extra fields are allowed",
			value:    `{"id":1,"name":"Alice","active":true}`,
			validate: validateUser,
			want:     true,
		},
		{
			name:     "custom validation failed",
			value:    `{"id":0,"name":"Alice"}`,
			validate: validateUser,
			want:     false,
		},
		{
			name:     "nil validator",
			value:    `{"id":1}`,
			validate: nil,
			want:     true,
		},
		{
			name:     "empty object",
			value:    `{}`,
			validate: validateUser,
			want:     false,
		},
		{
			name:     "array",
			value:    `[{"id":1,"name":"Alice"}]`,
			validate: validateUser,
			want:     false,
		},
		{
			name:     "wrong field type",
			value:    `{"id":"one","name":"Alice"}`,
			validate: validateUser,
			want:     false,
		},
		{
			name:     "invalid JSON",
			value:    `{"id":1`,
			validate: validateUser,
			want:     false,
		},
		{
			name:     "null",
			value:    `null`,
			validate: validateUser,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidJSONObject[userData](tt.value, tt.validate); got != tt.want {
				t.Fatalf("IsValidJSONObject(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
