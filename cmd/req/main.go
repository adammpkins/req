package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/adammpkins/req/internal/grammar"
	"github.com/adammpkins/req/internal/output"
	"github.com/adammpkins/req/internal/parser"
	"github.com/adammpkins/req/internal/planner"
	"github.com/adammpkins/req/internal/runtime"
	"github.com/adammpkins/req/internal/session"
	"github.com/adammpkins/req/internal/tui"
	"github.com/adammpkins/req/internal/types"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

func main() {
	var (
		showHelp    = flag.Bool("help", false, "Show help message")
		showVersion = flag.Bool("version", false, "Show version information")
		dryRun      = flag.Bool("dry-run", false, "Print execution plan without executing")
		tuiMode     = flag.Bool("tui", false, "Launch interactive TUI mode")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: req <verb> <target> [clauses...]\n\n")
		fmt.Fprintf(os.Stderr, "Verbs: read, save, send, upload, watch, inspect, authenticate, session\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  req read https://api.example.com/users as=json\n")
		fmt.Fprintf(os.Stderr, "  req send https://api.example.com/users with='{\"name\":\"Ada\"}'\n")
		fmt.Fprintf(os.Stderr, "  req send https://api.example.com/users using=PUT with='{\"name\":\"Ada\"}'\n")
		fmt.Fprintf(os.Stderr, "  req save https://example.com/file.zip to=file.zip\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	if *showHelp {
		flag.Usage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("req version %s (commit: %s, built: %s)\n", version, commit, buildDate)
		os.Exit(0)
	}

	// Launch TUI mode if requested
	if *tuiMode {
		if err := tui.Launch(); err != nil {
			printError(err)
			os.Exit(1)
		}
		return
	}

	// Get remaining args after flags
	args := flag.Args()

	// Remove any remaining flags from args (in case they appear after command args)
	filteredArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if arg == "--help" || arg == "-help" || arg == "-h" {
			flag.Usage()
			os.Exit(0)
		}
		if arg == "--version" || arg == "-version" || arg == "-v" {
			fmt.Printf("req version %s (commit: %s, built: %s)\n", version, commit, buildDate)
			os.Exit(0)
		}
		if arg == "--dry-run" || arg == "-dry-run" {
			*dryRun = true
			continue
		}
		filteredArgs = append(filteredArgs, arg)
	}
	args = filteredArgs

	// If no args provided, launch TUI mode
	if len(args) == 0 {
		if err := tui.Launch(); err != nil {
			printError(err)
			os.Exit(1)
		}
		return
	}

	// Handle help command
	if len(args) > 0 && args[0] == "help" {
		printHelp()
		os.Exit(0)
	}

	// Handle explain command
	if len(args) > 0 && args[0] == "explain" {
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "Usage: req explain \"<command>\"\n")
			os.Exit(5)
		}
		command := strings.Join(args[1:], " ")
		if err := explainCommand(command); err != nil {
			printError(err)
			os.Exit(5)
		}
		os.Exit(0)
	}

	// Join args into a single command string
	command := strings.Join(args, " ")

	// Parse the command
	cmd, err := parser.Parse(command)
	if err != nil {
		printError(err)
		os.Exit(5) // Grammar error
	}

	// Handle session commands specially
	if cmd.Verb == types.VerbSession {
		if err := handleSessionCommand(cmd); err != nil {
			printError(err)
			os.Exit(5)
		}
		return
	}

	// Plan the execution
	plan, err := planner.Plan(cmd)
	if err != nil {
		printError(err)
		os.Exit(5) // Grammar/planning error
	}

	// Output the plan (dry-run mode)
	if *dryRun {
		formatted, err := output.FormatPlan(plan)
		if err != nil {
			printError(fmt.Errorf("failed to format plan: %w", err))
			os.Exit(5)
		}
		fmt.Println(string(formatted))
		return
	}

	// Execute the plan
	executor, err := runtime.NewExecutor(plan)
	if err != nil {
		printError(fmt.Errorf("failed to create executor: %w", err))
		os.Exit(5) // Grammar error
	}

	if err := executor.Execute(plan); err != nil {
		printError(err)
		// Check error type for exit code
		if execErr, ok := err.(*runtime.ExecutionError); ok {
			os.Exit(execErr.Code)
		}
		os.Exit(4) // Network error (default)
	}
}

// printError prints an error with helpful diagnostics.
func printError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)

	// Check if it's a ParseError with suggestions
	if parseErr, ok := err.(*parser.ParseError); ok && parseErr.Suggest != "" {
		fmt.Fprintf(os.Stderr, "Hint: Try using '%s' instead\n", parseErr.Suggest)
	}
}

// printHelp prints the grammar summary.
func printHelp() {
	fmt.Print(grammar.FormatHelp())
}

// explainCommand prints the parsed plan for a command without executing it.
func explainCommand(command string) error {
	cmd, err := parser.Parse(command)
	if err != nil {
		return err
	}

	plan, err := planner.Plan(cmd)
	if err != nil {
		return err
	}

	formatted, err := output.FormatPlan(plan)
	if err != nil {
		return fmt.Errorf("failed to format plan: %w", err)
	}

	fmt.Println(string(formatted))
	return nil
}

// handleSessionCommand handles session management commands.
func handleSessionCommand(cmd *types.Command) error {
	host, err := session.ExtractHost(cmd.Target.URL)
	if err != nil {
		return fmt.Errorf("invalid host: %w", err)
	}

	switch cmd.SessionSubcommand {
	case "show":
		sess, err := session.LoadSession(host)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if sess == nil {
			fmt.Printf("No session found for %s\n", host)
			return nil
		}

		// Check if JSON output requested
		asJSON := false
		for _, clause := range cmd.Clauses {
			if asClause, ok := clause.(types.AsClause); ok && asClause.Format == "json" {
				asJSON = true
				break
			}
		}

		if asJSON {
			// Machine-friendly JSON output
			data, err := json.MarshalIndent(sess, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal session: %w", err)
			}
			fmt.Println(string(data))
		} else {
			// Human-readable redacted output
			redacted := session.RedactSession(sess)
			fmt.Printf("Session for %s:\n", redacted.Host)
			if len(redacted.Cookies) > 0 {
				fmt.Println("Cookies:")
				for name := range redacted.Cookies {
					fmt.Printf("  %s: ***\n", name)
				}
			}
			if redacted.Authorization != "" {
				fmt.Printf("Authorization: %s\n", redacted.Authorization)
			}
		}
		return nil

	case "clear":
		if err := session.DeleteSession(host); err != nil {
			return fmt.Errorf("failed to delete session: %w", err)
		}
		fmt.Printf("Session cleared for %s\n", host)
		return nil

	case "use":
		sess, err := session.LoadSession(host)
		if err != nil {
			return fmt.Errorf("failed to load session: %w", err)
		}
		if sess == nil {
			return fmt.Errorf("no session found for %s", host)
		}
		// Print environment variable stub for shell scoping
		fmt.Printf("export REQ_SESSION_HOST=%s\n", host)
		return nil

	default:
		return fmt.Errorf("unknown session subcommand: %s", cmd.SessionSubcommand)
	}
}
