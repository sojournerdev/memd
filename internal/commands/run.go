package commands

import "io"

func Run(args []string, out, errOut io.Writer) int {
	if len(args) < 2 {
		PrintHelp(errOut)
		return ExitUsage
	}

	switch args[1] {
	case "help":
		PrintHelp(out)
		return ExitOK
	case "version":
		return Version(out)
	case "init":
		return Init(out, errOut)
	case "doctor":
		return Doctor(out, errOut)
	default:
		tryWritef(errOut, "memd: unknown command %q\n\n", args[1])
		PrintHelp(errOut)
		return ExitUsage
	}
}

const (
	ExitOK    = 0
	ExitError = 1
	ExitUsage = 2
)
