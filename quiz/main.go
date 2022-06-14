package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

var problemsFilePath = flag.String("problems", "problems.csv", "problems csv file")
var timeout = flag.Duration("timeout", 30*time.Second, "time limit for all questions to be answered within")

func main() {
	flag.Parse()
	if err := runQuiz(*problemsFilePath, *timeout); err != nil {
		fmt.Println(fmt.Errorf("Error occurred: %s", err))
		os.Exit(1)
	}
}

func runQuiz(problemsFilePath string, timeout time.Duration) error {
	questions, err := readQuestions(problemsFilePath)
	if err != nil {
		return fmt.Errorf("read questions: %s", err)
	}

	_, err = input("Press enter to start")
	if err != nil {
		return fmt.Errorf("input: %s", err)
	}

	score, err := askQuestions(questions, timeout)
	if err != nil {
		return fmt.Errorf("ask questions: %s", err)
	}

	fmt.Printf("Score: %d / %d\n", score, len(questions))

	return nil
}

type question struct {
	question, answer string
}

func readQuestions(filePath string) ([]question, error) {
	f, err := os.Open(*problemsFilePath)
	if err != nil {
		return nil, fmt.Errorf("open problems file: %s", err)
	}

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse problems file: %s", err)
	}

	questions := make([]question, len(records))
	for i, record := range records {
		questions[i] = question{
			question: record[0],
			answer:   record[1],
		}
	}
	return questions, nil
}

func askQuestions(questions []question, timeLimit time.Duration) (int, error) {
	timer := time.NewTimer(timeLimit)
	score := 0
	for _, question := range questions {
		answerChan := make(chan string)
		errChan := make(chan error)
		go func() {
			answer, err := input(question.question)
			if err != nil {
				errChan <- err
				return
			}
			answerChan <- answer
		}()

		select {
		case <-timer.C:
			fmt.Printf("\nTimed out after %s\n", timeLimit)
			return score, nil
		case answer := <-answerChan:
			if answer == question.answer {
				score++
			}
		case err := <-errChan:
			return 0, fmt.Errorf("input: %s", err)
		}
	}
	return score, nil
}

func input(msg string) (string, error) {
	stdin := bufio.NewReader(os.Stdin)
	fmt.Printf("%s ", msg)
	input, err := stdin.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("read from stdin: %s", err)
	}
	return strings.TrimSpace(input), nil
}
