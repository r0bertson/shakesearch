package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"index/suffixarray"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
)

func main() {
	searcher := Searcher{}
	err := searcher.Load("completeworks.txt")
	if err != nil {
		log.Fatal(err)
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

	http.HandleFunc("/search", handleSearch(searcher))

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	fmt.Printf("Listening on port %s...", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type Searcher struct {
	CompleteWorks string
	SuffixArray   *suffixarray.Index
}

func handleSearch(searcher Searcher) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		query, ok := r.URL.Query()["q"]
		if !ok || len(query[0]) < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("missing search query in URL params"))
			return
		}
		results := searcher.Search(query[0])
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		err := enc.Encode(results)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("encoding failure"))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf.Bytes())
	}
}

func (s *Searcher) Load(filename string) error {
	dat, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("load: %w", err)
	}
	s.CompleteWorks = string(dat)
	s.CompleteWorks = strings.ReplaceAll(s.CompleteWorks, "\r\n", "<br />") //TODO: PARSE THIS WHEN FORMATTING
	lowerCased := strings.ToLower(s.CompleteWorks)
	s.SuffixArray = suffixarray.New([]byte(lowerCased))
	return nil
}

type ChunkedResult struct {
	Indexes []int
	Result  string
}

const SearchPreSuffixSize = 250

func (s *Searcher) Search(query string) []string {
	idxs := s.SuffixArray.Lookup([]byte(query), -1)
	sort.Ints(idxs)

	results := []string{}
	chunks := []ChunkedResult{}
	currentChunk := ChunkedResult{}
	currentIndexValue := 0
	numberOfIndexes := len(idxs)
	for i := 0; i < numberOfIndexes; i++ {
		currentIndexValue = idxs[i]
		currentChunk.Indexes = append(currentChunk.Indexes, currentIndexValue)
		nextIndex := i + 1
		if nextIndex < numberOfIndexes && idxs[nextIndex]-currentIndexValue > SearchPreSuffixSize {
			chunks = append(chunks, currentChunk)
			currentChunk = ChunkedResult{}
		}
	}
	for _, chunk := range chunks {
		results = append(results, s.CompleteWorks[chunk.Indexes[0]-SearchPreSuffixSize:chunk.Indexes[len(chunk.Indexes)-1]+SearchPreSuffixSize])
	}
	return results
}
