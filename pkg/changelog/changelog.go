package changelog

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// GetBody returns section from changelog for provided version
func GetBody(version string, filename string) (*string, error) {
	var body string

	file, err := read(filename)
	if err != nil {
		return &body, err
	}

	margins := getMargins(version, file)

	body = strings.Join(getContent(margins, file), "\n")

	return &body, nil
}

// read file line by line to []string
func read(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return []string{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

// getReleasesLines returns line numbers where every release start in changelog
func getReleasesLines(file []string) []int {
	regex := regexp.MustCompile("## \\[[0-9]+.[0-9]+.[0-9]+\\].*")
	var lines []int

	for i, line := range file {
		if regex.MatchString(line) {
			lines = append(lines, i)
		}
	}

	return lines
}

// getEndOfFirstRelease returns line number when first releast ends.
// It may be an end of file or a start of an 'Unreleased'.
func getEndOfFirstRelease(start int, file []string) int {
	regex := regexp.MustCompile("\\[.*\\]:.*")

	for i := start; i < len(file); i++ {
		if regex.MatchString(file[i]) {
			return i - 1
		}
	}

	return len(file)
}

// getMargins returns margins of requested version body
func getMargins(version string, file []string) map[string]int {
	releaseLines := getReleasesLines(file)

	margins := make(map[string]int)
	for i, line := range releaseLines {
		v := strings.Split(strings.Trim(file[line], "## ["), "] ")[0]

		if v == version {
			margins["start"] = line + 1

			switch i < len(releaseLines)-1 {
			// not first version
			case true:
				margins["end"] = releaseLines[i+1] - 1
			// first version
			case false:
				margins["end"] = getEndOfFirstRelease(margins["start"], file)
			}
		}
	}

	return margins
}

// getContent returns lines between margins
func getContent(margins map[string]int, file []string) []string {
	var content []string

	for i := margins["start"]; i < margins["end"]; i++ {
		content = append(content, file[i])
	}

	return content
}
