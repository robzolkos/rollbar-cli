package output

import (
	"io"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

// Format represents the output format type
type Format string

const (
	FormatTable    Format = "table"
	FormatJSON     Format = "json"
	FormatCompact  Format = "compact"
	FormatMarkdown Format = "markdown"
)

// Formatter is the interface for output formatters
type Formatter interface {
	FormatItems(w io.Writer, items []api.Item) error
	FormatItem(w io.Writer, item *api.Item) error
	FormatInstances(w io.Writer, instances []api.Instance) error
	FormatInstance(w io.Writer, instance *api.Instance) error
	FormatContext(w io.Writer, item *api.Item, instances []api.Instance) error
	FormatProjectInfo(w io.Writer, info *api.ProjectInfo) error
}

// New creates a new formatter based on the format type
func New(format Format, color bool) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{}
	case FormatCompact:
		return &CompactFormatter{}
	case FormatMarkdown:
		return &MarkdownFormatter{}
	case FormatTable:
		fallthrough
	default:
		return &TableFormatter{Color: color}
	}
}
