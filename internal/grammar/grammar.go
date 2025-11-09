// Package grammar defines the structured grammar data for req commands.
package grammar

import "fmt"

// Verb represents a req command verb.
type Verb struct {
	Name        string
	Description string
}

// Clause represents a req command clause.
type Clause struct {
	Name        string
	Description string
	Repeatable  bool
	Example     string
}

// Grammar contains the complete grammar definition.
type Grammar struct {
	Verbs   []Verb
	Clauses []Clause
}

// GetGrammar returns the canonical grammar definition.
func GetGrammar() Grammar {
	return Grammar{
		Verbs: []Verb{
			{Name: "read", Description: "GET, print to stdout"},
			{Name: "save", Description: "GET, write to file via to="},
			{Name: "send", Description: "default GET, POST if with= present"},
			{Name: "upload", Description: "POST when attach= or with= present, else error"},
			{Name: "watch", Description: "GET with SSE or polling"},
			{Name: "inspect", Description: "HEAD only"},
			{Name: "authenticate", Description: "login and store session state"},
			{Name: "session", Description: "session management (show, clear, use)"},
		},
		Clauses: []Clause{
			{Name: "using=", Description: "HTTP method override", Repeatable: false, Example: "using=PUT"},
			{Name: "include=", Description: "Add headers, params, cookies, basic auth", Repeatable: true, Example: "include='header: Authorization: Bearer token; param: q=search query; basic: user:pass'"},
			{Name: "with=", Description: "Request body", Repeatable: false, Example: "with=@user.json or with='{\"name\":\"Adam\"}'"},
			{Name: "expect=", Description: "Assertions on response", Repeatable: false, Example: "expect=status:200, header:Content-Type=application/json, contains:\"ok\""},
			{Name: "as=", Description: "Output format for stdout", Repeatable: false, Example: "as=json"},
			{Name: "to=", Description: "Destination path", Repeatable: false, Example: "to=out.json"},
			{Name: "retry=", Description: "Retry attempts for transient errors", Repeatable: false, Example: "retry=3"},
			{Name: "under=", Description: "Timeout or size limit", Repeatable: false, Example: "under=30s or under=10MB"},
			{Name: "via=", Description: "Proxy URL", Repeatable: false, Example: "via=http://proxy:8080"},
			{Name: "attach=", Description: "Multipart parts for upload or send", Repeatable: true, Example: "attach='part: name=avatar, file=@me.png; part: name=meta, value=xyz'"},
			{Name: "follow=", Description: "Redirect policy for write verbs", Repeatable: false, Example: "follow=smart"},
			{Name: "insecure=", Description: "Disable TLS verification for this request", Repeatable: false, Example: "insecure=true"},
		},
	}
}

// FormatHelp formats the grammar as help text.
func FormatHelp() string {
	g := GetGrammar()
	
	var help string
	help += "req - HTTP client DSL\n\n"
	help += "Usage: req <verb> <url> [clauses...]\n\n"
	help += "Verbs:\n"
	
	for _, verb := range g.Verbs {
		help += fmt.Sprintf("  %-13s - %s\n", verb.Name, verb.Description)
	}
	
	help += "\nClauses:\n"
	for _, clause := range g.Clauses {
		help += fmt.Sprintf("  %-13s - %s", clause.Name, clause.Description)
		if clause.Repeatable {
			help += " (repeatable)"
		}
		help += "\n"
		if clause.Example != "" {
			help += fmt.Sprintf("                 Example: %s\n", clause.Example)
		}
	}
	
	help += "\nExamples:\n"
	help += "  req read https://api.example.com/search include='param: q=search query' as=json\n"
	help += "  \n"
	help += "  req read https://httpbin.org/basic-auth/user/passwd include='basic: user:passwd' expect=status:200\n"
	help += "  \n"
	help += "  req send https://api.example.com/users \\\n"
	help += "    using=POST \\\n"
	help += "    include='header: Authorization: Bearer $TOKEN' \\\n"
	help += "    with='{\"name\":\"Adam\"}' \\\n"
	help += "    expect=status:201, header:Content-Type=application/json \\\n"
	help += "    as=json\n"
	help += "  \n"
	help += "  req upload https://api.example.com/upload \\\n"
	help += "    attach='part: name=file, file=@./avatar.png, type=image/png' \\\n"
	help += "    as=json\n"
	help += "  \n"
	help += "  req authenticate https://api.example.com/login \\\n"
	help += "    using=POST \\\n"
	help += "    with='{\"user\":\"adam\",\"pass\":\"xyz\"}'\n"
	help += "  \n"
	help += "  req read https://api.example.com/me as=json\n\n"
	help += "For more information, see the grammar documentation.\n"
	
	return help
}

