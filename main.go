package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

const (
	TAB       = 9
	BACKSPACE = 8
	DELETE    = 127
	ESCAPE    = 27
	CTRL_C    = 3
)

var (
	autoCompleteTriggered bool     // to keep track of keypresses after the autocomplete feature is triggered
	suggestions           []string // list of suggestions for current word
	suggestionIndex       int      // index to track currently displayed suggestion
)

// The core data structure
type Trie struct {
	children  map[rune]*Trie
	wordCount int
}

// Descibes a word and how many times its been used
type Word struct {
	value string
	count int
}

// To sort suggestions based on usage
type Suggestions []Word

func (m Suggestions) Len() int           { return len(m) }
func (m Suggestions) Less(i, j int) bool { return m[i].count > m[j].count }
func (m Suggestions) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

func TrieConstructor() *Trie {
	return &Trie{
		children:  make(map[rune]*Trie),
		wordCount: 0,
	}
}

// Insert word into the Trie
func (root *Trie) Insert(word string) {
	for _, s := range word {
		if root.children[s] == nil {
			root.children[s] = TrieConstructor()
		}
		root = root.children[s]
	}
	root.wordCount++
}

// Returns list of suggestions for auto-completion. The suggestions are sorted in order of usage
func (root *Trie) Autofill(word string) []string {
	var output Suggestions
	var result []string

	if len(word) == 0 {
		return result
	}

	for _, s := range word {
		if root.children[s] == nil {
			return result
		} else {
			root = root.children[s]
		}
	}

	dfs(root, "", &output)
	sort.Sort(output)

	result = make([]string, len(output))
	for i, word := range output {
		result[i] = word.value
	}

	return result
}

func dfs(root *Trie, prefix string, output *Suggestions) {
	if root.wordCount > 0 {
		*output = append(*output, Word{prefix, root.wordCount})
	}

	for k, v := range root.children {
		dfs(v, prefix+string(k), output)
	}
}

func main() {
	// Enable raw mode to capture keypresses instantly - from stack overflow
	oldState, err := term.MakeRaw(int(syscall.Stdin))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer term.Restore(int(syscall.Stdin), oldState)

	data, err := os.ReadFile("words.txt")
	if err != nil {
		fmt.Println("ReadFile failed:", err)
	}

	// Convert the file content to a string and split it into words
	content := string(data)
	words := strings.Fields(content) // Splits on spaces, newlines, and tabs ( better than strings.Split(content, " "))

	ch := make(chan string, 1000)
	// Goroutine to render text on terminal
	go render(ch)

	var input []rune             // Store input characters
	inputChan := make(chan byte) // Channel for keypresses
	trie := TrieConstructor()

	// Insert all words from "words.txt"
	for _, word := range words {
		trie.Insert(word)
	}

	// Goroutine to read input
	go inputReader(inputChan)

	timer := time.NewTimer(200 * time.Millisecond) // timer to trigger autocomplete suggestions

	ctx, cancel := context.WithCancel(context.TODO())
	fmt.Println("START TYPING")
	for {
		select {
		case <-timer.C:
			// get current word being typed
			word := getCurrentWord(input)
			suggestions = trie.Autofill(word)
			if len(suggestions) == 0 {
				continue
			}
			autoCompleteTriggered = true
			cancel()

			ctx, cancel = context.WithCancel(context.TODO())
			go recommendation(ctx, suggestions[suggestionIndex%len(suggestions)], input, ch)

		case key, ok := <-inputChan:
			if !ok {
				return // Exit if input channel is closed
			}

			// Reset timer on each keypress
			timer.Reset(200 * time.Millisecond)

			// Key press detected while autocomplete suggestion is displayed66
			if autoCompleteTriggered {
				cancel()
				if key == TAB { // Loop through suggestions
					suggestionIndex++
					// ctx, cancel = context.WithTimeout(context.TODO(), 10*time.Second)
					ctx, cancel = context.WithCancel(context.TODO())
					go recommendation(ctx, suggestions[suggestionIndex%len(suggestions)], input, ch)
					continue
				} else if key == '\n' || key == '\r' { // Suggestion has been selected. Perform autocomplete
					input = append(input, []rune(suggestions[suggestionIndex%len(suggestions)])...)
					key = ' '
				}

				autoCompleteTriggered = false
				suggestions = []string{}
				suggestionIndex = 0
			}

			// Ignore TAB and Enter -> to simplify getCurrentWord() and getLastWord() logic
			if key == TAB || key == '\n' || key == '\r' {
				continue
			}

			// On detecting SPACE, store the last typed word into the Trie
			if key == ' ' {
				word := getLastWord(input)
				trie.Insert(word)
			}

			// Handle backspace
			if key == BACKSPACE || key == DELETE {
				if len(input) > 0 {
					input = input[:len(input)-1]
					fmt.Print("\b \b")
					ch <- string(input)
				}
				continue
			}

			// Add character and send to render() function
			input = append(input, rune(key))
			ch <- string(input)
		}
	}
}

// Goroutine which sends input + suggestion to render() with a blinking effect
func recommendation(ctx context.Context, r string, input []rune, inputchan chan string) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	alt := []string{string(input) + r, string(input)}

	for i := 0; ; i++ {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			inputchan <- alt[i%2]
		}
	}
}

// To get the current word being typed
// Eg:- this is a tes  --> getCurrentWord() returns tes
func getCurrentWord(input []rune) string {
	var str []rune
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] == ' ' {
			break
		} else {
			str = append(append([]rune{}, input[i]), str...)
		}
	}
	return string(str)
}

// To get the previous word that was typed
// Eg:- this is a test  --> getLastWord() returns test
func getLastWord(input []rune) string {
	var str []rune
	var wordEncountered bool
	for i := len(input) - 1; i >= 0; i-- {
		if input[i] == ' ' && wordEncountered {
			break
		} else if input[i] != ' ' {
			str = append(append([]rune{}, input[i]), str...)
			wordEncountered = true
		}
	}
	return string(str)
}

// Read keypressed and sends it to main loop
func inputReader(inputChan chan byte) {
	var b [1]byte
	for {
		_, err := os.Stdin.Read(b[:])
		if err != nil {
			fmt.Println("\nError reading input:", err)
			close(inputChan)
			return
		}
		// Exit program on Ctrl+C or Esc
		if b[0] == ESCAPE || b[0] == CTRL_C {
			close(inputChan)
			return
		}
		inputChan <- b[0]
	}
}

// Render function
func render(in <-chan string) {
	for str := range in {
		fmt.Print("\033[H\033[2J") // Clear screen
		fmt.Print(str)
		time.Sleep(50 * time.Millisecond)
	}
}
