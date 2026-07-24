package results

import "testing"

func TestNormalizeFIN(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "normalizes lowercase", input: "ab12cd3", want: "AB12CD3"},
		{name: "trims whitespace", input: "  a1b2c3d ", want: "A1B2C3D"},
		{name: "rejects short value", input: "A1B2C3", wantErr: true},
		{name: "rejects special character", input: "A1B-2C3", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := NormalizeFIN(test.input)
			if (err != nil) != test.wantErr {
				t.Fatalf("NormalizeFIN() error = %v, wantErr %v", err, test.wantErr)
			}
			if got != test.want {
				t.Fatalf("NormalizeFIN() = %q, want %q", got, test.want)
			}
		})
	}
}
