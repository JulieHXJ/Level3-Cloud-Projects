package user

import "testing"

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "Cloud3!Secure42",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Aa1!",
			wantErr:  true,
		},
		{
			name:     "missing uppercase",
			password: "cloud3!secure42",
			wantErr:  true,
		},
		{
			name:     "missing lowercase",
			password: "CLOUD3!SECURE42",
			wantErr:  true,
		},
		{
			name:     "missing number",
			password: "Cloud!SecurePass",
			wantErr:  true,
		},
		{
			name:     "missing symbol",
			password: "Cloud3Secure42",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)

			if tt.wantErr && err == nil {
				t.Fatal("expected password validation to fail")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("expected password to be valid: %v", err)
			}
		})
	}
}