package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

var joiner string

func logic2(domainPart1 []string, domainPart2 []string) (output string) {
	noMatchCount := 0

	for i, _ := range domainPart1 {

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

	// to only print pattern with one replace_this , like this jira.{replace_this}.dev.foobar.example.com
	// does not print pattern like this , jira.{replace_this}.{replace_this}.foobar.example.com
	if noMatchCount == 1 {
		return output
	} else {
		return ""
	}
}

func logic(domainMap map[int][][]string) {

	maxLength2process := 14

	for i := 2; i <= maxLength2process; i++ {
		//looping each domain of length i
		for j, domain := range domainMap[i] {
			for l, domain2 := range domainMap[i] {
				pattern := ""
				noMatchCount := 0
				// looping each word of domain
				for k, _ := range domain {
					// if domain and domain2 are not same
					if j != l {
						// if word on domain and domain2 is same
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
									//if no match on first word and number of dash on both not same, don't check other words
									if k == 0 {
										break
									}
								}
							} else {
								//if no match on first word and no - in words , don't check other words
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
				// to only print pattern with one replace_this , like this jira.{replace_this}.dev.foobar.example.com
				// does not print pattern like this , jira.{replace_this}.{replace_this}.foobar.example.com
				if noMatchCount == 1 && pattern != "" {
					fmt.Println(strings.TrimSuffix(pattern, "."))
				}
			}
		}
	}
}

func main() {

	domainMap := make(map[int][][]string)

	file, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		domain := scanner.Text()
		splittedDomain := strings.Split(domain, ".")
		domainLength := len(splittedDomain)
		domainMap[domainLength] = append(domainMap[domainLength], splittedDomain)

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	logic(domainMap)
}
