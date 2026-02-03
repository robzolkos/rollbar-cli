package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/robzolkos/rollbar-cli/internal/api"
)

// CompactFormatter outputs minimal, token-efficient format for AI agents
type CompactFormatter struct{}

func formatCompactTime(t time.Time) string {
	if t.IsZero() {
		return "?"
	}
	d := time.Since(t)
	switch {
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// isAppFrame returns true if the frame is from app code (not vendor/gem/node_modules)
func isAppFrame(frame api.Frame) bool {
	f := frame.Filename
	// Vendor/gem patterns to exclude
	vendorPatterns := []string{
		"/vendor/",
		"/bundle/",
		"/gems/",
		"node_modules/",
		"/usr/lib/",
		"/usr/local/",
		".rvm/",
		".rbenv/",
	}
	for _, pattern := range vendorPatterns {
		if strings.Contains(f, pattern) {
			return false
		}
	}
	// App code patterns to include
	appPatterns := []string{
		"/app/app/",
		"/app/lib/",
		"/app/config/",
		"/src/",
		"/lib/",
	}
	for _, pattern := range appPatterns {
		if strings.Contains(f, pattern) {
			return true
		}
	}
	// If it starts with /app/ but isn't vendor, likely app code
	if strings.HasPrefix(f, "/app/") && !strings.Contains(f, "/vendor/") {
		return true
	}
	// Browser JS without node_modules is likely app code
	if strings.HasPrefix(f, "http") && !strings.Contains(f, "node_modules") {
		// Check if it's a minified bundle (long hash in filename)
		if strings.Contains(f, "-") && len(f) > 100 {
			return false // Likely minified bundle
		}
		return true
	}
	return false
}

// separateFrames splits frames into app code and vendor code
func separateFrames(frames []api.Frame) (app []api.Frame, vendor []api.Frame) {
	for _, frame := range frames {
		if isAppFrame(frame) {
			app = append(app, frame)
		} else {
			vendor = append(vendor, frame)
		}
	}
	return
}

func (f *CompactFormatter) FormatItems(w io.Writer, items []api.Item) error {
	for _, item := range items {
		// First line: counter, title, level, occurrences
		fmt.Fprintf(w, "#%d %s [%s] %d occ\n",
			item.Counter,
			item.Title,
			item.LevelString,
			item.TotalOccurrences,
		)
		// Second line: timing info
		fmt.Fprintf(w, "  Last: %s | First: %s\n",
			formatCompactTime(item.LastOccurrenceTime),
			formatCompactTime(item.FirstOccurrenceTime),
		)
	}
	return nil
}

func (f *CompactFormatter) FormatItem(w io.Writer, item *api.Item) error {
	fmt.Fprintf(w, "#%d %s\n", item.Counter, item.Title)
	fmt.Fprintf(w, "Level: %s | Status: %s | Occ: %d\n",
		item.LevelString, item.Status, item.TotalOccurrences)
	fmt.Fprintf(w, "Env: %s | Framework: %s\n", item.Environment, item.Framework)
	fmt.Fprintf(w, "Last: %s | First: %s\n",
		formatCompactTime(item.LastOccurrenceTime),
		formatCompactTime(item.FirstOccurrenceTime))
	return nil
}

func (f *CompactFormatter) FormatInstances(w io.Writer, instances []api.Instance) error {
	for _, inst := range instances {
		msg := ""
		if inst.Data.Body.Trace != nil {
			msg = inst.Data.Body.Trace.Exception.Class + ": " + inst.Data.Body.Trace.Exception.Message
		} else if inst.Data.Body.Message != nil {
			msg = inst.Data.Body.Message.Body
		}
		if len(msg) > 80 {
			msg = msg[:77] + "..."
		}
		fmt.Fprintf(w, "%d [%s] %s %s\n",
			inst.ID,
			inst.Data.Level,
			formatCompactTime(inst.Time),
			msg,
		)
	}
	return nil
}

func (f *CompactFormatter) FormatInstance(w io.Writer, instance *api.Instance) error {
	fmt.Fprintf(w, "Occurrence %d\n", instance.ID)
	fmt.Fprintf(w, "Level: %s | Env: %s | Time: %s\n",
		instance.Data.Level,
		instance.Data.Environment,
		formatCompactTime(instance.Time))

	if instance.Data.Body.Trace != nil {
		fmt.Fprintf(w, "Exception: %s: %s\n",
			instance.Data.Body.Trace.Exception.Class,
			instance.Data.Body.Trace.Exception.Message)

		if len(instance.Data.Body.Trace.Frames) > 0 {
			fmt.Fprintln(w, "Stack:")
			for i, frame := range instance.Data.Body.Trace.Frames {
				if i >= 5 {
					fmt.Fprintf(w, "  ...+%d more\n", len(instance.Data.Body.Trace.Frames)-5)
					break
				}
				fmt.Fprintf(w, "  %s:%d %s()\n", frame.Filename, frame.Lineno, frame.Method)
			}
		}
	}

	if instance.Data.Request != nil && instance.Data.Request.URL != "" {
		fmt.Fprintf(w, "Request: %s %s\n", instance.Data.Request.Method, instance.Data.Request.URL)
		if ua := instance.Data.Request.Headers["User-Agent"]; ua != "" {
			fmt.Fprintf(w, "Browser: %s\n", ua)
		}
	}

	if instance.Data.Person != nil && instance.Data.Person.ID != "" {
		person := string(instance.Data.Person.ID)
		if instance.Data.Person.Email != "" {
			person = instance.Data.Person.Email
		}
		fmt.Fprintf(w, "Person: %s\n", person)
	}

	return nil
}

func (f *CompactFormatter) FormatContext(w io.Writer, item *api.Item, instances []api.Instance) error {
	fmt.Fprintf(w, "# Error #%d: %s\n\n", item.Counter, item.Title)
	fmt.Fprintf(w, "Level: %s | Status: %s | Occ: %d\n", item.LevelString, item.Status, item.TotalOccurrences)
	fmt.Fprintf(w, "Env: %s | Framework: %s\n", item.Environment, item.Framework)
	fmt.Fprintf(w, "First: %s | Last: %s\n\n",
		item.FirstOccurrenceTime.Format(time.RFC3339),
		item.LastOccurrenceTime.Format(time.RFC3339))

	if len(instances) > 0 {
		inst := instances[0]
		if inst.Data.Body.Trace != nil {
			fmt.Fprintf(w, "## Exception\n%s: %s\n\n",
				inst.Data.Body.Trace.Exception.Class,
				inst.Data.Body.Trace.Exception.Message)

			if len(inst.Data.Body.Trace.Frames) > 0 {
				appFrames, vendorFrames := separateFrames(inst.Data.Body.Trace.Frames)

				// Show app code first (most useful for debugging)
				if len(appFrames) > 0 {
					fmt.Fprintln(w, "## App Code (source of error)")
					for i, frame := range appFrames {
						if i >= 10 {
							fmt.Fprintf(w, "  ...+%d more app frames\n", len(appFrames)-10)
							break
						}
						// Show code context if available
						if frame.Code != "" {
							fmt.Fprintf(w, "%s:%d %s()\n  > %s\n", frame.Filename, frame.Lineno, frame.Method, frame.Code)
						} else {
							fmt.Fprintf(w, "%s:%d %s()\n", frame.Filename, frame.Lineno, frame.Method)
						}
					}
					fmt.Fprintln(w)
				} else {
					fmt.Fprintln(w, "## Stack Trace")
					fmt.Fprintln(w, "⚠ No app code found in stack trace (only vendor/gem frames)")
					fmt.Fprintln(w, "  Tip: Configure Rollbar to capture app frames or upload sourcemaps")
					fmt.Fprintln(w)
				}

				// Show vendor frames (collapsed)
				if len(vendorFrames) > 0 && len(appFrames) == 0 {
					fmt.Fprintln(w, "## Vendor Frames (top 5)")
					for i, frame := range vendorFrames {
						if i >= 5 {
							fmt.Fprintf(w, "  ...+%d more vendor frames\n", len(vendorFrames)-5)
							break
						}
						fmt.Fprintf(w, "%s:%d %s()\n", frame.Filename, frame.Lineno, frame.Method)
					}
					fmt.Fprintln(w)
				}
			}
		}

		if inst.Data.Request != nil && inst.Data.Request.URL != "" {
			fmt.Fprintln(w, "## Request")
			fmt.Fprintf(w, "%s %s\n", inst.Data.Request.Method, inst.Data.Request.URL)
			if inst.Data.Request.UserIP != "" {
				fmt.Fprintf(w, "IP: %s\n", inst.Data.Request.UserIP)
			}
			if ua := inst.Data.Request.Headers["User-Agent"]; ua != "" {
				fmt.Fprintf(w, "Browser: %s\n", ua)
			}
			fmt.Fprintln(w)
		}

		if inst.Data.Person != nil && inst.Data.Person.ID != "" {
			fmt.Fprintln(w, "## Person")
			parts := []string{"ID: " + string(inst.Data.Person.ID)}
			if inst.Data.Person.Email != "" {
				parts = append(parts, "Email: "+inst.Data.Person.Email)
			}
			if inst.Data.Person.Username != "" {
				parts = append(parts, "Username: "+inst.Data.Person.Username)
			}
			fmt.Fprintln(w, strings.Join(parts, " | "))
			fmt.Fprintln(w)
		}

		if inst.Data.Server != nil && inst.Data.Server.Host != "" {
			fmt.Fprintln(w, "## Server")
			fmt.Fprintf(w, "Host: %s\n", inst.Data.Server.Host)
			if inst.Data.Server.Branch != "" {
				fmt.Fprintf(w, "Branch: %s\n", inst.Data.Server.Branch)
			}
			if inst.Data.Server.CodeVersion != "" {
				fmt.Fprintf(w, "Version: %s\n", inst.Data.Server.CodeVersion)
			}
		}
	}

	return nil
}

func (f *CompactFormatter) FormatProjectInfo(w io.Writer, info *api.ProjectInfo) error {
	fmt.Fprintf(w, "Project: %s (ID: %d) - OK\n", info.Name, info.ID)
	return nil
}
