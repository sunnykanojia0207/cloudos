package intent

import (
	"fmt"
	"regexp"
	"strings"
)

// parsePattern defines a pattern that maps user input to an intent type
type parsePattern struct {
	regex   *regexp.Regexp
	intent  IntentType
	extract func(matches []string) map[string]string
}

// Parser parses user input strings into structured Intents using rule-based matching
type Parser struct {
	patterns []parsePattern
}

// NewParser creates a parser with built-in patterns
func NewParser() *Parser {
	p := &Parser{}
	p.patterns = []parsePattern{
		{
			// "create project X", "create a project X", "create new project X", "make project X"
			regex:  regexp.MustCompile(`(?i)^create\s+(?:a\s+|an\s+|new\s+)?project\s+(.+)$`),
			intent: IntentCreateProject,
			extract: func(m []string) map[string]string {
				name := strings.TrimSpace(m[1])
				id := toID(name)
				return map[string]string{"name": name, "id": id}
			},
		},
		{
			// "list projects", "show projects", "list all projects", "show all projects"
			regex:  regexp.MustCompile(`(?i)^(?:list|show)\s+(?:all\s+)?projects$`),
			intent: IntentListProjects,
			extract: func(m []string) map[string]string {
				return nil
			},
		},
		{
			// "delete project X", "remove project X"
			regex:  regexp.MustCompile(`(?i)^(?:delete|remove)\s+project\s+(.+)$`),
			intent: IntentDeleteProject,
			extract: func(m []string) map[string]string {
				name := strings.TrimSpace(m[1])
				id := toID(name)
				return map[string]string{"name": name, "id": id}
			},
		},
		{
			// "show controllers", "list controllers", "list all controllers"
			regex:  regexp.MustCompile(`(?i)^(?:show|list)\s+(?:all\s+)?controllers$`),
			intent: IntentShowControllers,
			extract: func(m []string) map[string]string {
				return nil
			},
		},
		{
			// "show resources", "list resources", "list all resources", "show resource kinds"
			regex:  regexp.MustCompile(`(?i)^(?:show|list)\s+(?:all\s+)?(?:resources|resource\s+kinds)$`),
			intent: IntentShowResources,
			extract: func(m []string) map[string]string {
				return nil
			},
		},
		{
			// "show health", "system health", "check health"
			regex:  regexp.MustCompile(`(?i)^(?:show|check|system)\s+health$`),
			intent: IntentShowHealth,
			extract: func(m []string) map[string]string {
				return nil
			},
		},
	}
	return p
}

// Parse converts a user input string into a structured Intent
func (p *Parser) Parse(input string) (*Intent, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty input: please type a request")
	}

	for _, pattern := range p.patterns {
		matches := pattern.regex.FindStringSubmatch(input)
		if matches != nil {
			params := pattern.extract(matches)
			return &Intent{
				Type:   pattern.intent,
				Raw:    input,
				Status: IntentPending,
				Params: params,
			}, nil
		}
	}

	return nil, fmt.Errorf("unrecognized intent: %q — try something like \"create project my-app\", \"list projects\", \"show health\"", input)
}

// toID converts a display name to a DNS-safe identifier
func toID(name string) string {
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")
	// Remove any characters that aren't alphanumeric or hyphen
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	id = reg.ReplaceAllString(id, "")
	return id
}
