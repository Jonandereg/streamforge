// Package version contains build-time version information for StreamForge.
package version

// Version is the semantic version of the build, injected at build time.
var Version = "dev"

// Commit is the git commit hash of the build, injected at build time.
var Commit = "none"

// BuildDate is the UTC date the binary was built, injected at build time.
var BuildDate = "unknown"
