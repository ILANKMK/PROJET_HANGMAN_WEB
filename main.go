package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type GameState struct {
	Word           string
	GuessedWord    string
	Guesses        []string
	RemainingTries int
	Message        string
	GameOver       bool
	Won            bool
	HangmanArt     string
	Level          string
	Score          int
}

var (
	easyWords     = []string{"CHAT", "CHIEN", "MAISON", "ARBRE", "LIVRE"}
	mediumWords   = []string{"GOLANG", "PYTHON", "JAVASCRIPT", "COMPUTER", "INTERNET"}
	hardWords     = []string{"DÉVELOPPEMENT", "ARCHITECTURE", "ALGORITHME", "FRAMEWORK", "AUTHENTICATION"}
	game          GameState
	hangmanStates = []string{
		`
  +---+
  |   |
      |
      |
      |
      |
=========`,
		`
  +---+
  |   |
  O   |
      |
      |
      |
=========`,
		`
  +---+
  |   |
  O   |
  |   |
      |
      |
=========`,
		`
  +---+
  |   |
  O   |
 /|   |
      |
      |
=========`,
		`
  +---+
  |   |
  O   |
 /|\  |
      |
      |
=========`,
		`
  +---+
  |   |
  O   |
 /|\  |
 /    |
      |
=========`,
		`
  +---+
  |   |
  O   |
 /|\  |
 / \  |
      |
=========`,
	}
)

func getWordsByLevel(level string) []string {
	switch level {
	case "easy":
		return easyWords
	case "medium":
		return mediumWords
	case "hard":
		return hardWords
	default:
		return mediumWords
	}
}

func getTriesByLevel(level string) int {
	switch level {
	case "easy":
		return 10
	case "medium":
		return 8
	case "hard":
		return 5
	default:
		return 10
	}
}

func initGame(level string) {
	rand.Seed(time.Now().UnixNano())
	words := getWordsByLevel(level)
	game.Word = words[rand.Intn(len(words))]
	game.GuessedWord = strings.Repeat("_", len(game.Word))
	game.Guesses = []string{}
	game.RemainingTries = getTriesByLevel(level)
	game.Message = "Bienvenue au niveau " + strings.ToUpper(level) + "!"
	game.GameOver = false
	game.Won = false
	game.HangmanArt = hangmanStates[0]
	game.Level = level
	game.Score = 0
}

func calculateScore() int {
	baseScore := map[string]int{
		"easy":   100,
		"medium": 200,
		"hard":   300,
	}

	return baseScore[game.Level] * game.RemainingTries
}

func makeGuess(guess string) {
	guess = strings.ToUpper(guess)

	if len(guess) > 1 {
		if guess == game.Word {
			game.GuessedWord = game.Word
			game.Won = true
			game.GameOver = true
			game.Score = calculateScore()
			game.Message = "Félicitations! Score: " + string(rune(game.Score))
		} else {
			game.RemainingTries--
			game.Message = "Ce n'est pas le bon mot!"
			game.HangmanArt = hangmanStates[len(hangmanStates)-1-game.RemainingTries]
			if game.RemainingTries == 0 {
				game.GameOver = true
				game.Message = "Perdu! Le mot était: " + game.Word
			}
		}
		return
	}

	if contains(game.Guesses, guess) {
		game.Message = "Vous avez déjà essayé cette lettre!"
		return
	}

	game.Guesses = append(game.Guesses, guess)

	if !strings.Contains(game.Word, guess) {
		game.RemainingTries--
		game.Message = "Mauvaise lettre!"
		game.HangmanArt = hangmanStates[len(hangmanStates)-1-game.RemainingTries]
		if game.RemainingTries == 0 {
			game.GameOver = true
			game.Message = "Perdu! Le mot était: " + game.Word
		}
		return
	}

	newGuessedWord := []rune(game.GuessedWord)
	for i, letter := range game.Word {
		if string(letter) == guess {
			newGuessedWord[i] = letter
		}
	}
	game.GuessedWord = string(newGuessedWord)
	game.Message = "Bonne lettre!"

	if game.GuessedWord == game.Word {
		game.Won = true
		game.GameOver = true
		game.Score = calculateScore()
		game.Message = "Félicitations! Score: " + string(rune(game.Score))
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func handleGuess(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Erreur lors de l'analyse du formulaire", http.StatusBadRequest)
			return
		}

		guess := r.FormValue("guess")
		if len(guess) > 0 && !game.GameOver {
			makeGuess(guess)
		}
	}

	tmpl, err := template.ParseFiles("templates/game.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, game)
}

func handleLevel(w http.ResponseWriter, r *http.Request) {
	level := r.URL.Query().Get("level")
	if level == "" {
		level = "medium"
	}
	initGame(level)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	initGame("medium")

	http.HandleFunc("/", handleGuess)
	http.HandleFunc("/level", handleLevel)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Serveur démarré sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
