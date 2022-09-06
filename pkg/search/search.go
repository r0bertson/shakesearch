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

type WorkSearcher struct {
	Works []Work
}

type Work struct {
	Title       string
	Text        string
	SuffixArray *suffixarray.Index
}

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

type Result struct {
	WorkTitle string   `json:"work_title"`
	Fragments []string `json:"fragments"`
	//TODO: fragments could be verse-based
}

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

type ChunkedResult struct {
	Indexes []int
}

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

func (r *Result) Format() {
	for i := 0; i < len(r.Fragments); i++ {
		firstLineIndex := strings.Index(r.Fragments[i], WindowsLineBreak)
		lastLineIndex := strings.LastIndex(r.Fragments[i], WindowsLineBreak)
		r.Fragments[i] = r.Fragments[i][firstLineIndex:lastLineIndex]                        //removes potentially broken lines
		r.Fragments[i] = strings.ReplaceAll(r.Fragments[i], WindowsLineBreak, HTMLLineBreak) //fix line breaks
	}
}
