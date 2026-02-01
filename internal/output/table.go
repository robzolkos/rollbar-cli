package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

// ANSI color codes
const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
	colorBold    = "\033[1m"
)

// TableFormatter outputs data as human-readable tables
type TableFormatter struct {
	Color bool
}

func (f *TableFormatter) color(code, text string) string {
	if !f.Color {
		return text
	}
	return code + text + colorReset
}

func (f *TableFormatter) levelColor(level string) string {
	switch level {
	case "critical":
		return f.color(colorRed+colorBold, level)
	case "error":
		return f.color(colorRed, level)
	case "warning":
		return f.color(colorYellow, level)
	case "info":
		return f.color(colorBlue, level)
	case "debug":
		return f.color(colorGray, level)
	default:
		return level
	}
}

func (f *TableFormatter) statusColor(status string) string {
	switch status {
	case "active":
		return f.color(colorRed, status)
	case "resolved":
		return f.color(colorGreen, status)
	case "muted":
		return f.color(colorGray, status)
	default:
		return status
	}
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		m := int(d.Minutes())
		if m == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", m)
	case d < 24*time.Hour:
		h := int(d.Hours())
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	case d < 7*24*time.Hour:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	default:
		weeks := int(d.Hours() / 24 / 7)
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func (f *TableFormatter) FormatItems(w io.Writer, items []api.Item) error {
	if len(items) == 0 {
		fmt.Fprintln(w, "No items found.")
		return nil
	}

	// Header
	fmt.Fprintf(w, "%-7s %-50s %-10s %-10s %8s %-15s\n",
		"#", "TITLE", "LEVEL", "STATUS", "OCC", "LAST SEEN")
	fmt.Fprintln(w, strings.Repeat("-", 110))

	for _, item := range items {
		title := truncate(item.Title, 50)
		fmt.Fprintf(w, "%-7d %-50s %-10s %-10s %8d %-15s\n",
			item.Counter,
			title,
			f.levelColor(item.LevelString),
			f.statusColor(item.Status),
			item.TotalOccurrences,
			formatRelativeTime(item.LastOccurrenceTime),
		)
	}

	return nil
}

func (f *TableFormatter) FormatItem(w io.Writer, item *api.Item) error {
	fmt.Fprintf(w, "%s Item #%d: %s\n\n",
		f.color(colorBold, ""),
		item.Counter,
		item.Title,
	)

	fmt.Fprintf(w, "Level:       %s\n", f.levelColor(item.LevelString))
	fmt.Fprintf(w, "Status:      %s\n", f.statusColor(item.Status))
	fmt.Fprintf(w, "Environment: %s\n", item.Environment)
	fmt.Fprintf(w, "Framework:   %s\n", item.Framework)
	fmt.Fprintf(w, "Platform:    %s\n", item.Platform)
	fmt.Fprintln(w)
	fmt.Fprintf(w, "Total Occurrences: %d\n", item.TotalOccurrences)
	fmt.Fprintf(w, "First Seen:        %s (%s)\n",
		item.FirstOccurrenceTime.Format(time.RFC3339),
		formatRelativeTime(item.FirstOccurrenceTime))
	fmt.Fprintf(w, "Last Seen:         %s (%s)\n",
		item.LastOccurrenceTime.Format(time.RFC3339),
		formatRelativeTime(item.LastOccurrenceTime))

	return nil
}

func (f *TableFormatter) FormatInstances(w io.Writer, instances []api.Instance) error {
	if len(instances) == 0 {
		fmt.Fprintln(w, "No occurrences found.")
		return nil
	}

	fmt.Fprintf(w, "%-20s %-12s %-40s %-20s\n",
		"ID", "LEVEL", "MESSAGE", "TIME")
	fmt.Fprintln(w, strings.Repeat("-", 100))

	for _, inst := range instances {
		msg := ""
		if inst.Data.Body.Trace != nil {
			msg = truncate(inst.Data.Body.Trace.Exception.Message, 40)
		} else if inst.Data.Body.Message != nil {
			msg = truncate(inst.Data.Body.Message.Body, 40)
		}

		fmt.Fprintf(w, "%-20d %-12s %-40s %-20s\n",
			inst.ID,
			f.levelColor(inst.Data.Level),
			msg,
			formatRelativeTime(inst.Time),
		)
	}

	return nil
}

func (f *TableFormatter) FormatInstance(w io.Writer, instance *api.Instance) error {
	fmt.Fprintf(w, "%sOccurrence %d%s\n\n", colorBold, instance.ID, colorReset)

	fmt.Fprintf(w, "Level:       %s\n", f.levelColor(instance.Data.Level))
	fmt.Fprintf(w, "Environment: %s\n", instance.Data.Environment)
	fmt.Fprintf(w, "Time:        %s (%s)\n",
		instance.Time.Format(time.RFC3339),
		formatRelativeTime(instance.Time))

	if instance.Data.Body.Trace != nil {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%sException%s\n", colorBold, colorReset)
		fmt.Fprintf(w, "  Class:   %s\n", instance.Data.Body.Trace.Exception.Class)
		fmt.Fprintf(w, "  Message: %s\n", instance.Data.Body.Trace.Exception.Message)

		if len(instance.Data.Body.Trace.Frames) > 0 {
			fmt.Fprintln(w)
			fmt.Fprintf(w, "%sStack Trace%s\n", colorBold, colorReset)
			for i, frame := range instance.Data.Body.Trace.Frames {
				if i >= 10 {
					fmt.Fprintf(w, "  ... and %d more frames\n", len(instance.Data.Body.Trace.Frames)-10)
					break
				}
				fmt.Fprintf(w, "  %s:%d in %s()\n", frame.Filename, frame.Lineno, frame.Method)
			}
		}
	}

	if instance.Data.Request != nil && instance.Data.Request.URL != "" {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%sRequest%s\n", colorBold, colorReset)
		fmt.Fprintf(w, "  %s %s\n", instance.Data.Request.Method, instance.Data.Request.URL)
	}

	if instance.Data.Person != nil && instance.Data.Person.ID != "" {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%sPerson%s\n", colorBold, colorReset)
		if instance.Data.Person.Email != "" {
			fmt.Fprintf(w, "  %s (%s)\n", instance.Data.Person.Email, instance.Data.Person.ID)
		} else {
			fmt.Fprintf(w, "  ID: %s\n", instance.Data.Person.ID)
		}
	}

	return nil
}

func (f *TableFormatter) FormatContext(w io.Writer, item *api.Item, instances []api.Instance) error {
	// Just use item formatter for table output
	if err := f.FormatItem(w, item); err != nil {
		return err
	}

	if len(instances) > 0 {
		fmt.Fprintln(w)
		fmt.Fprintf(w, "%sRecent Occurrences%s\n", colorBold, colorReset)
		for _, inst := range instances {
			fmt.Fprintf(w, "\n  %s [%s]\n", inst.Time.Format(time.RFC3339), inst.Data.Level)
			if inst.Data.Body.Trace != nil {
				fmt.Fprintf(w, "    %s: %s\n",
					inst.Data.Body.Trace.Exception.Class,
					inst.Data.Body.Trace.Exception.Message)
			}
		}
	}

	return nil
}

func (f *TableFormatter) FormatProjectInfo(w io.Writer, info *api.ProjectInfo) error {
	fmt.Fprintf(w, "Project: %s (ID: %d)\n", info.Name, info.ID)
	fmt.Fprintln(w, "Authentication: OK")
	return nil
}
