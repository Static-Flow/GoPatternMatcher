package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)
/**
This project performs quick pattern matches against HTTP responses. Simply pipe in a list of URLs and the pattern you'd
like to match and this tool will show you the results.
 */
func main() {
	start := time.Now()
	pattern := flag.String("pattern", "", "Pattern definition to look for")
	findAll := flag.Bool("findall",false, "Find all matches not just first one")
	workers := flag.Int("workers", 20, "Number of workers to process urls")
	timeout := flag.Int("timeout", 10000, "timeout in milliseconds")
	flag.Parse()
	if len(*pattern) == 0 {
		log.Fatalln("Please supply a pattern to search for")
	}
	//Thanks to @tomnomnom for the following design pattern for WaitGroups from his awesome HTTProbe project https://github.com/tomnomnom/httprobe
	timeoutDuration := time.Duration(*timeout * 1000000)

	var transport = &http.Transport{
		MaxIdleConns:      30,
		IdleConnTimeout:   time.Second,
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   timeoutDuration,
			KeepAlive: time.Second,
		}).DialContext,
	}

	client := &http.Client{
		Transport:     transport,
		CheckRedirect: nil,
		Timeout:       timeoutDuration,
	}

	var wg sync.WaitGroup
	urlsToSearch := make(chan string)
	totalMatches := 0
	fmt.Printf("searching for %s\n", *pattern)
	fmt.Printf("Creating %d worker(s)\n", *workers)
	for i := 0; i < *workers; i++ {
		wg.Add(1)
		/*
		This following goroutine is where the magic happens
		It pulls URLs from the group, sends a GET request, then runs the supplied regular expression against the response
		Results are printed to screen along with a final tally of the total found
		 */
		go func() {
			for urlToSearch := range urlsToSearch {
				resp, err := client.Get(urlToSearch)
				if err == nil {
					body, _ := ioutil.ReadAll(resp.Body)
					//All patterns are surrounded by .* so you don't have to
					re := regexp.MustCompile(`.*` + *pattern + `.*`)
					if *findAll {
						matches := re.FindAllString(string(body), -1)
						for _, match := range matches {
							fmt.Printf(strings.Trim(match, " ") + "\n")
						}
						totalMatches += len(matches)
					} else {
						match := re.FindString(string(body))
						if len(match) > 0 {
							fmt.Printf(strings.Trim(match, " ") + "\n")
							totalMatches += 1
						}
					}
				}
				resp.Body.Close()
				fmt.Printf("Search of %s complete\n\n",urlToSearch)
			}
			wg.Done()
		}()
	}
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		urlsToSearch <- s.Text()
	}
	close(urlsToSearch)
	fmt.Println("All URLs scheduled")
	wg.Wait()
	fmt.Printf("Found %d matches for %s\n",totalMatches,*pattern)
	fmt.Printf("Finished in: %s\n", time.Since(start))
}