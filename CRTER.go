package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
)

const banner = `

  _____ _____ _______ ______ _____  
 / ____|  __ \__   __|  ____|  __ \ 
| |    | |__) | | |  | |__  | |__) |
| |    |  _  /  | |  |  __| |  _  / 
| |____| | \ \  | |  | |____| | \ \ 
 \_____|_|  \_\ |_|  |______|_|  \_\
                                    

`

func fetchCrtShDomains(domain string) ([]string, error) {
	url := fmt.Sprintf("https://crt.sh/?q=%%25.%s", domain)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`<TD>\*\.([^<]+)</TD>`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	var domains []string
	for _, match := range matches {
		domains = append(domains, match[1])
	}

	return domains, nil
}

func main() {
	fmt.Println(banner)

	if len(os.Args) != 3 {
		fmt.Printf("Usage: %s <domainlist> <output>\n", os.Args[0])
		os.Exit(1)
	}

	domainlist := os.Args[1]
	output := os.Args[2]

	file, err := os.Open(domainlist)
	if err != nil {
		fmt.Printf("Error opening domainlist: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	outFile, err := os.Create(output)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	writer := bufio.NewWriter(outFile)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := scanner.Text()
		domains, err := fetchCrtShDomains(domain)
		if err != nil {
			fmt.Printf("Error fetching domains for %s: %v\n", domain, err)
			continue
		}

		fmt.Printf("Domains for %s:\n", domain)
		for _, d := range domains {
			fmt.Println(d)
			writer.WriteString(d + "\n")
		}
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading domainlist: %v\n", err)
		os.Exit(1)
	}

	writer.Flush()
}
