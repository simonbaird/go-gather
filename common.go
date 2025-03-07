package gogather

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// URLType is an enum for URL types
type URIType int

const (
	GitURI URIType = iota
	HTTPURI
	FileURI
	Unknown
)

var GetHomeDir = os.UserHomeDir

// String returns the string representation of the URLType
func (t URIType) String() string {
	return [...]string{"GitURI", "HTTPURI", "FileURI", "Unknown"}[t]
}

// ExpandTilde expands a leading tilde in the file path to the user's home directory
func ExpandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := GetHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[2:])
	}
	return path
}

// ClassifyURI classifies the input string as a Git URI, HTTP(S) URI, or file path
func ClassifyURI(input string) (URIType, error) {
	// Check for special prefixes first
	if strings.HasPrefix(input, "git::") {
		return GitURI, nil
	}
	if strings.HasPrefix(input, "file::") {
		return FileURI, nil
	}
	if strings.HasPrefix(input, "http::") {
		return HTTPURI, nil
	}

	if strings.HasPrefix(input, "github.com") || strings.HasPrefix(input, "gitlab.com") {
		return GitURI, nil
	}

	// Regular expression for Git URIs
	gitURIPattern := regexp.MustCompile(`^(git@[\w\.\-]+:[\w\.\-]+/[\w\.\-]+(\.git)?|https?://[\w\.\-]+/[\w\.\-]+/[\w\.\-]+(\.git)?|git://[\w\.\-]+/[\w\.\-]+/[\w\.\-]+(\.git)?|[\w\.\-]+/[\w\.\-]+/[\w\.\-]+//.*|file://.*\.git|[\w\.\-]+/[\w\.\-]+(\.git)?)$`)
	// Regular expression for HTTP URIs (with or without protocol)
	httpURIPattern := regexp.MustCompile(`^((http://|https://)[\w\-]+(\.[\w\-]+)+.*)$`)
	// Regular expression for file paths
	filePathPattern := regexp.MustCompile(`^(\./|\../|/|[a-zA-Z]:\\|~\/|file://).*`)

	// Check if the input matches the file path pattern first
	if filePathPattern.MatchString(input) {
		// Expand the tilde in the file path if it exists
		input = ExpandTilde(input)
		// Check if the input ends with ".git" to classify as GitURI
		if strings.HasSuffix(input, ".git") {
			return GitURI, nil
		}
		return FileURI, nil
	}

	// Check if the input matches the Git URI pattern
	if gitURIPattern.MatchString(input) {
		return GitURI, nil
	}

	// Check if the input matches the HTTP URI pattern
	if httpURIPattern.MatchString(input) {
		// Parse the input as a URI
		parsedURI, err := url.Parse(input)
		if err == nil && (parsedURI.Scheme == "http" || parsedURI.Scheme == "https") {
			return HTTPURI, nil
		}
	}

	// Check for unsupported schemes
	parsedURI, err := url.Parse(input)
	if err == nil && parsedURI.Scheme != "" && parsedURI.Scheme != "http" && parsedURI.Scheme != "https" {
		return Unknown, fmt.Errorf("unsupported source protocol: %s", parsedURI.Scheme)
	}

	// Check if the input contains a dot but lacks a valid scheme
	if strings.Contains(input, ".") {
		return Unknown, fmt.Errorf("got %s. HTTP(S) URIs require a scheme (http:// or https://)", input)
	}

	return Unknown, nil
}

// ValidateFileDestination validates the d1estination path for saving files
func ValidateFileDestination(destination string) error {
	// Expand the tilde in the file path if it exists
	destination = ExpandTilde(destination)
	// Check if the destination file exists.
	_, err := os.Stat(destination)
	if err == nil {
		return fmt.Errorf("destination file already exists: %s", destination)
	}
	if os.IsNotExist(err) {
		return nil
	}
	return nil
}
