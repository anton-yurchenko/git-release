package changelog

import (
	"bufio"
	"regexp"

	"github.com/spf13/afero"
)

// Read changelog line by line and return content as []string
func (c *Changes) Read(fs afero.Fs) ([]string, error) {
	lines := make([]string, 0)

	file, err := fs.Open(c.File)
	if err != nil {
		return lines, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, nil
}

// GetEndOfFirstRelease returns line number on which the first release ends.
// It may be an end of file or a start of an 'Unreleased Versions'.
func GetEndOfFirstRelease(content []string) int {
	expression := "^(?P<prefix>\\[)(?P<unreleased>[^\\]]*)(?P<postfix>\\][^(].*)$"
	regex := regexp.MustCompile(expression)

	for i := 0; i < len(content); i++ {
		if regex.MatchString(content[i]) {
			return i - 1
		}
	}

	return len(content)
}

// GetReleasesLines returns line numbers where each release starts
func GetReleasesLines(content []string) []int {
	lines := make([]int, 0)

	expression := "^(?P<prefix>##\\s*\\[)(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:-(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?(?P<postfix>\\])(?P<title>.*)$"
	regex := regexp.MustCompile(expression)

	for i, line := range content {
		if regex.MatchString(line) {
			lines = append(lines, i)
		}
	}

	return lines
}

// GetMargins returns margins of requested version body
func (c *Changes) GetMargins(content []string) map[string]int {
	margins := make(map[string]int)

	releaseLines := GetReleasesLines(content)

	for i, line := range releaseLines {
		expression := "^(?P<prefix>##\\s*\\[)(?P<major>0|[1-9]\\d*)\\.(?P<minor>0|[1-9]\\d*)\\.(?P<patch>0|[1-9]\\d*)(?:(?P<sep1>-)(?P<prerelease>(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:(?P<sep2>\\+)(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?(?P<postfix>\\])(?P<date>.*)$"
		regex := regexp.MustCompile(expression)
		var v string

		if regex.MatchString(content[line]) {
			if regex.ReplaceAllString(content[line], "${5}") == "-" {
				if regex.ReplaceAllString(content[line], "${7}") == "+" {
					v = regex.ReplaceAllString(content[line], "${2}.${3}.${4}${5}${6}${7}${8}")
				} else {
					v = regex.ReplaceAllString(content[line], "${2}.${3}.${4}${5}${6}")
				}
			} else if regex.ReplaceAllString(content[line], "${7}") == "+" {
				v = regex.ReplaceAllString(content[line], "${2}.${3}.${4}${7}${8}")
			} else {
				v = regex.ReplaceAllString(content[line], "${2}.${3}.${4}")
			}
		}

		if v == c.Version {
			margins["start"] = line + 1

			switch i < len(releaseLines)-1 {
			// not first version
			case true:
				margins["end"] = releaseLines[i+1] - 1
			// first version
			case false:
				margins["end"] = GetEndOfFirstRelease(content)
			}
		}
	}

	return margins
}

// GetContent returns lines between margins
func GetContent(margins map[string]int, content []string) []string {
	releseContent := make([]string, 0)

	if margins["start"] > len(content) || margins["end"] > len(content) {
		return releseContent
	}

	for i := margins["start"]; i < margins["end"]; i++ {
		releseContent = append(releseContent, content[i])
	}

	return releseContent
}
