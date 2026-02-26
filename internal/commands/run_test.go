package commands

import (
	"bytes"
	"strings"
	"testing"
)

func TestRun_RoutingContracts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantCode    int
		wantOutHelp bool
		wantErrHelp bool
		wantErrSub  string
		wantOutSub  string
	}{
		{
			name:        "no command prints help to errOut",
			args:        []string{"memd"},
			wantCode:    ExitUsage,
			wantErrHelp: true,
		},
		{
			name:        "help prints help to out",
			args:        []string{"memd", "help"},
			wantCode:    ExitOK,
			wantOutHelp: true,
		},
		{
			name:        "unknown command prints error and help to errOut",
			args:        []string{"memd", "nope"},
			wantCode:    ExitUsage,
			wantErrHelp: true,
			wantErrSub:  `memd: unknown command "nope"`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var out, errOut bytes.Buffer
			got := Run(tt.args, &out, &errOut)

			if got != tt.wantCode {
				t.Fatalf("Run(%v) = %d, want %d; out=%q errOut=%q",
					tt.args, got, tt.wantCode, out.String(), errOut.String())
			}

			if tt.wantOutSub != "" && !strings.Contains(out.String(), tt.wantOutSub) {
				t.Fatalf("out missing %q; got:\n%s", tt.wantOutSub, out.String())
			}
			if tt.wantErrSub != "" && !strings.Contains(errOut.String(), tt.wantErrSub) {
				t.Fatalf("errOut missing %q; got:\n%s", tt.wantErrSub, errOut.String())
			}

			if tt.wantOutHelp {
				mustLookLikeHelp(t, out.String())
			} else if out.Len() != 0 {
				t.Fatalf("out = %q, want empty", out.String())
			}

			if tt.wantErrHelp {
				mustLookLikeHelp(t, errOut.String())
			} else if errOut.Len() != 0 {
				t.Fatalf("errOut = %q, want empty", errOut.String())
			}
		})
	}
}

func mustLookLikeHelp(t *testing.T, s string) {
	t.Helper()

	// Validate help is being printed without freezing formatting.
	for _, want := range []string{
		"Usage:",
		"memd <command>",
		"Commands:",
		"Exit codes:",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("expected help output to contain %q; got:\n%s", want, s)
		}
	}
}
