package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sync"
)

func readLinesFromFile(filename string) ([]string, error) {
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

func processURL(url string, keywordRegexes []*regexp.Regexp, titleRegex *regexp.Regexp, wg *sync.WaitGroup) {
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

func main() {
	keywords, err := readLinesFromFile("keywords.txt")
	if err != nil {
		panic(fmt.Sprintf("Failed to read keywords from file: %v", err))
	}

	urls, err := readLinesFromFile("urls.txt")
	if err != nil {
		panic(fmt.Sprintf("Failed to read URLs from file: %v", err))
	}

	titleRegex := regexp.MustCompile(".*'>(.*?)<span class=\"vs\">.*")
	keywordRegexes := make([]*regexp.Regexp, len(keywords))
	for i, keyword := range keywords {
		keywordRegexes[i] = regexp.MustCompile(keyword)
	}

	const concurrentLimit = 5
	sem := make(chan struct{}, concurrentLimit) // semaphore pattern for limiting concurrency
	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		sem <- struct{}{} // acquire a token
		go func(u string) {
			processURL(u, keywordRegexes, titleRegex, &wg)
			<-sem // release a token
		}(url)
	}

	wg.Wait()
}
