package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
)

var Version = "" //yet to be implemented.

func ReadLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func ProcessURL(url string, keywordRegexes []*regexp.Regexp, titleRegex *regexp.Regexp, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to fetch URL %s: %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	contentBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body from URL %s: %v\n", url, err)
		return
	}
	content := string(contentBytes)

	var talks []string
	encounteredTitles := make(map[string]bool)

	for _, regex := range keywordRegexes {
		matches := regex.FindAllIndex([]byte(content), -1)

		for _, match := range matches {
			titleMatch := titleRegex.FindStringSubmatch(content[match[0]:])
			if len(titleMatch) > 1 {
				title := titleMatch[1]

				if len(title) >= 40 && !encounteredTitles[title] {
					encounteredTitles[title] = true
					talks = append(talks, "- "+title)
				}
			}
		}
	}

	if len(talks) > 0 {
		fmt.Println("Conference schedule link:", url)
		fmt.Println("Talks:")
		for _, talk := range talks {
			fmt.Println(talk)
		}
	}
}
