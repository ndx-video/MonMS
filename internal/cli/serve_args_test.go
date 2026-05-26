package cli

import "testing"

func TestEnsureServeSubcommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"empty", nil, []string{"serve"}},
		{"explicit serve", []string{"serve"}, []string{"serve"}},
		{"daemon flag only", []string{"-d"}, []string{"serve", "-d"}},
		{"dev flag", []string{"--dev"}, []string{"serve", "--dev"}},
		{"init unchanged", []string{"init"}, []string{"init"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnsureServeSubcommand(tt.args)
			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}
