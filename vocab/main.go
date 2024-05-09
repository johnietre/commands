package main

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Word is a vocab word
type Word struct {
	Word     string
	Mastered bool
}

type Definition struct {
	Word       string
	Definition string
}

var (
	db       *sql.DB
	dbPath   string
	inReader = bufio.NewReader(os.Stdin)
)

func openDB(filePath string) *sql.DB {
	// Check if the database exists; if not, create it and run the script
	exists := true
	if _, err := os.Stat(dbPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Fatal(err)
		}
		exists = false
	}
	// Open the database
	database, err := sql.Open("sqlite3", filePath)
	if err != nil {
		log.Fatal(err)
	}
	// Turn on foreign key support
	if _, err := database.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		log.Fatal(err)
	}
	// If the database didn't exist, run the script to create the tables
	if !exists {
		_, thisDir, _, ok := runtime.Caller(0)
		if !ok {
			log.Fatal("error locating sql script")
		}
		scriptBytes, err := ioutil.ReadFile(path.Join(path.Dir(thisDir), "script.sql"))
		if err != nil {
			log.Fatal(err)
		}
		if _, err := database.Exec(string(scriptBytes)); err != nil {
			log.Fatal(err)
		}
	}
	return database
}

func init() {
	log.SetFlags(0)
	// Get the "JCMDS_PATH" environment variable and construct the database path
	// from it
	dbPath = os.Getenv("JCMDS_PATH")
	if dbPath == "" {
		log.Fatal(`"JCMDS_PATH" environment variable not set`)
	}
	dbPath = path.Join(dbPath, "vocab", "vocab.db")
	db = openDB(dbPath)
}

func main() {
	args := os.Args
	if len(args) == 1 {
		printHelp()
	}
	switch args[1] {
	case "add":
		commandAdd(args[2:])
	case "list":
		commandList(args[2:])
	case "quiz":
		commandQuiz(args[2:])
	case "master":
		commandMaster(args[2:])
	case "unmaster":
		commandUnmaster(args[2:])
	case "delete":
		commandDelete(args[2:])
	case "clear":
		commandClear(args[2:])
	case "reset":
		commandReset(args[2:])
	case "help":
		printHelp()
	default:
		log.Printf("unknown command: %s", strings.Join(args[1:], " "))
		printHelp()
	}
	db.Close()
}

// Adds a definition to a word, adding the word if necessary
func commandAdd(args []string) {
	// Get the word and definition from the user
	var word, def string
	if len(args) == 0 {
		word = getWord()
		def = getDefinition()
	} else if len(args) == 1 {
		word = args[0]
		def = getDefinition()
	} else if len(args) == 2 {
		word = args[0]
		def = args[1]
	} else {
		log.Println("invalid command:", strings.Join(args, " "))
		printHelp()
	}
	// Add the word and definition to the database
	if _, err := db.Exec(`INSERT OR IGNORE INTO words VALUES (?, ?)`, word, 0); err != nil {
		log.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO definitions VALUES (?, ?)`, word, def); err != nil {
		log.Fatal(err)
	}
}

// Prints all words or all definitions for a word and whether the word has been
// mastered or not
func commandList(args []string) {
	if len(args) == 0 {
		// Get the words from the database and print them out
		rows, err := db.Query(`SELECT * FROM words ORDER BY word`)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		word, mastered, empty := "", 0, true
		for rows.Next() {
			empty = false
			if err := rows.Scan(&word, &mastered); err != nil {
				log.Fatal(err)
			}
			if mastered == 1 {
				word += " ✓"
			}
			fmt.Println(word)
		}
		if empty {
			log.Fatal("no words")
		}
	} else if len(args) == 1 {
		// Get whether the given word is mastered or not, if it exists
		word := args[0]
		row := db.QueryRow(`SELECT mastered FROM words WHERE word=?`, word)
		mastered := 0
		if err := row.Scan(&mastered); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				log.Fatal("word doesn't exist")
			}
			log.Fatal(err)
		}
		// Get the definitions of the word and print the out
		rows, err := db.Query(`SELECT definition FROM definitions WHERE word=?`, word)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		if mastered == 1 {
			word += " ✓"
		}
		fmt.Println(word)
		def := ""
		for rows.Next() {
			if err := rows.Scan(&def); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("    %s\n", def)
		}
	} else {
		log.Println("invalid command:", strings.Join(args, " "))
		printHelp()
	}
}

// Picks a random word, a random definition for that word, and three other
// random words and random definitions for each of them and presents the four
// definitions to the user
func commandQuiz(args []string) {
	if len(args) != 0 {
		log.Println("invalid command:", strings.Join(args, " "))
		printHelp()
	}
	// Set the random seed
	rand.Seed(time.Now().Unix())
	// Get the random word which is to be the correct choice
	// Mastered words are ignored since they are mastered (duh)
	rows, err := db.Query(`SELECT word FROM words WHERE mastered=0`)
	if err != nil {
		log.Fatal(err)
	}
	words := make([]string, 0)
	for rows.Next() {
		word := ""
		if err := rows.Scan(&word); err != nil {
			log.Fatal(err)
		}
		words = append(words, word)
	}
	rows.Close()
	l := len(words)
	if l == 0 {
		log.Fatal("no words")
	}
	randomWord := words[rand.Intn(l)]
	// Get a random definition from the available definitions of the word
	rows, err = db.Query(`SELECT definition FROM definitions WHERE word=?`, randomWord)
	if err != nil {
		log.Fatal(err)
	}
	wordDefs := make([]string, 0)
	for rows.Next() {
		def := ""
		if err := rows.Scan(&def); err != nil {
			log.Fatal(err)
		}
		wordDefs = append(wordDefs, def)
	}
	rows.Close()
	l = len(wordDefs)
	// This should always be false, so long as the database isn't tampered with
	// outside of the program
	if l == 0 {
		log.Fatal("no definitions")
	}
	correctDef := wordDefs[rand.Intn(l)]
	// Get the other three random choices from the available definitions
	rows, err = db.Query(`SELECT * FROM definitions WHERE NOT word=?`, randomWord)
	if err != nil {
		log.Fatal(err)
	}
	wordsAndDefs := make(map[string][]string)
	for rows.Next() {
		word, def := "", ""
		if err := rows.Scan(&word, &def); err != nil {
			log.Fatal(err)
		}
		wordsAndDefs[word] = append(wordsAndDefs[word], def)
	}
	rows.Close()
	// Pick the three other choices (or less if there aren't enough for three)
	choices, n := [][2]string{[2]string{randomWord, correctDef}}, 1
	for word, defs := range wordsAndDefs {
		def := defs[rand.Intn(len(defs))]
		choices = append(choices, [2]string{word, def})
		n++
		if n == 4 {
			break
		}
	}
	// Shuffle the choices
	rand.Shuffle(n, func(i, j int) {
		choices[i], choices[j] = choices[j], choices[i]
	})
	// Print out the options and also find the index of the correct choice
	correctIndex := 0
	fmt.Printf("%s:\n", randomWord)
	for i, pair := range choices {
		if pair[0] == randomWord {
			correctIndex = i
		}
		fmt.Printf("%d) %s\n", i+1, pair[1])
	}
	// Get the choice from the user and print out whether it was correct or not
	// as well as print the correct choice, regardless of whether the user was
	// correct or not
	// The possible error is irrelevant because the default returned integer 0 is
	// an invalid choice and would be handled the same as a non-nil error
	choice, _ := strconv.Atoi(getLine("Choice: "))
	choice--
	if choice > 3 || choice < 0 {
		fmt.Println("invalid choice")
	} else if choice != correctIndex {
		fmt.Printf("INCORRECT, that definition is for: %s\n", choices[choice][0])
		fmt.Printf("The answer is: %d) %s\n", correctIndex+1, correctDef)
	} else {
		fmt.Println("CORRECT!!!")
	}
}

// Masters a word (sets a word to mastered)
func commandMaster(args []string) {
	// Get the word from the user
	word := ""
	if len(args) == 0 {
		word = getWord()
	} else if len(args) == 1 {
		word = args[0]
	} else {
		log.Println("invalid command:", strings.Join(args, " "))
		printHelp()
	}
	// Update the word to being mastered
	if res, err := db.Exec(`UPDATE words SET mastered=1 WHERE word=?`, word); err != nil {
		log.Fatal(err)
	} else if n, err := res.RowsAffected(); err != nil {
		log.Fatal(err)
	} else if n == 0 {
		log.Fatalf(`no word "%s"`, word)
	}
}

// "Unmasters" a word (sets a word's mastered status to false)
func commandUnmaster(args []string) {
	// Get the word from the user
	word := ""
	if len(args) == 0 {
		word = getWord()
	} else if len(args) == 1 {
		word = args[0]
	} else {
		log.Println("invalid command:", strings.Join(args, " "))
		printHelp()
	}
	// Update the word's mastered status to false
	if res, err := db.Exec(`UPDATE words SET mastered=0 WHERE word=?`, word); err != nil {
		log.Fatal(err)
	} else if n, err := res.RowsAffected(); err != nil {
		log.Fatal(err)
	} else if n == 0 {
		log.Fatalf(`no word "%s"`, word)
	}
}

// Deletes a definition of a word or a word and all its definitions
// If the last definition of a word is deleted, the word is deleted as well
func commandDelete(args []string) {
	// Get the word from the user
	var word, def string
	if len(args) == 0 {
		word = getWord()
	} else if len(args) == 1 {
		word = args[0]
	} else if len(args) == 2 {
		word = args[0]
		def = args[1]
	} else {
		log.Println("invalid command:", strings.Join(args, " "))
		printHelp()
	}
	if def == "" {
		// If desired, list the definitions out for the user
		choice := getLine("Delete the whole word and all its definitions? [Y/n] ")
		if strings.ToLower(choice) == "y" {
			// Delete the word and all its definitions
			if _, err := db.Exec(`DELETE FROM words WHERE word=?`, word); err != nil {
				log.Fatal(err)
			}
			return
		}
		// List the definitions of the word
		rows, err := db.Query(`SELECT definition FROM definitions WHERE word=?`, word)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		count := 0
		var defs []string
		for rows.Next() {
			count++
			if err := rows.Scan(&def); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("%d) %s\n", count, def)
			defs = append(defs, def)
		}
		if count == 0 {
			log.Fatalf(`no word "%s"`, word)
		}
		fmt.Println("0) Delete all definitions")
		fmt.Println("-1) Cancel")
		// Get the user's choice
		n, err := strconv.Atoi(getLine("Choice: "))
		// Here, there error must be checked since 0 (the default 'n' returned) is
		// a valid option
		if err != nil || n > count || n < -1 {
			log.Fatal("invalid choice")
		}
		if n == -1 {
			return
		} else if n == 0 {
			if _, err := db.Exec(`DELETE FROM words WHERE word=?`, word); err != nil {
				log.Fatal(err)
			}
			return
		}
		// Delete the definition
		if _, err := db.Exec(`DELETE FROM definitions WHERE definition=?`, defs[n-1]); err != nil {
			log.Fatal(err)
		}
		// Check to see if all definitions for the word have been deleted
		// If so, delete the word from the words table
		row := db.QueryRow(`SELECT definition FROM definitions WHERE word=?`, word)
		if err := row.Scan(&def); err != nil && errors.Is(err, sql.ErrNoRows) {
			if _, err := db.Exec(`DELETE FROM words WHERE word=?`, word); err != nil {
				log.Fatal(err)
			}
		}
	}
}

// Clears all words and definitions from the database
func commandClear(args []string) {
	if len(args) != 0 {
		log.Println("invalid command:", strings.Join(args, " "))
		printHelp()
	}
	// Get confirmation
	choice := getLine("Clear all words and definitions? [Y/n] ")
	if strings.ToLower(choice) == "y" {
		if _, err := db.Exec(`DELETE FROM words`); err != nil {
			log.Fatal(err)
		}
	}
}

// Clears all words and definitions from the database
func commandReset(args []string) {
	if len(args) != 0 {
		log.Println("invalid command:", strings.Join(args, " "))
		printHelp()
	}
	// Get confirmation
	choice := getLine("Reset all data? [Y/n] ")
	if strings.ToLower(choice) == "y" {
		db.Close()
		if err := os.Remove(dbPath); err != nil {
			log.Fatal(err)
		}
		db = openDB(dbPath)
	}
}

// Prints the help screen
func printHelp() {
	format := "    %-28s%s\n"
	fmt.Println("Vocab is a program for quizzing vocabulary words")
	fmt.Println("\t\tUsage: vocab [command]")
	fmt.Printf(format, "add [word [definition]]", "Add a definition")
	fmt.Printf(format, "list [word]", "List all words or definitions of a word")
	fmt.Printf(format, "quiz", "Quiz on a random word")
	fmt.Printf(format, "master [word]", "Mark a word as mastered (won't be show up in quizzes)")
	fmt.Printf(format, "unmaster [word]", `Mark a word as "unmastered"`)
	fmt.Printf(format, "delete [word [definition]]", "Delete a word or definition")
	fmt.Printf(format, "clear", "Delete all words")
	fmt.Printf(format, "reset", "Resets all program data")
	fmt.Printf(format, "help", "Print help")
	os.Exit(0)
}

// Gets a word, exitting if the word is empty
func getWord() string {
	word := getLine("Word: ")
	if word == "" {
		log.Fatal("must input word")
	}
	return word
}

// Gets a definition, exitting if the definition is empty
func getDefinition() string {
	def := getLine("Definition: ")
	if def == "" {
		log.Fatal("must input definition")
	}
	return def
}

// Prints out the prompt, if given, and returns a line of input from the user
// without the newline
func getLine(prompt string) string {
	if prompt != "" {
		fmt.Print(prompt)
	}
	line, err := inReader.ReadString('\n')
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(line)
}
