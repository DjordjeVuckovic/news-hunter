package dto

import (
	"testing"

	"github.com/google/uuid"
)

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		name        string
		rank        float32
		id          uuid.UUID
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid cursor",
			rank:    0.95,
			id:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			wantErr: false,
		},
		{
			name:        "nil UUID",
			rank:        0.5,
			id:          uuid.Nil,
			wantErr:     true,
			errContains: "cannot be nil",
		},
		{
			name:    "zero rank",
			rank:    0.0,
			id:      uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeCursor(tt.rank, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeCursor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("EncodeCursor() error = %v, should contain %v", err, tt.errContains)
				}
			}
			if !tt.wantErr && encoded == "" {
				t.Error("EncodeCursor() returned empty string for valid input")
			}
		})
	}
}

func TestDecodeCursor(t *testing.T) {
	validID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	validEncoded, _ := EncodeCursor(0.95, validID)

	tests := []struct {
		name        string
		encoded     string
		wantErr     bool
		errContains string
		wantNil     bool
	}{
		{
			name:    "valid cursor",
			encoded: validEncoded,
			wantErr: false,
		},
		{
			name:    "empty string returns nil",
			encoded: "",
			wantErr: false,
			wantNil: true,
		},
		{
			name:        "invalid base64",
			encoded:     "not-valid-base64!!!",
			wantErr:     true,
			errContains: "decode cursor",
		},
		{
			name:        "invalid JSON",
			encoded:     "aW52YWxpZC1qc29u", // base64 of "invalid-json"
			wantErr:     true,
			errContains: "unmarshal cursor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := DecodeCursor(tt.encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeCursor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("DecodeCursor() error = %v, should contain %v", err, tt.errContains)
				}
			}
			if tt.wantNil && decoded != nil {
				t.Error("DecodeCursor() should return nil for empty string")
			}
			if !tt.wantErr && !tt.wantNil && decoded == nil {
				t.Error("DecodeCursor() returned nil for valid input")
			}
		})
	}
}

func TestCursorRoundtrip(t *testing.T) {
	tests := []struct {
		name string
		rank float32
		id   uuid.UUID
	}{
		{
			name: "typical rank",
			rank: 0.75,
			id:   uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		},
		{
			name: "high rank",
			rank: 25.5,
			id:   uuid.MustParse("987fcdeb-51a2-43d7-b890-123456789abc"),
		},
		{
			name: "zero rank",
			rank: 0.0,
			id:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := EncodeCursor(tt.rank, tt.id)
			if err != nil {
				t.Fatalf("EncodeCursor() failed: %v", err)
			}

			decoded, err := DecodeCursor(encoded)
			if err != nil {
				t.Fatalf("DecodeCursor() failed: %v", err)
			}

			if decoded.Score != tt.rank {
				t.Errorf("Score mismatch: got %v, want %v", decoded.Score, tt.rank)
			}
			if decoded.ID != tt.id {
				t.Errorf("ID mismatch: got %v, want %v", decoded.ID, tt.id)
			}
		})
	}
}

func TestMustEncodeCursor(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		id := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
		defer func() {
			if r := recover(); r != nil {
				t.Error("MustEncodeCursor() panicked on valid input")
			}
		}()
		result := MustEncodeCursor(0.5, id)
		if result == "" {
			t.Error("MustEncodeCursor() returned empty string")
		}
	})

	t.Run("nil UUID panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustEncodeCursor() should panic on nil UUID")
			}
		}()
		MustEncodeCursor(0.5, uuid.Nil)
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
