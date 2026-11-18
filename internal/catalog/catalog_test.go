package catalog

import "testing"

func TestEscapeLike(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"%", `\%`},
		{"_", `\_`},
		{`\`, `\\`},
		{`a%b_c\`, `a\%b\_c\\`},
		{`\%`, `\\\%`}, // single pass: the added \ is not escaped again
		{"Smart", "Smart"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			if got := escapeLike(tt.in); got != tt.want {
				t.Errorf("escapeLike(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
