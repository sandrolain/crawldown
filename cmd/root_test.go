package main

import "testing"

func TestValidateGetInvocation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		options *getOptions
		args    []string
		wantErr bool
	}{
		{
			name:    "accepts root default invocation",
			options: &getOptions{outputDir: "./out"},
			args:    []string{"https://example.com"},
		},
		{
			name:    "accepts single page without positional url",
			options: &getOptions{outputDir: "./out", singleURL: "https://example.com/page"},
			args:    nil,
		},
		{
			name:    "requires output flag",
			options: &getOptions{},
			args:    []string{"https://example.com"},
			wantErr: true,
		},
		{
			name:    "requires url or single flag",
			options: &getOptions{outputDir: "./out"},
			args:    nil,
			wantErr: true,
		},
		{
			name:    "rejects too many positional args",
			options: &getOptions{outputDir: "./out"},
			args:    []string{"https://example.com", "https://example.org"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateGetInvocation(test.options, test.args)
			if test.wantErr && err == nil {
				t.Fatal("expected an error but got nil")
			}

			if !test.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
