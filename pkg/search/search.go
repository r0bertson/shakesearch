package search

import (
	"fmt"
	"index/suffixarray"
	"os"
	"pulley.com/shakesearch/pkg/utils"
	"regexp"
	"sort"
	"strings"
)

const SectionSeparator = "☞"
const WorkSeparator = "►"
const PreSuffixSize = 250
const WindowsLineBreak = "\r\n"
const HTMLLineBreak = "<br>"

var lineBreakRegExp = regexp.MustCompile(`\r?\n`)

// WorkSearcher is a search for all Shakespeare works.
type WorkSearcher struct {
	Works []Work
}

// Work carries a reference to a specific work and its index.
type Work struct {
	Title       string
	Text        string
	SuffixArray *suffixarray.Index
}

// Load reads a file and split it into works, assuming that this file is formatted using the special characters
// WorkSeparator and SectionSeparator
func (ws *WorkSearcher) Load(filename string) error {
	dat, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("load: %w", err)
	}
	all := string(dat)
	worksRaw := strings.Split(all, WorkSeparator)[1:]
	works := []Work{}
	for _, workRaw := range worksRaw {
		workRaw = strings.ReplaceAll(workRaw, SectionSeparator, "") //this is a future feature
		title := lineBreakRegExp.Split(workRaw, -1)[0]
		workRaw = workRaw[strings.Index(workRaw, "\r\n"):]
		lower := strings.ToLower(workRaw)
		works = append(works, Work{
			Title:       strings.TrimSpace(title),
			Text:        workRaw,
			SuffixArray: suffixarray.New([]byte(lower)),
		})

	}
	ws.Works = works
	return nil
}

// Result contains a reference to a work title and all fragments that contains the queried terms.
type Result struct {
	WorkTitle string   `json:"work_title"`
	Fragments []string `json:"fragments"`
	//TODO: fragments could be verse-based and used instead of PreSuffixSize
}

// Search checks if the provided keywords appear in a work, returning each fragment and work title.
func (w *Work) Search(keys []string) *Result {
	idxs := []int{}
	for _, key := range keys {
		idxs = append(idxs, w.SuffixArray.Lookup([]byte(key), -1)...)
	}
	if len(idxs) == 0 {
		return nil
	}
	idxs = utils.Unique(idxs)
	sort.Ints(idxs)

	chunks := chunkSimilarResults(idxs)
	fragments := []string{}
	for _, chunk := range chunks {
		start := chunk.Indexes[0] - PreSuffixSize
		end := chunk.Indexes[len(chunk.Indexes)-1] + PreSuffixSize
		if start < 0 {
			start = 0
		}
		if end > len(w.Text)-1 {
			end = len(w.Text) - 1
		}
		fragments = append(fragments, w.Text[start:end])
	}
	return &Result{
		WorkTitle: w.Title,
		Fragments: fragments,
	}
}

// Search queries all works looking for the keywords provided.
func (ws *WorkSearcher) Search(keywords []string) []Result {
	results := []Result{}
	for _, work := range ws.Works {
		if result := work.Search(keywords); result != nil {
			result.Format()
			results = append(results, *result)
		}
	}
	return results
}

// ChunkedResult carries a group of indexes where the terms found are near each other
type ChunkedResult struct {
	Indexes []int
}

// chunkSimilarResults avoids repeating part of two or more fragments if a searched term appears too frequently.
// These chunked results can be bigger than the rest, but the result is easier on the eyes.
func chunkSimilarResults(indexes []int) []ChunkedResult {
	chunks := []ChunkedResult{}
	currentChunk := ChunkedResult{}
	currentIndexValue := 0
	numberOfIndexes := len(indexes)
	for i := 0; i < numberOfIndexes; i++ {
		currentIndexValue = indexes[i]
		currentChunk.Indexes = append(currentChunk.Indexes, currentIndexValue)
		nextIndex := i + 1
		if nextIndex < numberOfIndexes && indexes[nextIndex]-currentIndexValue > PreSuffixSize {
			chunks = append(chunks, currentChunk)
			currentChunk = ChunkedResult{}
		}
	}
	chunks = append(chunks, currentChunk) //last chunk
	return chunks
}

// Format transforms each fragment in which a query term was found to make it prettier to see on the browser
func (r *Result) Format() {
	for i := 0; i < len(r.Fragments); i++ {
		firstLineIndex := strings.Index(r.Fragments[i], WindowsLineBreak)
		lastLineIndex := strings.LastIndex(r.Fragments[i], WindowsLineBreak)
		r.Fragments[i] = r.Fragments[i][firstLineIndex:lastLineIndex]                        //removes potentially broken lines
		r.Fragments[i] = strings.ReplaceAll(r.Fragments[i], WindowsLineBreak, HTMLLineBreak) //fix line breaks
		r.Fragments[i] = strings.Trim(r.Fragments[i], HTMLLineBreak)
	}
}
