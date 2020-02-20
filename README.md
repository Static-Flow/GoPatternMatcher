# GoPatternMatcher
This tool allows for quickly searching for a specified pattern within HTTP Response bodies. Simply pipe in a list of URLs, specify your pattern and hit enter.


# Help Output
```
Usage of gpm:
  -context int
        Number of characters on both sides of a match to include. (0 to include whole line, could be large for minified JS) (default 50)
  -findall
        Find all matches not just first one
  -pattern string
        Pattern definition to look for
  -timeout int
        timeout in milliseconds (default 10000)
  -workers int
        Number of workers to process urls (default 20)

```


# Examples

## Simple use case searching for username
cat urls.txt | gpm --pattern admin

## Specify --findall to search for all occurrences of .js files
cat urls.txt | gpm --patttern .js --findall

## Specify number of workers to concurrently process the URLs
cat urls.txt | gpm --pattern .js --findall --workers 30

## Specify timeout of web reuqests in milliseconds
cat urls.txt | gpm --pattern .js --findall --workers 30 --timeout 20000

## Specify number of characters on each side of match  to include
cat urls.txt | gpm --pattern .js --findall --workers 30 --timeout 20000 --context 100
