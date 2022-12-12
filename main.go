//go:build (darwin && cgo) || linux || windows

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type BookResult struct {
	Title   string `json:"full_title"`
	Authors []struct {
		Key string `json:"key"`
	} `json:"authors"`
}

type Book struct {
	Title   string `json:"title"`
	Authors []struct {
		Author struct {
			Key string `json:"key"`
		} `json:"author"`
	} `json:"authors"`
	Key            string `json:"key"`
	LatestRevision int    `json:"latest_revision"`
	Revision       int    `json:"revision"`
	Created        struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"created"`
	LastModified struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	} `json:"last_modified"`
}

type WorksResult struct {
	Links struct {
		Self   string `json:"self"`
		Author string `json:"author"`
		Next   string `json:"next"`
	} `json:"links"`
	Size    int    `json:"size"`
	Entries []Book `json:"entries"`
}

type AuthorOutput struct {
	Name  string `json:"name"`
	Books struct {
		Key string `json:"key"`
	} `json:"books"`
}

// GoLang task:
// 1. Create simple cli application for finding all works from authors of specific book
// 2. Application has to find all authors for book and it will print list of all their works
// 3. Create list of works for each author (name, revision)
// 4. Print result to stout in yam format sorted by author name, count of revision (asc, desc as argument). Names of authors have to be part of output.
// As source of information about books, works and authors use this api: https://openlibrary.org/developers/api
func main() {
	fmt.Println("Retrieving Book by ISBN.")
	if len(os.Args) < 2 {
		_, _ = fmt.Fprintf(os.Stderr, "%d argument required! Please provide an ISBN as an argument of the specific Book you want to search\n", 1)
		os.Exit(1)
	}

	s := fmt.Sprintf("https://openlibrary.org/isbn/%s.json", url.QueryEscape(os.Args[1]))

	//fmt.Printf("URL to be called: %s\n", s)

	response, err := http.Get(s)
	if err != nil {
		panic(err)
	}
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var bookResponse BookResult
	_ = json.Unmarshal(responseBytes, &bookResponse)

	//fmt.Printf("Book retrieved: %s\n", responseBytes)
	fmt.Printf("\n\nAPI Response as struct:\n%+v\n\n", bookResponse)
	fmt.Printf("Book retrieved: %s, \n", bookResponse.Title)

	//var output []AuthorOutput

	for _, author := range bookResponse.Authors {
		s := fmt.Sprintf("https://openlibrary.org%s.json", author.Key)

		resp, authorErr := http.Get(s)
		if authorErr != nil {
			panic(authorErr)
		}
		authorBytes, authorErr := io.ReadAll(resp.Body)
		if authorErr != nil {
			panic(authorErr)
		}

		fmt.Printf("Author retrieved: %s\n", authorBytes)
		var authorData map[string]interface{}
		_ = json.Unmarshal(authorBytes, &authorData)
		fmt.Printf("\n\nAuthor RAW:\n%+v\n\n", authorData)
		fmt.Println(authorData["name"])

		s = fmt.Sprintf("https://openlibrary.org%s/works.json?limit=100", author)

		//fmt.Printf("URL to be called: %s\n", s)

		resp, authorErr = http.Get(s)
		if authorErr != nil {
			panic(authorErr)
		}
		authorBytes, authorErr = io.ReadAll(resp.Body)
		if authorErr != nil {
			panic(authorErr)
		}

		var authorResponse WorksResult
		_ = json.Unmarshal(authorBytes, &authorResponse)

		fmt.Printf("Author retrieved: %s\n", authorBytes)
		fmt.Printf("\n\nBooks:\n%+v\n\n", authorResponse.Entries)
		//fmt.Printf("\n\nAPI Response as struct:\n%+v\n\n", bookResponse)
		//fmt.Printf("Book retrieved: %s, \n", bookResponse.Title)
	}
}

func request(api string, value string) (*http.Response, error) {
	s := fmt.Sprintf("https://openlibrary.org/search.json?title=%s", url.QueryEscape(value))
	resp, err := http.Get(s)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
