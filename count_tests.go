package main

import "bufio"
import "fmt"
import "os"
import "path/filepath"
import "regexp"
import "strings"

type FileData struct {
	shortName string
	path string
	numTests int
	deepTests int
	extended []string
}

func (fileData *FileData) setDeepTests(shortNameToFileData map[string]*FileData) int {
	if fileData.deepTests >= 0 {
		return fileData.deepTests
	}
	fileData.deepTests = 0
	totalTests := fileData.numTests
	for i := range fileData.extended {
		extended := fileData.extended[i]
		parentFileData := shortNameToFileData[extended]
		if parentFileData != nil {
			totalTests += parentFileData.setDeepTests(shortNameToFileData)
		}
	}
	fileData.deepTests = totalTests
	return totalTests
}

func main() {
	sourceCodeSuffixes := []string {".java", ".scala"}

	shortNameToFileData := make(map[string]*FileData)

	err := filepath.Walk(".", func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		for i := range sourceCodeSuffixes {
			suffix := sourceCodeSuffixes[i]
			if strings.HasSuffix(path, suffix) {
				noSuffix := path[:len(path) - len(suffix)]
				//fmt.Printf("path='%s', len(suffix)='%d', noSuffix='%s'\n", path, len(suffix), noSuffix)
				lastIndex := strings.LastIndex(noSuffix, "/")
				if lastIndex >= 0 {
					noSuffix = noSuffix[lastIndex+1:]
				}
				shortNameToFileData[noSuffix] = &FileData {
					shortName: noSuffix,
					path: path,
					numTests: -1,
					deepTests: -1,
					extended: make([]string, 0),
				}
				break
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	re := regexp.MustCompile(`class ([a-zA-Z0-9_+-]*).*extends ([a-zA-Z0-9_+-]*)[^a-zA-Z0-9_+-]`)
	for _, fileData := range(shortNameToFileData) {
		file, err2 := os.Open(fileData.path)
		if err2 != nil {
			panic(err2)
		}
		defer file.Close()
		scanner := bufio.NewScanner(file)
		numTests := 0
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "@Test") {
				numTests++
			}
			match := re.FindStringSubmatch(line)
			if match != nil {
				childClassName := match[1]
				parentClassName := match[2]
				if childClassName == fileData.shortName {
					fileData.extended = append(fileData.extended, parentClassName)
				}
			}
		}
		if err := scanner.Err(); err != nil {
			panic(err)
		}
		fileData.numTests = numTests
	}

	for shortName, fileData := range(shortNameToFileData) {
		fileData.setDeepTests(shortNameToFileData)
		if fileData.deepTests == 0 {
			delete(shortNameToFileData, shortName)
		}
	}

	for _, fileData := range(shortNameToFileData) {
		fmt.Printf("shortName=%s, numTests=%d, deepTests=%d, extended=%s\n",
				fileData.shortName, fileData.numTests, fileData.deepTests, strings.Join(fileData.extended, ","))
	}

	totalTests := 0
	totalDeepTests := 0
	numFiles := 0
	for _, fileData := range(shortNameToFileData) {
		totalTests += fileData.numTests
		totalDeepTests += fileData.deepTests
		numFiles++
	}
	fmt.Printf("%d total tests, %d deep tests in %d source code files\n", totalTests, totalDeepTests, numFiles)
}
