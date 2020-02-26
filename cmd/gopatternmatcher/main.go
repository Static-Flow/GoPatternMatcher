package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
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
	findAll := flag.Bool("findall", false, "Find all matches not just first one")
	workers := flag.Int("workers", 20, "Number of workers to process URLs")
	timeout := flag.Int("timeout", 10000, "timeout in milliseconds")
	context := flag.Int("context", 50, "Number of characters on both sides of a match to include. (0 to include whole line, could be large for minified JS)")
	path := flag.String("path", "", "Path to append to input URLs")
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
				resp, err := client.Get(urlToSearch + *path)
				if err == nil {
					s := bufio.NewScanner(resp.Body)
					re := regexp.MustCompile(*pattern)
					line := 1
					for s.Scan() {
						searchSpace := strings.TrimSpace(s.Text())
						match := re.FindStringIndex(searchSpace)
						if len(match) > 0 {
							if *context == 0 {
								fmt.Printf("Found match on line %d: %s\n", line, searchSpace)
							} else {
								leftSlice := match[0] - *context
								if leftSlice < 0 {
									leftSlice = 0
								}
								rightSlice := match[1] + *context
								if rightSlice > len(searchSpace) {
									rightSlice = len(searchSpace)
								}
								fmt.Printf("Found match on line %d at offset %d: %s\n", line, match[0], searchSpace[leftSlice:rightSlice])
								//fmt.Println(searchSpace[leftSlice:rightSlice])
							}
							totalMatches += 1
							if !*findAll {
								break
							}
						}
						line++
					}
				}
				resp.Body.Close()
				fmt.Printf("Search of %s complete\n\n", urlToSearch+*path)
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
	fmt.Printf("Found %d matches for %s\n", totalMatches, *pattern)
	fmt.Printf("Finished in: %s\n", time.Since(start))
}
