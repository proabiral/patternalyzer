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

	// Only return patterns with exactly one {replace_this}
	if noMatchCount == 1 {
		return output
	} else {
		return ""
	}
}

func logic(domainMap map[int][][]string, wordlist []string) map[string]bool {
	maxLength2process := 14
	patterns := make(map[string]bool) // Use a map to deduplicate patterns

	for i := 2; i <= maxLength2process; i++ {
		for j, domain := range domainMap[i] {
			for l, domain2 := range domainMap[i] {
				pattern := ""
				noMatchCount := 0

				for k := range domain {
					if j != l {
						if domain[k] == domain2[k] {
							pattern += domain[k] + "."
						} else {
							noMatchCount++

							innerPattern := ""
							if strings.Contains(domain[k], "-") && strings.Contains(domain2[k], "-") {
								parts1 := strings.Split(domain[k], "-")
								parts2 := strings.Split(domain2[k], "-")
								partLength1 := len(parts1)
								partLength2 := len(parts2)
								if partLength1 == partLength2 {
									innerPattern = logic2(parts1, parts2)
								} else {
									if k == 0 {
										break
									}
								}
							} else {
								if k == 0 {
									break
								}
							}
							if innerPattern != "" {
								pattern += innerPattern
							} else {
								pattern += "{replace_this}."
							}
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

func main() {
	// Define command-line flags
	wordlistFile := flag.String("w", "", "Path to the wordlist file")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("Usage: go run main.go [-w wordlist.txt] domains.txt")
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

	patterns := logic(domainMap, wordlist)

	// Print results
	for pattern := range patterns {
		if *wordlistFile != "" {
			// Replace {replace_this} with words from the wordlist
			for _, word := range wordlist {
				fmt.Println(strings.Replace(pattern, "{replace_this}", word, 1))
			}
		} else {
			// Print the pattern as is
			fmt.Println(pattern)
		}
	}
}
