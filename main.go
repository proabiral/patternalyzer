package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var joiner string

func logic2(domainPart1 []string, domainPart2 []string) (output string) {
	noMatchCount := 0

	for i := range domainPart1 {
		if i == len(domainPart1)-1 {
			joiner = "."
		} else {
			joiner = "-"
		}

		if domainPart1[i] == domainPart2[i] {
			output += domainPart1[i] + joiner
		} else {
			noMatchCount++
			output += "{replace_this}" + joiner
		}
	}

	if noMatchCount == 1 {
		return output
	} else {
		return ""
	}
}

func logic(domainMap map[int][][]string) map[string]bool {
	maxLength2process := 14
	patterns := make(map[string]bool)

	for i := 2; i <= maxLength2process; i++ {
		for j, domain := range domainMap[i] {
			for l, domain2 := range domainMap[i] {
				if j == l {
					continue
				}
				
				pattern := ""
				noMatchCount := 0

				for k := range domain {
					if domain[k] == domain2[k] {
						pattern += domain[k] + "."
					} else {
						noMatchCount++

						innerPattern := ""
						if strings.Contains(domain[k], "-") && strings.Contains(domain2[k], "-") {
							parts1 := strings.Split(domain[k], "-")
							parts2 := strings.Split(domain2[k], "-")
							if len(parts1) == len(parts2) {
								innerPattern = logic2(parts1, parts2)
							}
						}
						
						if innerPattern != "" {
							pattern += innerPattern
						} else {
							if k == 0 && !strings.Contains(domain[k], "-") {
								break
							}
							pattern += "{replace_this}."
						}
					}
				}

				if noMatchCount == 1 && pattern != "" {
					patterns[strings.TrimSuffix(pattern, ".")] = true
				}
			}
		}
	}

	return patterns
}

func readWordlist(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var wordlist []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		wordlist = append(wordlist, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return wordlist, nil
}

func hasExactRepetition(domain, word string) bool {
	// Check for word.word pattern
	if strings.Contains(domain, word+"."+word) {
		return true
	}
	
	// Check for word-word pattern
	if strings.Contains(domain, word+"-"+word) {
		return true
	}
	
	return false
}

func main() {
	wordlistFile := flag.String("w", "", "Path to the wordlist file")
	uniqueFlag := flag.Bool("u", false, "Prevent exact word repetitions in the generated domains")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("Usage: go run main.go [-w wordlist.txt] [-u] domains.txt")
	}

	domainMap := make(map[int][][]string)

	file, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := scanner.Text()
		splittedDomain := strings.Split(domain, ".")
		domainLength := len(splittedDomain)
		domainMap[domainLength] = append(domainMap[domainLength], splittedDomain)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	var wordlist []string
	if *wordlistFile != "" {
		wordlist, err = readWordlist(*wordlistFile)
		if err != nil {
			log.Fatalf("Error reading wordlist: %v", err)
		}
	}

	patterns := logic(domainMap)

	for pattern := range patterns {
		if *wordlistFile != "" {
			for _, word := range wordlist {
				// Create the potential new domain first
				newDomain := strings.Replace(pattern, "{replace_this}", word, 1)
				
				// Check for exact repetitions when -u flag is set
				if *uniqueFlag && hasExactRepetition(newDomain, word) {
					continue
				}
				
				fmt.Println(newDomain)
			}
		} else {
			fmt.Println(pattern)
		}
	}
}
