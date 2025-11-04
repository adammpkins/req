// Package parser implements a lexer and parser for the req command grammar.
//
// Grammar (EBNF):
//
//	command = verb target [clauses]
//	verb = "read" | "save" | "send" | "upload" | "watch" | "inspect" | "auth" | "session" | "profile"
//	target = url
//	clauses = clause { clause }
//	clause = with_clause | headers_clause | params_clause | as_clause | to_clause |
//	         method_clause | retry_clause | backoff_clause | timeout_clause | proxy_clause |
//	         pick_clause | every_clause | until_clause | field_clause | flag_clause
//	with_clause = "with=" ( "json:" string | "form:" string | string )
//	headers_clause = "headers=" object
//	params_clause = "params=" object
//	as_clause = "as=" ( "json" | "csv" | "text" | "raw" )
//	to_clause = "to=" path
//	method_clause = "method=" ( "GET" | "POST" | "PUT" | "DELETE" | ... )
//	retry_clause = "retry=" number
//	backoff_clause = "backoff=" duration ".." duration
//	timeout_clause = "timeout=" duration
//	proxy_clause = "proxy=" url
//	pick_clause = "pick=" jsonpath
//	every_clause = "every=" duration
//	until_clause = "until=" predicate
//	field_clause = "field=" name "=" value
//	flag_clause = "insecure" | "verbose" | "resume"
package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/adammpkins/req/internal/types"
)

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

	parts := strings.Fields(input)
	tokens := make([]token, 0, len(parts))

	for i, part := range parts {
		pos := i
		// Handle clauses with equals
		if strings.Contains(part, "=") {
			// Split on = but keep the = as a token
			eqIdx := strings.Index(part, "=")
			key := part[:eqIdx]
			value := part[eqIdx+1:]

			tokens = append(tokens, token{typ: tokenWord, value: key, pos: pos})
			tokens = append(tokens, token{typ: tokenEquals, value: "=", pos: pos})
			// Handle typed values like json:...
			if strings.Contains(value, ":") {
				colonIdx := strings.Index(value, ":")
				typeName := value[:colonIdx]
				typeValue := value[colonIdx+1:]
				tokens = append(tokens, token{typ: tokenWord, value: typeName, pos: pos})
				tokens = append(tokens, token{typ: tokenColon, value: ":", pos: pos})
				tokens = append(tokens, token{typ: tokenString, value: typeValue, pos: pos})
			} else if looksLikeURL(value) {
				tokens = append(tokens, token{typ: tokenURL, value: value, pos: pos})
			} else if looksLikeDuration(value) {
				tokens = append(tokens, token{typ: tokenDuration, value: value, pos: pos})
			} else {
				tokens = append(tokens, token{typ: tokenString, value: value, pos: pos})
			}
		} else if looksLikeURL(part) {
			tokens = append(tokens, token{typ: tokenURL, value: part, pos: pos})
		} else if isFlag(part) {
			tokens = append(tokens, token{typ: tokenFlag, value: part, pos: pos})
		} else {
			tokens = append(tokens, token{typ: tokenWord, value: part, pos: pos})
		}
	}

	tokens = append(tokens, token{typ: tokenEOF, pos: len(parts)})
	p.tokens = tokens
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
	return s == "insecure" || s == "verbose" || s == "resume"
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
		types.VerbWatch, types.VerbInspect, types.VerbAuth, types.VerbSession, types.VerbProfile:
		p.pos++
		return verb, nil
	default:
		suggest := suggestVerb(tok.value)
		return "", &ParseError{Position: tok.pos, Token: tok.value, Message: "unknown verb", Suggest: suggest}
	}
}

// suggestVerb suggests a similar verb.
func suggestVerb(input string) string {
	verbs := []string{"read", "save", "send", "upload", "watch", "inspect", "auth", "session", "profile"}
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
		return types.Target{}, &ParseError{Position: p.pos, Token: "", Message: "expected target URL"}
	}

	tok := p.tokens[p.pos]
	if tok.typ != tokenURL {
		return types.Target{}, &ParseError{Position: tok.pos, Token: tok.value, Message: "expected URL"}
	}

	p.pos++
	return types.Target{URL: tok.value}, nil
}

// parseClauses parses zero or more clauses.
func (p *Parser) parseClauses() ([]types.Clause, error) {
	var clauses []types.Clause

	for p.pos < len(p.tokens) {
		tok := p.tokens[p.pos]
		if tok.typ == tokenEOF {
			break
		}

		clause, err := p.parseClause()
		if err != nil {
			return nil, err
		}
		clauses = append(clauses, clause)
	}

	return clauses, nil
}

// parseClause parses a single clause.
func (p *Parser) parseClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected clause"}
	}

	tok := p.tokens[p.pos]

	// Handle flags
	if tok.typ == tokenFlag {
		p.pos++
		switch tok.value {
		case "insecure":
			return types.InsecureClause{}, nil
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
		case "headers":
			return p.parseHeadersClause()
		case "params":
			return p.parseParamsClause()
		case "as":
			return p.parseAsClause()
		case "to":
			return p.parseToClause()
		case "method":
			return p.parseMethodClause()
		case "retry":
			return p.parseRetryClause()
		case "backoff":
			return p.parseBackoffClause()
		case "timeout":
			return p.parseTimeoutClause()
		case "proxy":
			return p.parseProxyClause()
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
	clauses := []string{"with", "headers", "params", "as", "to", "method", "retry", "backoff", "timeout", "proxy", "pick", "every", "until", "field"}
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
		return types.WithClause{Type: typeName, Value: valueTok.value}, nil
	}

	// plain value
	valueTok := p.tokens[p.pos]
	p.pos++
	return types.WithClause{Value: valueTok.value}, nil
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

// parseMethodClause parses a "method=" clause.
func (p *Parser) parseMethodClause() (types.Clause, error) {
	if p.pos >= len(p.tokens) {
		return nil, &ParseError{Position: p.pos, Token: "", Message: "expected HTTP method"}
	}

	tok := p.tokens[p.pos]
	p.pos++
	return types.MethodClause{Method: tok.value}, nil
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
