package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	_ "github.com/lib/pq"
)

var database *sql.DB

type shaMessage struct {
	sha string
}

var gitRepoPath string
var gitRepoURL string
var workers sync.WaitGroup
var workQueue chan shaMessage

func isCommitRow(entry string) bool {
	res, _ := regexp.MatchString(`^(\d\d\d\d-\d\d-\d\d)--`, entry)
	return res
}

func processCommitEntry(entry, sha string) {
	allLines := strings.Split(entry, "\n")
	if len(allLines) <= 1 {
		// Merge commits will have no numstats on them - the last
		// commit will contain the info
		return
	}

	var shaDateAuthorStr string
	var commitRows []string

	shaDateAuthorStr = allLines[0]
	commitRows = allLines[1:]

	shaDateAuthor := strings.Split(shaDateAuthorStr, "--")

	date := shaDateAuthor[0]
	author := shaDateAuthor[1]

	for _, commitRow := range commitRows {
		fileRowContents := strings.Split(commitRow, "\t")

		addedRows := fileRowContents[0]
		removedRows := fileRowContents[1]
		fileName := fileRowContents[2]

		// Binary files are stat:ed as - - (0 added 0 removed)
		if removedRows == "-" {
			removedRows = "0"
		}

		if addedRows == "-" {
			addedRows = "0"
		}

		if removedRows == "" || fileName == "" {
			panic(commitRow)
		}

		insertIntoLogStmt := `INSERT INTO log (
			sha,
			date,
			author,
			added,
			removed,
			filename,
			gitrepo
		) VALUES (
			$1,$2, $3, $4, $5, $6, $7
		) ON CONFLICT DO NOTHING;`

		if _, err := database.Exec(insertIntoLogStmt, sha, date, author, addedRows, removedRows, fileName, gitRepoURL); err != nil {
			panic(err)
		}
	}
}

func worker() {
	for shaMessage := range workQueue {
		commitMessageStr := gitLogCommand("--format=%b -n 1 " + shaMessage.sha)

		commitMessageTrimmed := strings.TrimSpace(commitMessageStr)
		commitMessageFlat := regexp.MustCompile("\n+").ReplaceAllString(commitMessageTrimmed, " ")
		commitMessageWordCount := len(strings.Split(commitMessageFlat, " "))

		if commitMessageFlat != "" {
			insertMessageStmt := `INSERT INTO message (
				sha,
				message,
				length,
				gitrepo
			) VALUES (
				$1,$2, $3, $4
			) ON CONFLICT DO NOTHING;`

			if _, err := database.Exec(insertMessageStmt, shaMessage.sha, commitMessageFlat, commitMessageWordCount, gitRepoURL); err != nil {
				panic(err)
			}
		}

		// Commit lines
		commitAddedRemovedLines := gitLogCommand("--numstat --pretty=format:%aI--%aN -n 1 " + shaMessage.sha)
		commitEntry := strings.TrimSpace(commitAddedRemovedLines)
		if commitEntry == "" {
			continue
		}

		processCommitEntry(commitEntry, shaMessage.sha)
	}

	workers.Done()
}

func gitLogCommand(argumentsStr string) string {
	gitLogArugments := strings.Split("--git-dir "+gitRepoPath+".git log "+argumentsStr, " ")
	var cmdOutput bytes.Buffer
	cmd := exec.Command("git", gitLogArugments...)
	cmd.Stdout = &cmdOutput

	err := cmd.Run()
	if err != nil {
		panic(err)
	}

	return cmdOutput.String()
}

func fireUpWorkers() {
	numberOfWorkers := runtime.NumCPU()
	workQueue = make(chan shaMessage, numberOfWorkers)

	workers = sync.WaitGroup{}

	for i := 0; i < numberOfWorkers; i++ {
		go worker()
		workers.Add(1)
	}
}

func main() {
	var err error
	pgDsn := os.Getenv("PG_DSN")
	if pgDsn == "" {
		pgDsn = "postgres://postgres@localhost:5432/spad-mats?sslmode=disable"
	}

	var sampleThreshold int
	countStr := os.Getenv("COUNT")
	if countStr != "" {
		sampleThreshold, err = strconv.Atoi(countStr)
		if err != nil {
			panic(err)
		}
	} else {
		sampleThreshold = int(math.MaxInt64)
	}

	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "Usage: go run cmd/import/main.go <path-to-git-repo> <git-repo-name>")
		os.Exit(1)
	}

	// !! TODO !! Add string magic for . .. and ~
	gitRepoPath = os.Args[1]
	gitRepoURL = os.Args[2]

	database, err = sql.Open("postgres", pgDsn)
	if err != nil {
		panic(err)
	}

	allShasStr := gitLogCommand("--no-renames --no-merges --format=%h")
	allShas := strings.Split(allShasStr, "\n")

	fireUpWorkers()

	if len(allShas) > sampleThreshold {
		addFactor := float64(len(allShas)) / float64(sampleThreshold)
		for i := float64(0); int(math.Floor(i)) < len(allShas); i = i + addFactor {
			shaIndex := int(math.Floor(i))

			sha := allShas[shaIndex]
			if sha == "" {
				continue
			}

			workQueue <- shaMessage{sha: sha}
		}
	} else {
		for _, sha := range allShas {
			if sha == "" {
				continue
			}

			workQueue <- shaMessage{sha: sha}
		}
	}

	close(workQueue)
	workers.Wait()
}
