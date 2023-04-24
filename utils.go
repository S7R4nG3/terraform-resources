package tfresources

import (
	"os"
	"regexp"
	"runtime"
)

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}

func newline() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	} else {
		return "\n"
	}

}

func removeString(regexes []string, source string) string {
	working := source
	for _, regex := range regexes {
		re := regexp.MustCompile(regex)
		working = re.ReplaceAllString(working, "")
	}
	return working
}
