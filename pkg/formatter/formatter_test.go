package formatter

import (
	"testing"
)

func TestFormatKWhPrice(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		expected string
	}{
		{"zero", 0.0, "0,0000 €"},
		{"positive", 0.3541, "0,3541 €"},
		{"rounding", 0.35416, "0,3542 €"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatKWhPrice(tt.price)
			if result != tt.expected {
				t.Errorf("FormatKWhPrice(%.4f) = %v, want %v", tt.price, result, tt.expected)
			}
		})
	}
}

func TestFormatPrice(t *testing.T) {
	tests := []struct {
		name     string
		price    float64
		expected string
	}{
		{"zero", 0.0, "0,00 €"},
		{"positive", 12.345, "12,35 €"},
		{"thousands", 1234.56, "1.234,56 €"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPrice(tt.price)
			if result != tt.expected {
				t.Errorf("FormatPrice(%.2f) = %v, want %v", tt.price, result, tt.expected)
			}
		})
	}
}
