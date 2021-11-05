package main

/* TODO:
 * Change "Favorite" field to "Fav"
 * Update places where getting user input to use the getline function
 * Update places where articles are printed to use the toString() method
 */

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// Article holds article information
type Article struct {
	Title    string   `json:"title"`
	URL      string   `json:"url"`
	Tags     []string `json:"tags"`
	Read     bool     `json:"read"`
	Favorite bool     `json:"favorite"`
}

func (a *Article) toString(short bool) string {
	res := a.Title
	res += fmt.Sprintf("\n\t%s", a.URL)
	if short {
		return res
	}
	if len(a.Tags) != 0 {
		res += fmt.Sprintf("\n\t%s", strings.Join(a.Tags, "|"))
	}
	if a.Read {
		res += "\n\tRead"
	} else {
		res += "\n\tUnread"
	}
	if a.Favorite {
		res += "\n\tFavorite"
	}
	return res
}

func (a *Article) String() string {
	return a.toString(true)
}

var (
	titleExpr = regexp.MustCompile(`<title.*?>(.*)</title>`)
	// For cleaning downloaded HTML
	scriptExpr = regexp.MustCompile(`<script.*?>.*?</script>`)
	bodyExpr   = regexp.MustCompile(`<body.*?>`)
)

func (a *Article) download() error {
	// Get the article HTML from the url
	resp, err := http.Get(a.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// Parse the response body html
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}
	// Get the title
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			a.Title = n.FirstChild.Data
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)
	/*
	  // Read the response body
	  data, err := ioutil.ReadAll(resp.Body)
	  if err != nil {
	    return err
	  }
	  // Get the title
	  if matches := titleExpr.FindSubmatch(data); len(matches) != 0 {
	    a.Title = string(matches[1])
	  }
	  // Replace move the scripts
	  data = scriptExpr.ReplaceAll(data, []byte{})
	*/
	return nil
}

var (
	articles []*Article
	input    = bufio.NewReader(os.Stdin)
)

func main() {
	log.SetFlags(log.Lshortfile)

	// Print the options out
	fmt.Println("OPTIONS")
	fmt.Println("1) New Article")
	fmt.Println("2) Print Articles")
	fmt.Println("3) Edit Article")
	fmt.Println("4) Delete Article")
	fmt.Println("5) Cancel")
	// Get the user's choice
	var (
		err    error
		choice = -1
	)
	for err != nil || choice < 1 || choice > 5 {
		fmt.Print("Choice: ")
		line, err := input.ReadString('\n')
		if err == nil {
			line = strings.TrimSpace(line)
			choice, err = strconv.Atoi(line)
		}
	}
	if choice == 5 {
		return
	}
	fmt.Println()

	// Open the file
	fileName := path.Join(os.Getenv("HOME"), ".articles.json")
	f, err := os.OpenFile(
		fileName,
		os.O_CREATE|os.O_RDONLY,
		0755,
	)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	// Decode the file
	if err := json.NewDecoder(f).Decode(&articles); err != nil && err.Error() != "EOF" {
		log.Println(err)
		os.Exit(1)
	}
	f.Close()
	// Defer writing the (possibly) updated articles to the file
	defer func() {
		// Write the articles to the file
		if data, err := json.Marshal(articles); err != nil {
			log.Println(err)
			os.Exit(1)
		} else if err := ioutil.WriteFile(fileName, data, 0755); err != nil {
			log.Println(err)
			os.Exit(1)
		}
	}()
	// Perform the appropraite action
	switch choice {
	case 1:
		newArticle()
	case 2:
		printArticles()
	case 3:
		editArticle()
	case 4:
		deleteArticle()
	}
}

func newArticle() {
	var article = new(Article)
	// Get the article fields
	getURL(article)
	getTags(article)
	getRead(article)
	getFav(article)
	// Try to get the title from the URL
	if err := article.download(); err != nil {
		log.Printf("error getting article title: %v", err)
		// Get the title from the user
		getTitle(article)
	}
	// See if there are any desired changes
	if !getChanges(article) {
		return
	}

	// Add the article to the list of articles
	articles = append(articles, article)

	fmt.Print("Another (Y/n): ")
	line, _ := input.ReadString('\n')
	if line = strings.ToLower(strings.TrimSpace(line)); line == "yes" || line == "y" {
		newArticle()
	}
}

func printArticles() {
	if choice := strings.ToLower(getline("Add queries (Y/n): ")); choice != "y" && choice != "yes" {
		displayArticles(true)
		return
	}
	/* TODO: Querying titles, urls, read, fav, and tags */
	/* TODO: Add option for AND and OR querying */
	var qArticle = new(Article)
	// These variables tell if a query field has been set (given) or not
	var setTitle, setURL, setTags, setRead, setFav bool
QueryLoop:
	for {
		// Print the query field
		const NA = "[Not Set]"
		fmt.Println("Query Fields")
		fmt.Print("1) Title: ")
		if setTitle {
			fmt.Println(qArticle.Title)
		} else {
			fmt.Println(NA)
		}
		fmt.Print("2) URL: ")
		if setURL {
			fmt.Println(qArticle.URL)
		} else {
			fmt.Println(NA)
		}
		fmt.Print("3) Tags: ")
		if setTags {
			fmt.Println(strings.Join(qArticle.Tags, "|"))
		} else {
			fmt.Println(NA)
		}
		fmt.Print("4) Read: ")
		if setRead {
			fmt.Println(qArticle.Read)
		} else {
			fmt.Println(NA)
		}
		fmt.Print("5) Favorite: ")
		if setFav {
			fmt.Println(qArticle.Favorite)
		} else {
			fmt.Println(NA)
		}
		fmt.Printf("1) Title: %s\n", qArticle.Title)
		fmt.Printf("2) URL: %s\n", qArticle.URL)
		fmt.Printf("3) Tags: %s\n", strings.Join(qArticle.Tags, "|"))
		fmt.Printf("4) Read: %v\n", qArticle.Read)
		fmt.Printf("5) Favorite: %v\n", qArticle.Favorite)
		fmt.Print("Choice (6 = cancel, negative to unset, anything else to finish): ")
		// Get the choice and perform the appropriate action
		line, _ := input.ReadString('\n')
		line = strings.TrimSpace(line)
		choice, err := strconv.Atoi(line)
		if err != nil {
			break
		}
		switch choice {
		case 1:
			getTitle(qArticle)
			setTitle = qArticle.Title != ""
		case 2:
			getURL(qArticle)
			setURL = qArticle.URL != ""
		case 3:
			getTags(qArticle)
			setTags = len(qArticle.Tags) != 0
		case 4:
			getRead(qArticle)
			setRead = true
		case 5:
			getFav(qArticle)
			setFav = true
		case -1:
			setTitle = false
		case -2:
			setURL = false
		case -3:
			setTags = false
		case -4:
			setRead = false
		case -5:
			setFav = false
		case 6:
			return
		default:
			break QueryLoop
		}
	}
	// Make the query article title lowercase for better searching
	qArticle.Title = strings.ToLower(qArticle.Title)
	for i, article := range articles {
		/* TODO: Use regexp for title, url, and tags? fields */
		if setTitle {
			if !strings.Contains(article.Title, qArticle.Title) {
				continue
			}
		}
		if setURL {
			if !strings.Contains(article.URL, qArticle.URL) {
				continue
			}
		}
		if setTags {
			var match bool
			for _, tag := range qArticle.Tags {
				// Require exact matches (may change)
				// Short-circcuit (may change if all matching tags required)
				if sort.SearchStrings(article.Tags, tag) != -1 {
					match = true
					break
				}
			}
			if !match {
				continue
			}
		}
		if setRead {
			if article.Read != qArticle.Read {
				continue
			}
		}
		if setFav {
			if article.Favorite != qArticle.Favorite {
				continue
			}
		}
		fmt.Printf("%d) %s\n", i+1, article)
	}
}

func editArticle() {
	// Get the length of the articles array
	l := len(articles)
	fmt.Println("Choose article")
	// Print the short version of the articles
	displayArticles(false)
	if l == 0 {
		return
	}
	// Get the desired article
	var (
		article *Article
		choice  int = -1
	)
	for choice < 0 || choice >= l {
		fmt.Print("Choice: ")
		line, _ := input.ReadString('\n')
		line = strings.TrimSpace(line)
		choice, _ = strconv.Atoi(line)
		choice--
	}
	article = &*(articles[choice])

	// Get the changes
	if !getChanges(article) {
		return
	}

	// Switch out the old article with the new changed one
	articles[choice] = article

	fmt.Print("Another (Y/n): ")
	line, _ := input.ReadString('\n')
	if line = strings.ToLower(strings.TrimSpace(line)); line == "yes" || line == "y" {
		editArticle()
	}
}

func deleteArticle() {
	// Get the length of the articles
	l := len(articles)
	// Print the short version of the articles
	fmt.Println("Choose article")
	displayArticles(false)
	if l == 0 {
		return
	}
	// Get the choice
	var (
		article *Article
		choice  int = -1
	)
	for choice < 0 || choice >= l {
		fmt.Print("Choice: ")
		line, _ := input.ReadString('\n')
		line = strings.TrimSpace(line)
		choice, _ = strconv.Atoi(line)
		choice--
	}
	article = articles[choice]

	// Check to make sure it's the article the user wants to delete
	fmt.Printf("%s\n\t%s\n", article.Title, article.URL)
	fmt.Print("Delete this article (Y/n): ")
	line, _ := input.ReadString('\n')
	if line = strings.ToLower(strings.TrimSpace(line)); line == "yes" || line == "y" {
		articles = append(articles[:choice], articles[choice+1:]...)
		fmt.Print("Another (Y/n): ")
		line, _ := input.ReadString('\n')
		if line = strings.ToLower(strings.TrimSpace(line)); line == "yes" || line == "y" {
			deleteArticle()
		}
	}
}

// Prompts the user and gets a line from stdin (whitespace trimmed)
func getline(prompt string) string {
	fmt.Print(prompt)
	line, _ := input.ReadString('\n')
	return strings.TrimSpace(line)
}

func displayArticles(full bool) {
	if len(articles) == 0 {
		fmt.Println("No articles")
		return
	}
	// Print the articles
	for i, article := range articles {
		fmt.Printf("%d) %s\n", i+1, article.Title)
		fmt.Printf("\t%s\n", article.URL)
		if full {
			fmt.Printf("\t%s\n", strings.Join(article.Tags, "|"))
			if article.Read {
				fmt.Println("\tRead")
			} else {
				fmt.Println("\tUnread")
			}
			if article.Favorite {
				fmt.Println("\tFav")
			}
		}
	}
}

func getTitle(article *Article) {
	fmt.Print("Title: ")
	line, _ := input.ReadString('\n')
	article.Title = strings.TrimSpace(line)
}

func getURL(article *Article) {
	fmt.Print("URL: ")
	line, _ := input.ReadString('\n')
	article.URL = strings.TrimSpace(line)
}

func getTags(article *Article) {
	fmt.Print(`Tags (separated by "|" with no space): `)
	line, _ := input.ReadString('\n')
	article.Tags = strings.Split(strings.ToLower(strings.TrimSpace(line)), "|")
	sort.Strings(article.Tags)
}

func getRead(article *Article) {
	fmt.Print("Read (Y/n): ")
	line, _ := input.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	article.Read = (line == "yes" || line == "y")
}

func getFav(article *Article) {
	fmt.Print("Favorite (Y/n): ")
	line, _ := input.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	article.Favorite = (line == "yes" || line == "y")
}

// Returns false if the user canceled
func getChanges(article *Article) bool {
	for {
		fmt.Println("Change Any?")
		fmt.Printf("1) Title: %s\n", article.Title)
		fmt.Printf("2) URL: %s\n", article.URL)
		fmt.Printf("3) Tags: %s\n", strings.Join(article.Tags, "|"))
		fmt.Printf("4) Read: %v\n", article.Read)
		fmt.Printf("5) Favorite: %v\n", article.Favorite)
		fmt.Print("Choice (6 = cancel, anything else to finish): ")
		line, _ := input.ReadString('\n')
		line = strings.TrimSpace(line)
		choice, err := strconv.Atoi(line)
		if err != nil {
			return true
		}
		switch choice {
		case 1:
			getTitle(article)
		case 2:
			getURL(article)
		case 3:
			getTags(article)
		case 4:
			getRead(article)
		case 5:
			getFav(article)
		case 6:
			return false
		default:
			return true
		}
	}
}
