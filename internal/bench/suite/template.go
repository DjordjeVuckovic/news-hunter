package suite

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type QueryTemplate struct {
	ID    string `yaml:"id"`
	Query string `yaml:"query"`
}

type TemplateParams map[string]any

var placeholderRegex = regexp.MustCompile(`\{\{(\w+)\}\}`)

func (t *QueryTemplate) Render(params TemplateParams, suiteDir string) (*ResolvedQuery, error) {
	result := placeholderRegex.ReplaceAllStringFunc(t.Query, func(match string) string {
		key := match[2 : len(match)-2]
		if val, ok := params[key]; ok {
			return formatValue(val)
		}
		return match
	})

	missing := findMissingPlaceholders(result)
	if len(missing) > 0 {
		return nil, fmt.Errorf("template %q missing params: %v", t.ID, missing)
	}

	return &ResolvedQuery{Query: result}, nil
}

func (t *QueryTemplate) RequiredParams() []string {
	seen := make(map[string]bool)
	var params []string

	matches := placeholderRegex.FindAllStringSubmatch(t.Query, -1)
	for _, m := range matches {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			params = append(params, m[1])
		}
	}

	return params
}

func (t *QueryTemplate) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("template has no id")
	}
	if t.Query == "" {
		return fmt.Errorf("template %q has no query", t.ID)
	}
	return nil
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case []string:
		return strings.Join(val, ", ")
	case []any:
		strs := make([]string, len(val))
		for i, item := range val {
			strs[i] = formatValue(item)
		}
		return strings.Join(strs, ", ")
	default:
		return fmt.Sprintf("%v", v)
	}
}

func findMissingPlaceholders(s string) []string {
	matches := placeholderRegex.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var missing []string
	for _, m := range matches {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			missing = append(missing, m[1])
		}
	}
	return missing
}

type TemplateRegistry struct {
	templates map[string]*QueryTemplate
}

func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]*QueryTemplate),
	}
}

func (r *TemplateRegistry) Register(t *QueryTemplate) error {
	if err := t.Validate(); err != nil {
		return err
	}
	if _, exists := r.templates[t.ID]; exists {
		return fmt.Errorf("template %q already registered", t.ID)
	}
	r.templates[t.ID] = t
	return nil
}

func (r *TemplateRegistry) Get(id string) (*QueryTemplate, bool) {
	t, ok := r.templates[id]
	return t, ok
}

func (r *TemplateRegistry) RenderQuery(templateID string, params TemplateParams, suiteDir string) (*ResolvedQuery, error) {
	t, ok := r.Get(templateID)
	if !ok {
		return nil, fmt.Errorf("template %q not found", templateID)
	}
	return t.Render(params, suiteDir)
}

func (r *TemplateRegistry) List() []string {
	ids := make([]string, 0, len(r.templates))
	for id := range r.templates {
		ids = append(ids, id)
	}
	return ids
}
