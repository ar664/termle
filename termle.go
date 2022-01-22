package main

import (
	"bufio"
	"embed"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	totalWordles = 2315
	blackSquare  = `⬛`
	yellowSquare = `🟨`
	greenSquare  = `🟩`
)

var (
	//go:embed guesses.txt
	//go:embed answers.txt
	files embed.FS

	dayFlag    = flag.Int("day", daysSinceFirstWordle(), "select a specific wordle by day")
	randomFlag = flag.Bool("random", false, "pick a random wordle")

	// UTC to avoid DST
	firstDay = time.Date(2021, time.June, 19, 0, 0, 0, 0, time.UTC)
	valid    = regexp.MustCompile(`^[A-Za-z]{5}$`)
)

type game struct {
	day            int
	currentTurn    int
	turnsRemaining int
	complete       bool
	won            bool
	answer         string
	validGuesses   map[string]struct{}
	public         [][]string
	private        [][]string
}

func main() {
	flag.Parse()

	var day int
	if *randomFlag {
		day = randomDay()
	} else {
		day = *dayFlag
	}

	g := newGame(day)
	s := bufio.NewScanner(os.Stdin)

	g.printTurn()
	for !g.complete && s.Scan() {
		guess := strings.ToUpper(strings.TrimSpace(s.Text()))
		if !valid.MatchString(guess) {
			g.printTurnWithError("Please enter a 5 letter word")
			continue
		}
		if _, ok := g.validGuesses[guess]; !ok {
			g.printTurnWithError("Not in word list")
			continue
		}
		g.addGuess(guess)
		g.printTurn()
	}
	if s.Err() != nil {
		panic(s.Err())
	}
	g.printScore()
}

func (g *game) printScore() {
	var turnS string
	if g.won {
		turnS = strconv.Itoa(g.currentTurn)
		fmt.Println("you won!")
	} else {
		turnS = "X"
		fmt.Println("you lose!")
		fmt.Println("Answer was", g.answer)
	}
	fmt.Printf("Wordle %v %v/6\n\n", g.day, turnS)
	for _, y := range g.private {
		for _, x := range y {
			fmt.Print(x)
		}
		fmt.Println()
	}
}

func green(l string) string {
	return "\033[37;102m" + l + "\033[0m"
}
func yellow(l string) string {
	return "\033[37;103m" + l + "\033[0m"
}
func white(l string) string {
	return "\033[0;107m" + l + "\033[0m"
}
func black(l string) string {
	return "\033[37;100m" + l + "\033[0m"
}
func clearBoard() {
	fmt.Print("\033c")
}
func prompt() {
	fmt.Print(">")
}

func newGame(day int) *game {
	b := make([][]string, 6)
	for i := range b {
		b[i] = make([]string, 5)
		for j := range b[i] {
			b[i][j] = black("_")
		}
	}
	p := make([][]string, 0, 6)
	return &game{
		day:            day,
		currentTurn:    0,
		turnsRemaining: 6,
		complete:       false,
		won:            false,
		answer:         answerForDay(day),
		validGuesses:   guessesSet(),
		public:         b,
		private:        p,
	}
}

func (g *game) addGuess(guess string) {
	private := make([]string, 5)
	for i, c := range guess {
		if c == rune(g.answer[i]) {
			g.public[g.currentTurn][i] = green(string(c))
			private[i] = greenSquare
		} else if strings.ContainsRune(g.answer, c) {
			g.public[g.currentTurn][i] = yellow(string(c))
			private[i] = yellowSquare
		} else {
			g.public[g.currentTurn][i] = black(string(c))
			private[i] = blackSquare
		}
	}
	g.private = append(g.private, private)
	g.turnsRemaining--
	g.currentTurn++
	if guess == g.answer || g.turnsRemaining == 0 {
		g.complete = true
	}
	if guess == g.answer {
		g.won = true
	}
}

func (g *game) print() {
	fmt.Printf("Wordle %v\n", g.day)
	for _, y := range g.public {
		for _, x := range y {
			fmt.Print(" " + x)
		}
		fmt.Println()
	}
}

func (g *game) printTurn() {
	clearBoard()
	g.print()
	prompt()
}

func (g *game) printTurnWithError(err string) {
	clearBoard()
	g.print()
	fmt.Println(err)
	prompt()
}

func guessesSet() map[string]struct{} {
	guessesFile, err := files.Open("guesses.txt")
	if err != nil {
		panic(err)
	}
	defer guessesFile.Close()
	validGuesses := make(map[string]struct{})
	guessesReader := bufio.NewScanner(guessesFile)

	for guessesReader.Scan() {
		validGuesses[strings.ToUpper(guessesReader.Text())] = struct{}{}
	}
	if guessesReader.Err() != nil {
		panic(err)
	}
	return validGuesses
}

func answerForDay(day int) string {
	answersFile, err := files.Open("answers.txt")
	if err != nil {
		panic(err)
	}
	defer answersFile.Close()
	seeker := answersFile.(io.ReadSeeker)
	_, err = seeker.Seek(int64(day*7), io.SeekStart)
	if err != nil {
		panic(err)
	}
	answer := make([]byte, 5)
	_, err = seeker.Read(answer)
	if err != nil {
		panic(err)
	}
	return strings.ToUpper(string(answer))
}

func daysSinceFirstWordle() int {
	year, month, day := time.Now().Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return int(today.Sub(firstDay).Hours() / 24)
}

func randomDay() int {
	return rand.Intn(totalWordles)
}
