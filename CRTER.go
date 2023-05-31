package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

const banner = `

  _____ _____ _______ ______ _____  
 / ____|  __ \__   __|  ____|  __ \ 
| |    | |__) | | |  | |__  | |__) |
| |    |  _  /  | |  |  __| |  _  / 
| |____| | \ \  | |  | |____| | \ \ 
 \_____|_|  \_\ |_|  |______|_|  \_\
                                    
				By Micro0x00

`

func fetchCrtShDomains(domain string, wg *sync.WaitGroup, ch chan<- []string, errCh chan<- error) {
	defer wg.Done()

	var url string
	if strings.Contains(domain, ".") {
		url = fmt.Sprintf("https://crt.sh/?q=%%25.%s", domain)
	} else {
		url = fmt.Sprintf("https://crt.sh/?q=%%25.%s%%25", domain)
	}

	resp, err := http.Get(url)
	if err != nil {
		errCh <- err
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errCh <- err
		return
	}

	re := regexp.MustCompile(`<TD>\*\.([^<]+)</TD>`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	var domains []string
	for _, match := range matches {
		domains = append(domains, match[1])
	}

	ch <- domains
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
	var wg sync.WaitGroup
	ch := make(chan []string)
	errCh := make(chan error)

	for scanner.Scan() {
		domain := scanner.Text()
		wg.Add(1)
		go fetchCrtShDomains(domain, &wg, ch, errCh)
	}

	go func() {
		wg.Wait()
		close(ch)
		close(errCh)
	}()

	for domains := range ch {
		for _, d := range domains {
			fmt.Println(d)
			writer.WriteString(d + "\n")
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading domainlist: %v\n", err)
		os.Exit(1)
	}

	for err := range errCh {
		fmt.Printf("Error fetching domains: %v\n", err)
	}

	writer.Flush()
}
