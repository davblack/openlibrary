//go:build (darwin && cgo) || linux || windows

package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
)

type BookResult struct {
	Title   string `json:"full_title"`
	Authors []struct {
		Key string `json:"key"`
	} `json:"authors"`
	Error string `json:"error"`
}

type Author struct {
	AlternateNames []string `json:"alternate_names"`
	Name           string   `json:"name"`
	Key            string   `json:"key"`
	Revision       int      `json:"revision"`
}

type WorksResult struct {
	Links struct {
		Self   string `json:"self"`
		Author string `json:"author"`
		Next   string `json:"next"`
	} `json:"links"`
	Size    int         `json:"size"`
	Entries []BookShort `json:"entries"`
}

type BookShort struct {
	Title    string `json:"title"`
	Revision int    `json:"revision"`
}
type AuthorOutput struct {
	Name     string      `json:"name"`
	Revision int         `json:"revision"`
	Books    []BookShort `json:"books"`
}

// GoLang task:
// 1. Create simple cli application for finding all works from authors of specific book
// 2. Application has to find all authors for book and it will print list of all their works
// 3. Create list of works for each author (name, revision)
// 4. Print result to stout in yam format sorted by author name, count of revision (asc, desc as argument). Names of authors have to be part of output.
// As source of information about books, works and authors use this api: https://openlibrary.org/developers/api
func main() {
	//test_isbn := "9780062899149"
	//test_isbn := "9781633412545" // multiple authors

	if len(os.Args) < 2 { // check if at least one parameter is provided - required ISBN
		_, _ = fmt.Fprintf(os.Stderr, "%d argument required! Please provide an ISBN as an argument of the specific Book you want to search\n", 1)
		os.Exit(1) // if not exit with non-zero code 1
	}

	var sortingOrder = "asc" // DEFAULT sorting order
	if len(os.Args) > 2 && strings.ToLower(os.Args[2]) == "desc" {
		sortingOrder = "desc" // change order if other than default specified by argument
	}

	// retrieve data from OpenLibrary by book ISBN
	s := fmt.Sprintf("https://openlibrary.org/isbn/%s.json", url.QueryEscape(os.Args[1]))

	response, err := http.Get(s)
	if err != nil {
		panic(err)
	}
	if response.StatusCode == http.StatusNotFound { // error checking for 404 response (by response status code)
		_, _ = fmt.Fprintf(os.Stderr, "Book ISBN:%s not found - most likely invalid ISBN provided\n", os.Args[1])
		os.Exit(2)
	}
	responseBytes, err := io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	var bookResponse BookResult
	_ = json.Unmarshal(responseBytes, &bookResponse) // convert JSON from response to custom struct

	if len(bookResponse.Error) > 0 { // error checking for 404 response (in returned data)
		_, _ = fmt.Fprintf(os.Stderr, "Book not found by provided ISBN %s\n", os.Args[1])
		os.Exit(3)
	}
	//fmt.Printf("Book retrieved: %s\n", responseBytes)
	//fmt.Printf("\n\nAPI Response as struct:\n%+v\n\n", bookResponse)

	// prepare a 'buffer' for output data (length is equivalent to authors count of the in input book)
	output := make([]AuthorOutput, len(bookResponse.Authors))

	var i = 0
	for _, author := range bookResponse.Authors {
		// retrieve author data - name & revision of the author are not included in 'works' API response
		resp, authorErr := http.Get(fmt.Sprintf("https://openlibrary.org%s.json", author.Key))
		if authorErr != nil {
			panic(authorErr)
		}
		if response.StatusCode == http.StatusNotFound { // error checking for 404 response (by response status code)
			_, _ = fmt.Fprintf(os.Stderr, "Author %s not found\n", author.Key)
			os.Exit(2)
		}
		authorBytes, authorErr := io.ReadAll(resp.Body)
		if authorErr != nil {
			panic(authorErr)
		}

		var authorData Author
		_ = json.Unmarshal(authorBytes, &authorData)

		// retrieve author's books (with increased limit - default is 50)
		resp, authorErr = http.Get(fmt.Sprintf("https://openlibrary.org%s/works.json?limit=1000", author.Key))
		if authorErr != nil {
			panic(authorErr)
		}
		authorBytes, authorErr = io.ReadAll(resp.Body)
		if authorErr != nil {
			panic(authorErr)
		}

		var authorResponse WorksResult
		_ = json.Unmarshal(authorBytes, &authorResponse)

		//fmt.Printf("Author retrieved: %s\n", authorBytes)
		//fmt.Printf("\n\nBooks:\n%+v\n\n", authorResponse.Entries)

		output[i] = AuthorOutput{
			Name:     authorData.Name,
			Revision: authorData.Revision,
			Books:    authorResponse.Entries,
		}
		i++
	}

	// sort the output data
	sort.SliceStable(output, func(i, j int) bool {
		var a, b AuthorOutput
		switch sortingOrder {
		case "desc":
			// for desc the input values are flipped to allow reuse of the comparison logic
			a = output[j]
			b = output[i]
		default: // default ascending order
			a = output[i]
			b = output[j]
		}

		if a.Name == b.Name { // 'fallback' secondary order if names are equal
			return a.Revision < b.Revision
		}
		return a.Name < b.Name
	})

	yamlData, err := yaml.Marshal(&output) // generate output in YAML format
	if err != nil {
		panic(err)
	}

	fmt.Printf("%s\n", yamlData)
}
