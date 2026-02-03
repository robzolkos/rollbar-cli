package output

import (
	"fmt"
	"io"
	"time"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

// MarkdownFormatter outputs data as markdown for documentation/AI context
type MarkdownFormatter struct{}

func (f *MarkdownFormatter) FormatItems(w io.Writer, items []api.Item) error {
	fmt.Fprintln(w, "# Rollbar Items")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "| # | Title | Level | Status | Occurrences | Last Seen |")
	fmt.Fprintln(w, "|---|-------|-------|--------|-------------|-----------|")

	for _, item := range items {
		title := item.Title
		if len(title) > 60 {
			title = title[:57] + "..."
		}
		fmt.Fprintf(w, "| %d | %s | %s | %s | %d | %s |\n",
			item.Counter,
			title,
			item.LevelString,
			item.Status,
			item.TotalOccurrences,
			formatCompactTime(item.LastOccurrenceTime),
		)
	}

	return nil
}

func (f *MarkdownFormatter) FormatItem(w io.Writer, item *api.Item) error {
	fmt.Fprintf(w, "# Item #%d: %s\n\n", item.Counter, item.Title)

	fmt.Fprintln(w, "## Summary")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- **Level:** %s\n", item.LevelString)
	fmt.Fprintf(w, "- **Status:** %s\n", item.Status)
	fmt.Fprintf(w, "- **Environment:** %s\n", item.Environment)
	fmt.Fprintf(w, "- **Framework:** %s\n", item.Framework)
	fmt.Fprintf(w, "- **Platform:** %s\n", item.Platform)
	fmt.Fprintf(w, "- **Total Occurrences:** %d\n", item.TotalOccurrences)
	fmt.Fprintf(w, "- **First Seen:** %s\n", item.FirstOccurrenceTime.Format(time.RFC3339))
	fmt.Fprintf(w, "- **Last Seen:** %s\n", item.LastOccurrenceTime.Format(time.RFC3339))

	return nil
}

func (f *MarkdownFormatter) FormatInstances(w io.Writer, instances []api.Instance) error {
	fmt.Fprintln(w, "# Occurrences")
	fmt.Fprintln(w)

	for i, inst := range instances {
		fmt.Fprintf(w, "## Occurrence %d (ID: %d)\n\n", i+1, inst.ID)
		fmt.Fprintf(w, "- **Time:** %s\n", inst.Time.Format(time.RFC3339))
		fmt.Fprintf(w, "- **Level:** %s\n", inst.Data.Level)
		fmt.Fprintf(w, "- **Environment:** %s\n", inst.Data.Environment)

		if inst.Data.Body.Trace != nil {
			fmt.Fprintf(w, "- **Exception:** %s: %s\n",
				inst.Data.Body.Trace.Exception.Class,
				inst.Data.Body.Trace.Exception.Message)
		}
		fmt.Fprintln(w)
	}

	return nil
}

func (f *MarkdownFormatter) FormatInstance(w io.Writer, instance *api.Instance) error {
	fmt.Fprintf(w, "# Occurrence %d\n\n", instance.ID)

	fmt.Fprintln(w, "## Summary")
	fmt.Fprintf(w, "- **Time:** %s\n", instance.Time.Format(time.RFC3339))
	fmt.Fprintf(w, "- **Level:** %s\n", instance.Data.Level)
	fmt.Fprintf(w, "- **Environment:** %s\n", instance.Data.Environment)
	fmt.Fprintln(w)

	if instance.Data.Body.Trace != nil {
		fmt.Fprintln(w, "## Exception")
		fmt.Fprintf(w, "- **Class:** %s\n", instance.Data.Body.Trace.Exception.Class)
		fmt.Fprintf(w, "- **Message:** %s\n", instance.Data.Body.Trace.Exception.Message)
		fmt.Fprintln(w)

		if len(instance.Data.Body.Trace.Frames) > 0 {
			fmt.Fprintln(w, "## Stack Trace")
			fmt.Fprintln(w, "```")
			for _, frame := range instance.Data.Body.Trace.Frames {
				fmt.Fprintf(w, "%s:%d in %s()\n", frame.Filename, frame.Lineno, frame.Method)
			}
			fmt.Fprintln(w, "```")
			fmt.Fprintln(w)
		}
	}

	if instance.Data.Request != nil && instance.Data.Request.URL != "" {
		fmt.Fprintln(w, "## Request")
		fmt.Fprintf(w, "- **Method:** %s\n", instance.Data.Request.Method)
		fmt.Fprintf(w, "- **URL:** %s\n", instance.Data.Request.URL)
		if instance.Data.Request.UserIP != "" {
			fmt.Fprintf(w, "- **User IP:** %s\n", instance.Data.Request.UserIP)
		}
		if ua := instance.Data.Request.Headers["User-Agent"]; ua != "" {
			fmt.Fprintf(w, "- **Browser:** %s\n", ua)
		}
		fmt.Fprintln(w)
	}

	if instance.Data.Person != nil && instance.Data.Person.ID != "" {
		fmt.Fprintln(w, "## Person")
		fmt.Fprintf(w, "- **ID:** %s\n", instance.Data.Person.ID)
		if instance.Data.Person.Email != "" {
			fmt.Fprintf(w, "- **Email:** %s\n", instance.Data.Person.Email)
		}
		if instance.Data.Person.Username != "" {
			fmt.Fprintf(w, "- **Username:** %s\n", instance.Data.Person.Username)
		}
		fmt.Fprintln(w)
	}

	if instance.Data.Server != nil && instance.Data.Server.Host != "" {
		fmt.Fprintln(w, "## Server")
		fmt.Fprintf(w, "- **Host:** %s\n", instance.Data.Server.Host)
		if instance.Data.Server.Root != "" {
			fmt.Fprintf(w, "- **Root:** %s\n", instance.Data.Server.Root)
		}
		if instance.Data.Server.Branch != "" {
			fmt.Fprintf(w, "- **Branch:** %s\n", instance.Data.Server.Branch)
		}
		if instance.Data.Server.CodeVersion != "" {
			fmt.Fprintf(w, "- **Code Version:** %s\n", instance.Data.Server.CodeVersion)
		}
	}

	return nil
}

func (f *MarkdownFormatter) FormatContext(w io.Writer, item *api.Item, instances []api.Instance) error {
	fmt.Fprintf(w, "# Bug Report: %s\n\n", item.Title)

	fmt.Fprintln(w, "## Summary")
	fmt.Fprintf(w, "- **Rollbar Item:** #%d\n", item.Counter)
	fmt.Fprintf(w, "- **Level:** %s\n", item.LevelString)
	fmt.Fprintf(w, "- **Status:** %s\n", item.Status)
	fmt.Fprintf(w, "- **Total Occurrences:** %d\n", item.TotalOccurrences)
	fmt.Fprintf(w, "- **First Seen:** %s\n", item.FirstOccurrenceTime.Format(time.RFC3339))
	fmt.Fprintf(w, "- **Last Seen:** %s\n", item.LastOccurrenceTime.Format(time.RFC3339))
	fmt.Fprintf(w, "- **Environment:** %s\n", item.Environment)
	fmt.Fprintf(w, "- **Framework:** %s\n", item.Framework)
	fmt.Fprintln(w)

	if len(instances) > 0 {
		inst := instances[0]

		if inst.Data.Body.Trace != nil {
			fmt.Fprintln(w, "## Exception Details")
			fmt.Fprintf(w, "- **Type:** %s\n", inst.Data.Body.Trace.Exception.Class)
			fmt.Fprintf(w, "- **Message:** %s\n", inst.Data.Body.Trace.Exception.Message)
			fmt.Fprintln(w)

			if len(inst.Data.Body.Trace.Frames) > 0 {
				fmt.Fprintln(w, "## Stack Trace")
				fmt.Fprintln(w, "```")
				for _, frame := range inst.Data.Body.Trace.Frames {
					line := fmt.Sprintf("%s:%d in %s()", frame.Filename, frame.Lineno, frame.Method)
					fmt.Fprintln(w, line)
					if frame.Code != "" {
						fmt.Fprintf(w, "  > %s\n", frame.Code)
					}
				}
				fmt.Fprintln(w, "```")
				fmt.Fprintln(w)

				// Affected code location
				topFrame := inst.Data.Body.Trace.Frames[0]
				fmt.Fprintln(w, "## Affected Code Location")
				fmt.Fprintf(w, "- **File:** %s\n", topFrame.Filename)
				fmt.Fprintf(w, "- **Line:** %d\n", topFrame.Lineno)
				fmt.Fprintf(w, "- **Function:** %s()\n", topFrame.Method)
				fmt.Fprintln(w)
			}
		}

		// Recent occurrences
		fmt.Fprintf(w, "## Recent Occurrences (%d)\n\n", len(instances))
		for i, occ := range instances {
			fmt.Fprintf(w, "### Occurrence %d - %s\n", i+1, occ.Time.Format(time.RFC3339))
			if occ.Data.Request != nil && occ.Data.Request.URL != "" {
				fmt.Fprintf(w, "- **Request:** %s %s\n", occ.Data.Request.Method, occ.Data.Request.URL)
				if ua := occ.Data.Request.Headers["User-Agent"]; ua != "" {
					fmt.Fprintf(w, "- **Browser:** %s\n", ua)
				}
			}
			if occ.Data.Person != nil && occ.Data.Person.Email != "" {
				fmt.Fprintf(w, "- **User:** %s\n", occ.Data.Person.Email)
			} else if occ.Data.Person != nil && occ.Data.Person.ID != "" {
				fmt.Fprintf(w, "- **User ID:** %s\n", occ.Data.Person.ID)
			}
			if occ.Data.Server != nil && occ.Data.Server.Host != "" {
				fmt.Fprintf(w, "- **Server:** %s\n", occ.Data.Server.Host)
			}
			fmt.Fprintln(w)
		}
	}

	return nil
}

func (f *MarkdownFormatter) FormatProjectInfo(w io.Writer, info *api.ProjectInfo) error {
	fmt.Fprintln(w, "# Rollbar Project Info")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "- **Project Name:** %s\n", info.Name)
	fmt.Fprintf(w, "- **Project ID:** %d\n", info.ID)
	fmt.Fprintln(w, "- **Authentication:** OK")
	return nil
}
