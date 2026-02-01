package version

// These variables are set at build time via ldflags
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Info returns version information as a formatted string
func Info() string {
	return Version
}

// Full returns detailed version information
func Full() string {
	return "rollbar " + Version + " (commit: " + Commit + ", built: " + BuildDate + ")"
}
