package game

var Worddictionary = map[string]bool{}

func MatchWord(word string, wordslist map[string]bool, startletter byte) string {
	if word[0] != startletter {
		return "wrong-letter"
	}
	if _, ok := Worddictionary[word]; !ok {
		return "no-such-word"
	}
	if _, ok := wordslist[word]; ok { // word has been re used
		return "word-reused"
	}
	return "word-correct"
}
