package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

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
							//if no match on first word don't check other words
							if k == 0 {
								break
							}
							pattern += "{replace_this}."
						}
					}
				}
				// to only print pattern with one replace_this , like this jira.{replace_this}.dev.foobar.example.com
				// does not print pattern like this , jira.{replace_this}.{replace_this}.foobar.example.com
				if noMatchCount == 1 && pattern != "" {
					fmt.Println(pattern)
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
		// splitting domains with dot and arranging them according to length.
		// split with dash(-) and underscope(_) in future ?
		splittedDomain := strings.Split(domain, ".")
		domainLength := len(splittedDomain)
		domainMap[domainLength] = append(domainMap[domainLength], splittedDomain)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	logic(domainMap)

}
