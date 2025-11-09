// Package parser implements a lexer and parser for the req command grammar.
//
// Grammar (EBNF):
//
//	command = verb target [clauses]
//	verb = "read" | "save" | "send" | "upload" | "watch" | "inspect" | "authenticate" | "session"
//	target = url
//	clauses = clause { clause }
//	clause = with_clause | include_clause | attach_clause | expect_clause | as_clause | to_clause |
//	         using_clause | retry_clause | under_clause | via_clause | follow_clause | insecure_clause
//	with_clause = "with=" ( string | "@file" | "@-" )
//	include_clause = "include=" items
//	attach_clause = "attach=" parts
//	expect_clause = "expect=" checks
//	as_clause = "as=" ( "json" | "csv" | "text" | "raw" )
//	to_clause = "to=" path
//	using_clause = "using=" ( "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "HEAD" | "OPTIONS" )
//	retry_clause = "retry=" number
//	under_clause = "under=" ( duration | size )
//	via_clause = "via=" url
//	follow_clause = "follow=smart"
//	insecure_clause = "insecure=" ( "true" | "false" )
package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/adammpkins/req/internal/types"
)

// isValidHTTPMethod checks if a method is a valid HTTP method.
func isValidHTTPMethod(method string) bool {
	validMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	methodUpper := strings.ToUpper(method)
	for _, valid := range validMethods {
		if methodUpper == valid {
			return true
		}
	}
	return false
}

// ParseError represents a parse error with position information.
type ParseError struct {
	Position int
	Token    string
	Message  string
	Suggest  string
}

func (e *ParseError) Error() string {
	if e.Suggest != "" {
		return fmt.Sprintf("parse error at position %d (token: %q): %s (did you mean %q?)", e.Position, e.Token, e.Message, e.Suggest)
	}
	return fmt.Sprintf("parse error at position %d (token: %q): %s", e.Position, e.Token, e.Message)
}

// Parser parses req commands into AST.
type Parser struct {
	tokens []token
	pos    int
}

// token represents a lexical token.
type token struct {
	typ   tokenType
	value string
	pos   int
}

type tokenType int

const (
	tokenEOF tokenType = iota
	tokenWord
	tokenURL
	tokenEquals
	tokenColon
	tokenDotDot
	tokenString
	tokenNumber
	tokenDuration
	tokenFlag
)

// Parse parses a command string into a Command AST.
func Parse(input string) (*types.Command, error) {
	p := &Parser{}
	p.tokenize(input)
	return p.parseCommand()
}

// tokenize tokenizes the input string.
func (p *Parser) tokenize(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		p.tokens = []token{{typ: tokenEOF, pos: 0}}
		return
	}

	// Split on whitespace while respecting quoted strings
	parts := tokenizeRespectingQuotes(input)
	tokens := make([]token, 0, len(parts))

	for i, part := range parts {
		pos := i
		// Check if this is a URL first (URLs with query params contain = but are not clauses)
		if looksLikeURL(part) {
			tokens = append(tokens, token{typ: tokenURL, value: part, pos: pos})
		} else if strings.Contains(part, "=") {
			// Handle clauses with equals
			// Split on = but keep the = as a token
			eqIdx := strings.Index(part, "=")
			key := part[:eqIdx]
			value := part[eqIdx+1:]

			tokens = append(tokens, token{typ: tokenWord, value: key, pos: pos})
			tokens = append(tokens, token{typ: tokenEquals, value: "=", pos: pos})
			// Handle typed values like json:... (but not if value is quoted or looks like JSON)
			valueTrimmed := strings.TrimSpace(value)
			isQuoted := (len(valueTrimmed) >= 2 && ((valueTrimmed[0] == '\'' && valueTrimmed[len(valueTrimmed)-1] == '\'') || (valueTrimmed[0] == '"' && valueTrimmed[len(valueTrimmed)-1] == '"')))
			looksLikeJSON := strings.HasPrefix(valueTrimmed, "{") || strings.HasPrefix(valueTrimmed, "[")
			// Only parse as typed value if it matches pattern "word:" at the start (like json:...)
			// and not quoted or JSON-like
			if !isQuoted && !looksLikeJSON && strings.Contains(value, ":") {
				colonIdx := strings.Index(value, ":")
				typeName := strings.TrimSpace(value[:colonIdx])
				// Only treat as typed if typeName is a simple word (no special chars)
				if isSimpleWord(typeName) {
					typeValue := value[colonIdx+1:]
					tokens = append(tokens, token{typ: tokenWord, value: typeName, pos: pos})
					tokens = append(tokens, token{typ: tokenColon, value: ":", pos: pos})
					tokens = append(tokens, token{typ: tokenString, value: typeValue, pos: pos})
				} else {
					tokens = append(tokens, token{typ: tokenString, value: value, pos: pos})
				}
			} else if looksLikeURL(value) {
				tokens = append(tokens, token{typ: tokenURL, value: value, pos: pos})
			} else if looksLikeDuration(value) {
				tokens = append(tokens, token{typ: tokenDuration, value: value, pos: pos})
			} else {
				tokens = append(tokens, token{typ: tokenString, value: value, pos: pos})
			}
		} else if isFlag(part) {
			tokens = append(tokens, token{typ: tokenFlag, value: part, pos: pos})
		} else {
			tokens = append(tokens, token{typ: tokenWord, value: part, pos: pos})
		}
	}

	tokens = append(tokens, token{typ: tokenEOF, pos: len(parts)})
	p.tokens = tokens
}

// tokenizeRespectingQuotes splits a string on whitespace while respecting quoted strings.
// Quoted strings (single or double quotes) are preserved as single tokens.
// Also preserves clause values (everything after =) as single tokens even if unquoted.
func tokenizeRespectingQuotes(s string) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)
	escape := false
	afterEquals := false // Track if we're in a clause value (after =)

	for i, r := range s {
		if escape {
			current.WriteRune(r)
			escape = false
			continue
		}

		if r == '\\' {
			escape = true
			current.WriteRune(r)
			continue
		}

		if (r == '\'' || r == '"') && !inQuotes {
			// Start of quoted string
			inQuotes = true
			quoteChar = r
			current.WriteRune(r)
			continue
		}

		if inQuotes && r == quoteChar {
			// End of quoted string
			inQuotes = false
			quoteChar = 0
			current.WriteRune(r)
			continue
		}

		if r == '=' && !inQuotes {
			// Found equals - mark that we're now in a clause value
			afterEquals = true
			current.WriteRune(r)
			continue
		}

		if !inQuotes && (r == ' ' || r == '\t' || r == '\n') {
			// Whitespace outside quotes
			if afterEquals {
				// We're in a clause value - check if next token starts a new clause
				// Look ahead to see if there's a word followed by =
				remaining := s[i+1:]
				remaining = strings.TrimLeft(remaining, " \t\n")
				// Check if remaining starts with a word followed by = (new clause)
				// or if it's empty/EOF
				if remaining == "" || looksLikeNewClause(remaining) {
					// This whitespace separates clauses - split here
					if current.Len() > 0 {
						parts = append(parts, current.String())
						current.Reset()
					}
					afterEquals = false
					continue
				}
				// Otherwise, preserve whitespace as part of the clause value
				current.WriteRune(r)
				continue
			}
			// Not in a clause value - split here
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			afterEquals = false
			continue
		}

		// If we hit a non-whitespace character after being in a clause value,
		// and it's not part of a quoted string, check if it starts a new clause
		if afterEquals && !inQuotes && r != ' ' && r != '\t' && r != '\n' {
			// Check if we're starting a new clause (word followed by =)
			// This is tricky - we need to look at what we've accumulated
			// For now, just continue accumulating
		}

		current.WriteRune(r)
	}

	// Add remaining content
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// looksLikeNewClause checks if a string looks like it starts a new clause (word=)
func looksLikeNewClause(s string) bool {
	// Find first space or equals
	for i, r := range s {
		if r == '=' {
			// Found = - check if there's a word before it
			if i > 0 {
				word := strings.TrimSpace(s[:i])
				// Check if it's a valid clause key
				validKeys := []string{"include", "expect", "with", "as", "to", "using", "retry", "under", "via", "follow", "insecure", "attach"}
				for _, key := range validKeys {
					if word == key {
						return true
					}
				}
			}
			return false
		}
		if r == ' ' || r == '\t' {
			// Found space before = - not a new clause
			return false
		}
	}
	return false
}

// isSimpleWord checks if a string is a simple word (letters, numbers, underscore, no special chars)
func isSimpleWord(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_') {
			return false
		}
	}
	return true
}

// looksLikeURL checks if a string looks like a URL.
func looksLikeURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

// looksLikeDuration checks if a string looks like a duration.
func looksLikeDuration(s string) bool {
	_, err := parseDuration(s)
	return err == nil
}

// isFlag checks if a string is a flag.
func isFlag(s string) bool {
	return s == "verbose" || s == "resume"
}

// parseCommand parses a command.
func (p *Parser) parseCommand() (*types.Command, error) {
	cmd := &types.Command{}

	// Parse verb
	verb, err := p.parseVerb()
	if err != nil {
		return nil, err
	}
	cmd.Verb = verb

	// Handle session subcommands (show, clear, use)
	if verb == types.VerbSession {
		if p.pos >= len(p.tokens) {
			return nil, &ParseError{Position: p.pos, Token: "", Message: "expected session subcommand (show, clear, use)"}
		}
		tok := p.tokens[p.pos]
		if tok.typ == tokenWord {
			subcmd := tok.value
			if subcmd == "show" || subcmd == "clear" || subcmd == "use" {
				cmd.SessionSubcommand = subcmd
				p.pos++
			} else {
				return nil, &ParseError{Position: tok.pos, Token: subcmd, Message: "unknown session subcommand (expected show, clear, or use)"}
			}
		}
	}

	// Parse target
	target, err := p.parseTarget()
	if err != nil {
		return nil, err
	}
	cmd.Target = target

	// Parse clauses
	clauses, err := p.parseClauses()
	if err != nil {
		return nil, err
	}
	cmd.Clauses = clauses

	return cmd, nil
}

// parseVerb parses a verb.
func (p *Parser) parseVerb() (types.Verb, error) {
	if p.pos >= len(p.tokens) {
		return "", &ParseError{Position: p.pos, Token: "", Message: "expected verb"}
	}

	tok := p.tokens[p.pos]
	if tok.typ != tokenWord {
		return "", &ParseError{Position: tok.pos, Token: tok.value, Message: "expected verb"}
	}

	verb := types.Verb(tok.value)
	switch verb {
	case types.VerbRead, types.VerbSave, types.VerbSend, types.VerbUpload,
		types.VerbWatch, types.VerbInspect, types.VerbAuthenticate, types.VerbSession:
		p.pos++
		return verb, nil
	default:
		suggest := suggestVerb(tok.value)
		return "", &ParseError{Position: tok.pos, Token: tok.value, Message: "unknown verb", Suggest: suggest}
	}
}

// suggestVerb suggests a similar verb.
func suggestVerb(input string) string {
	verbs := []string{"read", "save", "send", "upload", "watch", "inspect", "authenticate", "session"}
	best := ""
	minDist := 999
	for _, v := range verbs {
		dist := levenshteinDistance(input, v)
		if dist < minDist {
			minDist = dist
			best = v
		}
	}
	if minDist <= 2 {
		return best
	}
	return ""
}

// levenshteinDistance calculates the Levenshtein distance between two strings.
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
	}

	for i := 0; i <= len(a); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,
				matrix[i][j-1]+1,
				matrix[i-1][j-1]+cost,
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// parseTarget parses a target URL.
func (p *Parser) parseTarget() (types.Target, error) {
	if p.pos >= len(p.tokens) {
		return types.Target{}, &ParseError{Position: p.pos, Token: "", Message: "expected target URL or host"}
	}

	tok := p.tokens[p.pos]
	// For session commands, target might be a host instead of full URL
	if tok.typ == tokenURL {
		p.pos++
		return types.Target{URL: tok.value}, nil
	} else if tok.typ == tokenWord {
		// Might be a host name for session commands
		// Try to parse as URL, if it fails, treat as host
		if strings.Contains(tok.value, ".") || strings.Contains(tok.value, ":") {
			// Looks like a host, construct URL
			urlStr := "https://" + tok.value
			p.pos++
			return types.Target{URL: urlStr}, nil
		}
	}

	return types.Target{}, &ParseError{Position: tok.pos, Token: tok.value, Message: "expected URL or host"}
}

// parseClauses parses zero or more clauses.
func (p *Parser) parseClauses() ([]types.Clause, error) {
	var clauses []types.Clause
	singletonSeen := make(map[string]bool)

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokenEOF {
			break
		}

		clause, err := p.parseClause()
		if err != nil {
			return nil, err
		}

		// Check for duplicate singletons
		if singletonKey := getSingletonKey(clause); singletonKey != "" {
			if singletonSeen[singletonKey] {
				return nil, &ParseError{
					Position: tok.pos,
					Token:    tok.value,
					Message:  fmt.Sprintf("duplicate singleton clause '%s'", singletonKey),
					Suggest:  fmt.Sprintf("remove duplicate '%s=' clause", singletonKey),
				}
			}
			singletonSeen[singletonKey] = true
		}

		clauses = append(clauses, clause)
	}

	return clauses, nil
}

// getSingletonKey returns the key name for singleton clauses, or empty string for repeatable clauses.
func getSingletonKey(clause types.Clause) string {
	switch clause.(type) {
	case types.UsingClause:
		return "using"
	case types.WithClause:
		return "with"
	case types.ExpectClause:
		return "expect"
	case types.AsClause:
		return "as"
	case types.ToClause:
		return "to"
	case types.RetryClause:
		return "retry"
	case types.UnderClause:
		return "under"
	case types.ViaClause:
		return "via"
	case types.InsecureClause:
		return "insecure"
	case types.FollowClause:
		return "follow"
	case types.TimeoutClause:
		return "timeout"
	case types.BackoffClause:
		return "backoff"
	case types.PickClause:
		return "pick"
	case types.EveryClause:
		return "every"
	case types.UntilClause:
		return "until"
	case types.ProxyClause:
		return "proxy"
	case types.FieldClause:
		return "field"
	case types.VerboseClause:
		return "verbose"
	case types.ResumeClause:
		return "resume"
	// Repeatable clauses return empty string
	case types.IncludeClause, types.AttachClause:
		return ""
	default:
		return ""
	}
}

// parseClause parses a single clause.
func (p *Parser) parseClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected clause"}
	}

	tok := p.tokens[p.pos]

	// Handle flags (insecure is now a clause with =, but keep verbose and resume as flags)
	if tok.typ == tokenFlag {
		p.pos++
		switch tok.value {
		case "verbose":
			return types.VerboseClause{}, nil
		case "resume":
			return types.ResumeClause{}, nil
		}
	}

	// Handle clauses with equals
	if tok.typ == tokenWord && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].typ == tokenEquals {
		key := tok.value
		p.pos += 2 // skip key and =

		switch key {
		case "with":
			return p.parseWithClause()
		case "include":
			return p.parseIncludeClause()
		case "attach":
			return p.parseAttachClause()
		case "expect":
			return p.parseExpectClause()
		case "headers":
			return p.parseHeadersClause()
		case "params":
			return p.parseParamsClause()
		case "as":
			return p.parseAsClause()
		case "to":
			return p.parseToClause()
		case "using":
			return p.parseUsingClause()
		case "retry":
			return p.parseRetryClause()
		case "backoff":
			return p.parseBackoffClause()
		case "timeout":
			return p.parseTimeoutClause()
		case "under":
			return p.parseUnderClause()
		case "proxy":
			return p.parseProxyClause()
		case "via":
			return p.parseViaClause()
		case "follow":
			return p.parseFollowClause()
		case "insecure":
			return p.parseInsecureClause()
		case "pick":
			return p.parsePickClause()
		case "every":
			return p.parseEveryClause()
		case "until":
			return p.parseUntilClause()
		case "field":
			return p.parseFieldClause()
		default:
			suggest := suggestClause(key)
			return nil, &ParseError{Position: tok.pos, Token: key, Message: "unknown clause", Suggest: suggest}
		}
	}

	return nil, &ParseError{Position: tok.pos, Token: tok.value, Message: "expected clause"}
}

// suggestClause suggests a similar clause name.
func suggestClause(input string) string {
	clauses := []string{"with", "include", "attach", "expect", "headers", "params", "as", "to", "using", "retry", "backoff", "timeout", "under", "proxy", "via", "follow", "insecure", "pick", "every", "until", "field"}
	best := ""
	minDist := 999
	for _, c := range clauses {
		dist := levenshteinDistance(input, c)
		if dist < minDist {
			minDist = dist
			best = c
		}
	}
	if minDist <= 2 {
		return best
	}
	return ""
}

// parseWithClause parses a "with=" clause.
func (p *Parser) parseWithClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected with value"}
	}

	tok := p.tokens[p.pos]
	if tok.typ == tokenWord && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].typ == tokenColon {
		// typed value like json:...
		typeName := tok.value
		p.pos += 2 // skip type and :
		if p.pos >= len(p.tokens) {
			return nil, &ParseError{Position: p.pos, Token: "", Message: "expected with value"}
		}
		valueTok := p.tokens[p.pos]
		p.pos++
		value := valueTok.value
		isFile := strings.HasPrefix(value, "@") && value != "@-"
		isStdin := value == "@-"
		if isFile {
			value = value[1:] // Remove @ prefix
		}
		return types.WithClause{Type: typeName, Value: value, IsFile: isFile, IsStdin: isStdin}, nil
	}

	// plain value - collect all tokens until next clause
	var valueParts []string
	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokenEOF {
			break
		}
		// Stop if we hit another clause (word followed by =)
		if tok.typ == tokenWord && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].typ == tokenEquals {
			break
		}
		valueParts = append(valueParts, tok.value)
		p.pos++
	}

	if len(valueParts) == 0 {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected with value"}
	}

	// Join tokens with spaces
	value := strings.Join(valueParts, " ")
	
	// Unquote if needed
	value = unquoteString(value)
	
	isFile := strings.HasPrefix(value, "@") && value != "@-"
	isStdin := value == "@-"
	if isFile {
		value = value[1:] // Remove @ prefix
	}
	
	// Infer JSON type if value starts with { or [
	typeInferred := ""
	if !isFile && !isStdin {
		trimmed := strings.TrimSpace(value)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
			typeInferred = "json"
		}
	}
	
	return types.WithClause{Value: value, Type: typeInferred, IsFile: isFile, IsStdin: isStdin}, nil
}

// parseHeadersClause parses a "headers=" clause (simplified for v0.1.0).
func (p *Parser) parseHeadersClause() (types.Clause, error) {
	// Simplified: just parse a single key:value pair for now
	// Full object parsing will come later
	return types.HeadersClause{Headers: make(map[string]string)}, nil
}

// parseParamsClause parses a "params=" clause (simplified for v0.1.0).
func (p *Parser) parseParamsClause() (types.Clause, error) {
	// Simplified: just parse a single key=value pair for now
	return types.ParamsClause{Params: make(map[string]string)}, nil
}

// parseAsClause parses an "as=" clause.
func (p *Parser) parseAsClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected format"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	return types.AsClause{Format: tok.value}, nil
}

// parseToClause parses a "to=" clause.
func (p *Parser) parseToClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected destination"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	return types.ToClause{Destination: tok.value}, nil
}

// parseUsingClause parses a "using=" clause.
func (p *Parser) parseUsingClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected HTTP method"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	method := strings.ToUpper(tok.value)
	
	if !isValidHTTPMethod(method) {
		return nil, &ParseError{
			Position: tok.pos,
			Token:    tok.value,
			Message:  fmt.Sprintf("invalid HTTP method: %s (valid methods: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS)", tok.value),
		}
	}
	
	return types.UsingClause{Method: method}, nil
}

// parseRetryClause parses a "retry=" clause.
func (p *Parser) parseRetryClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected retry count"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	// Parse number (simplified)
	count := 3 // default
	if tok.typ == tokenNumber {
		// In a real implementation, parse the number
		// For now, just use default
	}
	return types.RetryClause{Count: count}, nil
}

// parseBackoffClause parses a "backoff=" clause.
func (p *Parser) parseBackoffClause() (types.Clause, error) {
	// Format: backoff=200ms..5s
	if p.pos+2 >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected backoff range"}
	}

	minTok := p.tokens[p.pos]
	p.pos++
	if p.tokens[p.pos].typ != tokenDotDot {
		return nil, &ParseError{Position: p.pos, Token: p.tokens[p.pos].value, Message: "expected .."}
	}
	p.pos++
	maxTok := p.tokens[p.pos]
	p.pos++

	minDur, err := parseDuration(minTok.value)
	if err != nil {
		return nil, &ParseError{Position: minTok.pos, Token: minTok.value, Message: "invalid duration"}
	}
	maxDur, err := parseDuration(maxTok.value)
	if err != nil {
		return nil, &ParseError{Position: maxTok.pos, Token: maxTok.value, Message: "invalid duration"}
	}

	return types.BackoffClause{Min: minDur, Max: maxDur}, nil
}

// parseDuration parses a duration string.
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// parseTimeoutClause parses a "timeout=" clause.
func (p *Parser) parseTimeoutClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected timeout duration"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	dur, err := parseDuration(tok.value)
	if err != nil {
		return nil, &ParseError{Position: tok.pos, Token: tok.value, Message: "invalid duration"}
	}
	return types.TimeoutClause{Duration: dur}, nil
}

// parseProxyClause parses a "proxy=" clause.
func (p *Parser) parseProxyClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected proxy URL"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	return types.ProxyClause{URL: tok.value}, nil
}

// parsePickClause parses a "pick=" clause.
func (p *Parser) parsePickClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected JSONPath expression"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	return types.PickClause{Path: tok.value}, nil
}

// parseEveryClause parses an "every=" clause.
func (p *Parser) parseEveryClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected interval duration"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	dur, err := parseDuration(tok.value)
	if err != nil {
		return nil, &ParseError{Position: tok.pos, Token: tok.value, Message: "invalid duration"}
	}
	return types.EveryClause{Interval: dur}, nil
}

// parseUntilClause parses an "until=" clause.
func (p *Parser) parseUntilClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected predicate"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	return types.UntilClause{Predicate: tok.value}, nil
}

// parseFieldClause parses a "field=" clause.
func (p *Parser) parseFieldClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected field name"}
	}

	nameTok := p.tokens[p.pos]
	p.pos++
	if p.pos >= len(p.tokens) || p.tokens[p.pos].typ != tokenEquals {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected ="}
	}
	p.pos++
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected field value"}
	}
	valueTok := p.tokens[p.pos]
	p.pos++

	return types.FieldClause{Name: nameTok.value, Value: valueTok.value}, nil
}

// parseIncludeClause parses an "include=" clause.
// Format: include='header: Name: Value; param: key=value; cookie: key=value'
func (p *Parser) parseIncludeClause() (types.Clause, error) {
	// Collect tokens until we have a complete include value
	// The value may contain colons, semicolons, and spaces
	var valueParts []string
	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokenEOF {
			break
		}
		// Stop if we hit another clause (word followed by =)
		if tok.typ == tokenWord && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].typ == tokenEquals {
			break
		}
		valueParts = append(valueParts, tok.value)
		p.pos++
	}

	if len(valueParts) == 0 {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected include value"}
	}

	// Join tokens, but handle colons and semicolons specially (no space before/after)
	value := ""
	for i, part := range valueParts {
		if i > 0 {
			// Add space only if previous token wasn't a colon/semicolon and current isn't a colon/semicolon
			prev := valueParts[i-1]
			if prev != ":" && prev != ";" && part != ":" && part != ";" {
				value += " "
			}
		}
		value += part
	}

	// Unquote the value if it's a single quoted string
	value = unquoteString(value)

	items, err := parseIncludeItems(value)
	if err != nil {
		return nil, &ParseError{Position: p.pos, Token: value, Message: err.Error()}
	}

	return types.IncludeClause{Items: items}, nil
}

// parseIncludeItems parses semicolon-separated include items.
func parseIncludeItems(value string) ([]types.IncludeItem, error) {
	var items []types.IncludeItem
	
	// For header items, semicolons in the value should not split items
	// We need to parse more carefully by looking for type tags
	value = strings.TrimSpace(value)
	if value == "" {
		return items, nil
	}
	
	// Split by semicolons, but be smarter about it:
	// - Look for type tags (header:, param:, cookie:)
	// - For header items, everything after "Name: " is the value, including semicolons
	// - Only split on semicolons that are followed by a type tag
	
	var current strings.Builder
	var parts []string
	
	i := 0
	for i < len(value) {
		// Look ahead for a type tag pattern: "; header:", "; param:", "; cookie:"
		if i > 0 && value[i] == ';' {
			// Check if this semicolon is followed by a type tag
			remaining := strings.TrimSpace(value[i+1:])
			if strings.HasPrefix(remaining, "header:") ||
				strings.HasPrefix(remaining, "param:") ||
				strings.HasPrefix(remaining, "cookie:") {
				// This semicolon is a separator
				if current.Len() > 0 {
					parts = append(parts, strings.TrimSpace(current.String()))
					current.Reset()
				}
				i++ // Skip the semicolon
				continue
			}
		}
		current.WriteByte(value[i])
		i++
	}
	
	if current.Len() > 0 {
		parts = append(parts, strings.TrimSpace(current.String()))
	}
	
	// If we didn't find any type tag separators, treat the whole thing as one item
	if len(parts) == 0 {
		parts = []string{value}
	}
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		item, err := parseIncludeItem(part)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	
	return items, nil
}

// parseIncludeItem parses a single include item (header:, param:, cookie:).
func parseIncludeItem(part string) (types.IncludeItem, error) {
	// Find the first colon to determine the type
	colonIdx := strings.Index(part, ":")
	if colonIdx == -1 {
		return types.IncludeItem{}, fmt.Errorf("missing colon in include item: %s", part)
	}
	
	typeTag := strings.TrimSpace(part[:colonIdx])
	rest := strings.TrimSpace(part[colonIdx+1:])
	
	switch typeTag {
	case "header":
		// Format: header: Name: Value
		headerColonIdx := strings.Index(rest, ":")
		if headerColonIdx == -1 {
			return types.IncludeItem{}, fmt.Errorf("header item missing Name colon Value: %s", part)
		}
		name := strings.TrimSpace(rest[:headerColonIdx])
		value := strings.TrimSpace(rest[headerColonIdx+1:])
		// Unquote if needed
		name = unquoteString(name)
		value = unquoteString(value)
		return types.IncludeItem{Type: "header", Name: name, Value: value}, nil
		
	case "param":
		// Format: param: key=value
		eqIdx := strings.Index(rest, "=")
		if eqIdx == -1 {
			return types.IncludeItem{}, fmt.Errorf("param item missing equals: %s", part)
		}
		key := strings.TrimSpace(rest[:eqIdx])
		value := strings.TrimSpace(rest[eqIdx+1:])
		key = unquoteString(key)
		value = unquoteString(value)
		return types.IncludeItem{Type: "param", Name: key, Value: value}, nil
		
	case "cookie":
		// Format: cookie: key=value
		eqIdx := strings.Index(rest, "=")
		if eqIdx == -1 {
			return types.IncludeItem{}, fmt.Errorf("cookie item missing equals: %s", part)
		}
		key := strings.TrimSpace(rest[:eqIdx])
		value := strings.TrimSpace(rest[eqIdx+1:])
		key = unquoteString(key)
		value = unquoteString(value)
		return types.IncludeItem{Type: "cookie", Name: key, Value: value}, nil
		
	case "basic":
		// Format: basic: username:password
		// The rest should be username:password
		rest = strings.TrimSpace(rest)
		rest = unquoteString(rest)
		// Validate that it contains at least one colon
		colonIdx := strings.Index(rest, ":")
		if colonIdx == -1 {
			return types.IncludeItem{}, fmt.Errorf("basic item must be in format username:password: %s", part)
		}
		// Store the full username:password as the value
		// We'll split it in the planner when encoding
		return types.IncludeItem{Type: "basic", Value: rest}, nil
		
	default:
		return types.IncludeItem{}, fmt.Errorf("unknown include item tag: %s (expected header, param, cookie, or basic)", typeTag)
	}
}

// splitRespectingQuotes splits a string by a delimiter while respecting quoted strings.
func splitRespectingQuotes(s string, delim rune) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	escape := false
	
	for _, r := range s {
		if escape {
			current.WriteRune(r)
			escape = false
			continue
		}
		
		if r == '\\' {
			escape = true
			current.WriteRune(r)
			continue
		}
		
		if r == '\'' || r == '"' {
			inQuotes = !inQuotes
			current.WriteRune(r)
			continue
		}
		
		if r == delim && !inQuotes {
			parts = append(parts, current.String())
			current.Reset()
			continue
		}
		
		current.WriteRune(r)
	}
	
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	
	return parts
}

// unquoteString removes surrounding quotes if present and handles escapes.
func unquoteString(s string) string {
	if len(s) >= 2 && ((s[0] == '\'' && s[len(s)-1] == '\'') || (s[0] == '"' && s[len(s)-1] == '"')) {
		s = s[1 : len(s)-1]
		// Handle escapes
		var result strings.Builder
		escape := false
		for _, r := range s {
			if escape {
				if r == '\\' || r == '\'' || r == '"' {
					result.WriteRune(r)
				} else {
					result.WriteRune('\\')
					result.WriteRune(r)
				}
				escape = false
			} else if r == '\\' {
				escape = true
			} else {
				result.WriteRune(r)
			}
		}
		if escape {
			result.WriteRune('\\')
		}
		return result.String()
	}
	return s
}

// parseAttachClause parses an "attach=" clause.
// Format: attach='part: name=..., file=@path; part: name=..., value=...'
func (p *Parser) parseAttachClause() (types.Clause, error) {
	// Collect tokens until we have a complete attach value
	// The value may contain spaces, commas, and semicolons
	var valueParts []string
	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokenEOF {
			break
		}
		// Stop if we hit another clause (word followed by =)
		if tok.typ == tokenWord && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].typ == tokenEquals {
			break
		}
		valueParts = append(valueParts, tok.value)
		p.pos++
	}

	if len(valueParts) == 0 {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected attach value"}
	}

	// Join tokens, but handle colons and semicolons specially (no space before/after)
	value := ""
	for i, part := range valueParts {
		if i > 0 {
			// Add space only if previous token wasn't a colon/semicolon and current isn't a colon/semicolon
			prev := valueParts[i-1]
			if prev != ":" && prev != ";" && part != ":" && part != ";" {
				value += " "
			}
		}
		value += part
	}
	
	// Unquote if needed
	value = unquoteString(value)
	
	parts, boundary, err := parseAttachItems(value)
	if err != nil {
		return nil, &ParseError{Position: p.pos, Token: value, Message: err.Error()}
	}

	return types.AttachClause{Parts: parts, Boundary: boundary}, nil
}

// parseAttachItems parses semicolon-separated attach items.
func parseAttachItems(value string) ([]types.AttachPart, string, error) {
	var parts []types.AttachPart
	var boundary string
	
	// Split by semicolons, respecting quotes
	items := splitRespectingQuotes(value, ';')
	
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		
		// Check if it's a boundary specification
		if strings.HasPrefix(item, "boundary:") {
			boundary = strings.TrimSpace(strings.TrimPrefix(item, "boundary:"))
			boundary = unquoteString(boundary)
			continue
		}
		
		// Parse part: specification
		if !strings.HasPrefix(item, "part:") {
			return nil, "", fmt.Errorf("expected 'part:' or 'boundary:', got: %s", item)
		}
		
		partSpec := strings.TrimSpace(strings.TrimPrefix(item, "part:"))
		part, err := parseAttachPart(partSpec)
		if err != nil {
			return nil, "", err
		}
		parts = append(parts, part)
	}
	
	return parts, boundary, nil
}

// parseAttachPart parses a single attach part specification.
// Format: name=..., file=@path or value=..., optional filename=..., optional type=...
func parseAttachPart(spec string) (types.AttachPart, error) {
	var part types.AttachPart
	
	// Parse comma-separated key=value pairs
	pairs := splitRespectingQuotes(spec, ',')
	
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		eqIdx := strings.Index(pair, "=")
		if eqIdx == -1 {
			return types.AttachPart{}, fmt.Errorf("missing equals in attach part: %s", pair)
		}
		
		key := strings.TrimSpace(pair[:eqIdx])
		value := strings.TrimSpace(pair[eqIdx+1:])
		value = unquoteString(value)
		
		switch key {
		case "name":
			part.Name = value
		case "file":
			if strings.HasPrefix(value, "@") {
				part.FilePath = value[1:] // Remove @
			} else {
				part.FilePath = value
			}
		case "value":
			part.Value = value
		case "filename":
			part.Filename = value
		case "type":
			part.Type = value
		default:
			return types.AttachPart{}, fmt.Errorf("unknown attach part key: %s", key)
		}
	}
	
	// Validate: name is required
	if part.Name == "" {
		return types.AttachPart{}, fmt.Errorf("attach part missing required 'name='")
	}
	
	// Validate: exactly one of file or value
	hasFile := part.FilePath != ""
	hasValue := part.Value != ""
	if hasFile && hasValue {
		return types.AttachPart{}, fmt.Errorf("attach part cannot have both 'file=' and 'value='")
	}
	if !hasFile && !hasValue {
		return types.AttachPart{}, fmt.Errorf("attach part must have either 'file=' or 'value='")
	}
	
	return part, nil
}

// parseExpectClause parses an "expect=" clause.
// Format: expect=status:200, header:Content-Type=application/json, contains:"text"
func (p *Parser) parseExpectClause() (types.Clause, error) {
	// Collect tokens until we have a complete expect value
	// The value may contain colons and commas, so we need to collect multiple tokens
	var valueParts []string
	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokenEOF {
			break
		}
		// Stop if we hit another clause (word followed by =)
		if tok.typ == tokenWord && p.pos+1 < len(p.tokens) && p.tokens[p.pos+1].typ == tokenEquals {
			break
		}
		valueParts = append(valueParts, tok.value)
		p.pos++
	}

	if len(valueParts) == 0 {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected expect value"}
	}

	// Join tokens, but handle colons specially (no space before/after)
	value := ""
	for i, part := range valueParts {
		if i > 0 {
			// Add space only if previous token wasn't a colon and current isn't a colon
			if valueParts[i-1] != ":" && part != ":" {
				value += " "
			}
		}
		value += part
	}
	
	// Unquote the value if it's a single quoted string
	value = unquoteString(value)
	
	checks, err := parseExpectChecks(value)
	if err != nil {
		return nil, &ParseError{Position: p.pos, Token: value, Message: err.Error()}
	}

	return types.ExpectClause{Checks: checks}, nil
}

// parseExpectChecks parses comma-separated expect checks.
func parseExpectChecks(value string) ([]types.ExpectCheck, error) {
	var checks []types.ExpectCheck
	
	// Split by commas, respecting quotes
	parts := splitRespectingQuotes(value, ',')
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		
		check, err := parseExpectCheck(part)
		if err != nil {
			return nil, err
		}
		checks = append(checks, check)
	}
	
	return checks, nil
}

// parseExpectCheck parses a single expect check.
func parseExpectCheck(part string) (types.ExpectCheck, error) {
	// Unquote if needed first
	unquoted := unquoteString(part)
	
	// Check types: status:, header:, contains:, jsonpath:, matches:
	if strings.HasPrefix(unquoted, "status:") {
		value := strings.TrimSpace(strings.TrimPrefix(unquoted, "status:"))
		return types.ExpectCheck{Type: "status", Value: value}, nil
	}
	
	if strings.HasPrefix(unquoted, "header:") {
		rest := strings.TrimSpace(strings.TrimPrefix(unquoted, "header:"))
		eqIdx := strings.Index(rest, "=")
		if eqIdx == -1 {
			return types.ExpectCheck{}, fmt.Errorf("header check missing equals: %s", part)
		}
		name := strings.TrimSpace(rest[:eqIdx])
		value := strings.TrimSpace(rest[eqIdx+1:])
		value = unquoteString(value)
		return types.ExpectCheck{Type: "header", Name: name, Value: value}, nil
	}
	
	if strings.HasPrefix(unquoted, "contains:") {
		value := strings.TrimSpace(strings.TrimPrefix(unquoted, "contains:"))
		value = unquoteString(value)
		return types.ExpectCheck{Type: "contains", Value: value}, nil
	}
	
	if strings.HasPrefix(unquoted, "jsonpath:") {
		value := strings.TrimSpace(strings.TrimPrefix(unquoted, "jsonpath:"))
		value = unquoteString(value)
		return types.ExpectCheck{Type: "jsonpath", Path: value}, nil
	}
	
	if strings.HasPrefix(unquoted, "matches:") {
		value := strings.TrimSpace(strings.TrimPrefix(unquoted, "matches:"))
		value = unquoteString(value)
		return types.ExpectCheck{Type: "matches", Regex: value}, nil
	}
	
	return types.ExpectCheck{}, fmt.Errorf("unknown expect check type: %s", part)
}

// parseFollowClause parses a "follow=" clause.
func (p *Parser) parseFollowClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected follow value"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	
	value := strings.ToLower(strings.TrimSpace(tok.value))
	if value != "smart" {
		return nil, &ParseError{Position: tok.pos, Token: tok.value, Message: "follow accepts only 'smart'"}
	}
	
	return types.FollowClause{Policy: "smart"}, nil
}

// parseUnderClause parses an "under=" clause (duration or size).
func (p *Parser) parseUnderClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected under value"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	
	value := strings.TrimSpace(tok.value)
	
	// Try parsing as duration first
	if dur, err := parseDuration(value); err == nil {
		return types.UnderClause{Duration: dur, IsSize: false}, nil
	}
	
	// Try parsing as size (e.g., "10MB", "1GB")
	if size, err := parseSize(value); err == nil {
		return types.UnderClause{Size: size, IsSize: true}, nil
	}
	
	return nil, &ParseError{Position: tok.pos, Token: tok.value, Message: "under value must be a duration (e.g., 30s) or size (e.g., 10MB)"}
}

// parseSize parses a size string like "10MB", "1GB", etc.
func parseSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)
	
	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
		"TB": 1024 * 1024 * 1024 * 1024,
	}
	
	for suffix, mult := range multipliers {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			var num float64
			if _, err := fmt.Sscanf(numStr, "%f", &num); err != nil {
				return 0, err
			}
			return int64(num * float64(mult)), nil
		}
	}
	
	return 0, fmt.Errorf("unknown size suffix")
}

// parseViaClause parses a "via=" clause.
func (p *Parser) parseViaClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected via URL"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	return types.ViaClause{URL: tok.value}, nil
}

// parseInsecureClause parses an "insecure=" clause.
func (p *Parser) parseInsecureClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected insecure value"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	
	value := strings.ToLower(strings.TrimSpace(tok.value))
	if value != "true" && value != "false" {
		return nil, &ParseError{Position: tok.pos, Token: tok.value, Message: "insecure accepts only 'true' or 'false'"}
	}
	
	return types.InsecureClause{Value: value == "true"}, nil
}
