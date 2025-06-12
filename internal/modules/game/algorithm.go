package game

import "strings"

var Worddictionary = map[string]bool{}

const (
	WrongLetter = "wrong-letter"
	NoSuchWord  = "no-such-word"
	WordReused  = "word-reused"
	WordCorrect = "word-correct"
)

func MatchWord(word string, wordslist map[string]bool, startletter byte) string {
	word = strings.TrimSpace(word)

	if word == "" || word[0] != startletter {
		return WrongLetter
	}
	if !Worddictionary[word] {
		return NoSuchWord
	}
	if wordslist[word] { // word has been re used
		return WordReused
	}
	return WordCorrect
}
