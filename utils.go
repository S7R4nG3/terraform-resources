package tfresources

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
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

func testResults(name string, diffs []string) string {
	yellow := "\033[33m"
	end := "\033[0m"
	s := []string{fmt.Sprintf("FAILED :: %s :: %sDiffs: ", name, newline())}
	for _, d := range diffs {
		t := yellow + d + end
		s = append(s, t)
	}
	r := strings.Join(s, newline())
	return r
}
