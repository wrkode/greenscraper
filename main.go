package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

var keywords = []string{"carbon", "climate", "energy", "green", "kepler", "sustainability", "sustainable"}

func main() {
	// Read URLs from urls.txt file into a slice
	file, err := os.Open("urls.txt")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	// Loop through the list of URLs
	for _, url := range urls {
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Failed to fetch URL %s: %v\n", url, err)
			continue
		}

		contentBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body from URL %s: %v\n", url, err)
			continue
		}
		content := string(contentBytes)
		resp.Body.Close()

		var talks []string
		encounteredTitles := make(map[string]bool)

		// Loop through the list of keywords
		for _, keyword := range keywords {
			regex := regexp.MustCompile(keyword)
			matches := regex.FindAllIndex([]byte(content), -1)

			for _, match := range matches {
				// Extract the title from the matched content
				titleRegex := regexp.MustCompile(".*'>(.*?)<span class=\"vs\">.*")
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

		// Print the conference schedule link and "Talks:" section if talks were found
		if len(talks) > 0 {
			fmt.Println("Conference schedule link:", url)
			fmt.Println("Talks:")
			for _, talk := range talks {
				fmt.Println(talk)
			}
		}
	}
}
