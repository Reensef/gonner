package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	ResetColor   = "\033[0m"
	RedColor     = "\033[31m"
	GreenColor   = "\033[32m"
	YellowColor  = "\033[33m"
	BlueColor    = "\033[34m"
	MagentaColor = "\033[35m"
	CyanColor    = "\033[36m"
	GrayColor    = "\033[37m"
	WhiteColor   = "\033[97m"
)

func main() {
	executeFlag := flag.String("execute", "./main.go", "Path to execute go file")
	flag.StringVar(executeFlag, "e", "", "Alias for --execute")

	testsFlag := flag.String("tests", "./data", "Path to tests dir")
	flag.StringVar(testsFlag, "t", "", "Alias for --tests")

	dataSuffixFlag := flag.String("data-suffix", "", "Siffix for tests data files")
	flag.StringVar(testsFlag, "d", "", "Alias for --data-suffix")

	answerSuffixFlag := flag.String("answer-suffix", ".a", "Siffix for tests answers files")
	flag.StringVar(testsFlag, "a", "", "Alias for --answer-suffix")

	flag.Parse()

	executeFilePath := *executeFlag
	testsDirPath := *testsFlag
	dataSuffix := *dataSuffixFlag
	answerSuffix := *answerSuffixFlag

	files, err := os.ReadDir(testsDirPath)
	if err != nil {
		fmt.Println("Error reading folder with tests:", err)
	}

	baseTestFileNameSet := make(map[string]bool)
	baseTestFileNames := make([]string, 0)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if path.Ext(file.Name()) == dataSuffix {
			baseName := strings.TrimSuffix(file.Name(), dataSuffix)
			baseTestFileNameSet[baseName] = true
		}
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if path.Ext(file.Name()) == answerSuffix {
			baseName := strings.TrimSuffix(file.Name(), answerSuffix)

			if !baseTestFileNameSet[baseName] {
				fmt.Println("No answers for test", baseName)
				continue
			}

			baseTestFileNames = append(baseTestFileNames, baseName)
		}
	}

	clear(baseTestFileNameSet)

	sort.Slice(baseTestFileNames, func(i, j int) bool {
		num1, err1 := strconv.Atoi(baseTestFileNames[i])
		num2, err2 := strconv.Atoi(baseTestFileNames[j])
		if err1 == nil && err2 == nil {
			return num1 < num2
		}

		return baseTestFileNames[i] < baseTestFileNames[j]
	})

	countCompleted := 0
	totalCount := len(baseTestFileNames)

	for _, baseFileName := range baseTestFileNames {
		fmt.Print("\r")
		fmt.Printf("In progress [%d/%d]", countCompleted, totalCount)
		inputFile, err := os.Open(testsDirPath + "/" + baseFileName + dataSuffix)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer inputFile.Close()

		actualOutReader, err := runProgram(executeFilePath, inputFile)
		if err != nil {
			fmt.Println(err)
			return
		}

		expectedOutFile, err := os.Open(testsDirPath + "/" + baseFileName + answerSuffix)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer expectedOutFile.Close()
		actualScanner := bufio.NewScanner(actualOutReader)
		expectedScanner := bufio.NewScanner(expectedOutFile)

		result, lineError := compareLineByLine(actualScanner, expectedScanner)
		if !result {
			fmt.Print("\r                            \r")
			fmt.Print(RedColor)
			fmt.Println("Error on line", lineError, "in test file", baseFileName)
			fmt.Println("Actual:")
			fmt.Println(actualScanner.Text())
			fmt.Println("Expected:")
			fmt.Println(expectedScanner.Text())
			fmt.Print(ResetColor)
			return
		}
		countCompleted++
	}

	fmt.Print("\r                            \r")

	fmt.Print(GreenColor)
	fmt.Println("All tests passed")
	fmt.Print(ResetColor)
}

func runProgram(executeFilePath string, in io.Reader) (out *bytes.Buffer, err error) {
	out = bytes.NewBuffer(nil)

	cmd := exec.Command("go", "run", executeFilePath)

	cmd.Stdin = in
	cmd.Stdout = out
	cmd.Stderr = out

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return out, nil
}

func compareLineByLine(scannerActual, scannerExpected *bufio.Scanner) (result bool, errorLine int) {
	if scannerActual == nil || scannerExpected == nil {
		return false, 0
	}

	lineNumber := 1
	match := true

	for scannerActual.Scan() {
		if !scannerExpected.Scan() {
			return false, lineNumber
		}

		actualLine := scannerActual.Text()
		expectedLine := scannerExpected.Text()
		if actualLine != expectedLine {
			return false, lineNumber
		}

		lineNumber++
	}

	for scannerExpected.Scan() {
		return false, lineNumber
	}

	return match, 0
}
