package textproc

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var ErrEmptyWords = errors.New("words is empty")

func AnalyzeWordsFrequency(data []string) (*TextAnalysis, error) {
	if data == nil {
		return nil, ErrEmptyWords
	}
	analysis, err := NewTextAnalysis()
	if err != nil {
		return nil, fmt.Errorf("AnalyzeWordFrequency: %w", err)
	}
	for _, word := range data {
		analysis.IncWordCount(word)
	}

	return analysis, nil
}

// TextAnalysis represents a struct which contains WordFrequency field and a Name field
// of this analysis.
type TextAnalysis struct {
	ID            string         `json:"id"`
	WordFrequency map[string]int `json:"wordFrequency"`

	mu sync.Mutex
}

// Creates a new TextAnalysis.
func NewTextAnalysis() (*TextAnalysis, error) {
	id, err := NewAnalysisID()
	if err != nil {
		return nil, fmt.Errorf("new id: %w", err)
	}

	return &TextAnalysis{
		ID:            id,
		WordFrequency: make(map[string]int),
	}, nil
}

// Adds new occurrence of a word.
// Goroutine safe.
func (ta *TextAnalysis) IncWordCount(word string) {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	if ta.WordFrequency == nil {
		ta.WordFrequency = make(map[string]int)
	}

	ta.WordFrequency[word]++
}

var defaultMaxInt int64 = 10000

func randomInt(x int64) (*big.Int, error) {
	if x == 0 {
		x = defaultMaxInt
	}
	i := big.NewInt(x)
	v, err := rand.Int(rand.Reader, i)
	if err != nil {
		return nil, fmt.Errorf("random int: %w", err)
	}

	return v, nil
}

// Returns a string of format:
// text_analysis_randomnumber_currentdate.
func NewAnalysisID() (string, error) {
	layout := "02_01_2006_15_04"

	now := time.Now().UTC()
	date := now.Format(layout)

	i, err := randomInt(10000)
	if err != nil {
		return "", fmt.Errorf("failed to get random int: %w", err)
	}
	b := new(strings.Builder)
	b.WriteString("analysis")
	b.WriteString("_")
	b.WriteString(date)
	b.WriteString("_")
	b.WriteString(i.String())

	return b.String(), nil
}

func NewAnalysisIDWithSuffix(suffix string) (string, error) {
	id, err := NewAnalysisID()
	if err != nil {
		return "", fmt.Errorf("generating analysis id: %w", err)
	}

	b := new(strings.Builder)
	b.WriteString(strings.Trim(suffix, " "))
	b.WriteString("_")
	b.WriteString(id)

	return b.String(), nil
}
