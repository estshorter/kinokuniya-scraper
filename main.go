package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ktnyt/go-moji"
)

var rPrice = regexp.MustCompile(`（本体¥(.*)）`)

// Exists checks if file/directory exists
func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func scrapeEach(url, cacheFilePath string, ch chan<- *Book) {
	book := new(Book)
	defer func() { ch <- book }()
	if strings.HasPrefix(url, "https://www.kinokuniya.co.jp") == false {
		book.Err = fmt.Errorf("Error: wrong url '%v'", url)
		return
	}
	urlSplitted := strings.Split(url, "/")
	if len(urlSplitted) < 1 {
		book.Err = fmt.Errorf("Error: cannot split by '/', '%v'", url)
		return
	}
	bookID := urlSplitted[len(urlSplitted)-1]
	if strings.HasPrefix(bookID, "dsg") == false {
		book.Err = fmt.Errorf("Error: wrong bookId '%v'", bookID)
		return
	}
	html, err := loadHTML(bookID, url, cacheFilePath)
	if err != nil {
		book.Err = err
		return
	}
	if err := parseHTML(html, book); err != nil {
		book.Err = err
		return
	}
}

func loadHTML(bookID, url, cacheFilePath string) (io.Reader, error) {
	cache := filepath.Join(cacheFilePath, bookID+".html")
	if Exists(cache) {
		return loadHTMLFromFile(cache)
	}
	return loadHTMLFromWeb(url, cache)
}

func parseHTML(html io.Reader, book *Book) error {
	doc, err := goquery.NewDocumentFromReader(html)
	if err != nil {
		return err
	}
	titleRaw := doc.Find("h3").First().Text()
	book.Title = sanitize(titleRaw)

	infobox := doc.Find(".infobox").First()
	li := infobox.Find("li")

	liA := li.First().Find("a")
	authors := make([]string, liA.Size())
	li.First().Find("a").Each(func(idx int, s *goquery.Selection) {
		authors[idx] = sanitize(s.Text())
	})
	book.Author = strings.Join(authors, "、")

	if writtenInJapanese(doc) {
		parseJapaneseBook(li, book)
	} else {
		parseNonJapaneseBook(doc, book)
	}

	isbnRaw, exists := doc.Find(`li[itemprop="identifier"]`).Attr("content")
	if exists == false {
		return fmt.Errorf("Error: cannot parse isbn")
	}
	book.Isbn = isbnRaw[5:]
	return nil
}

func writtenInJapanese(doc *goquery.Document) bool {
	if strings.Contains(doc.Find("ul.pankuzu").Text(), "和書") {
		return true
	}
	return false
}

func parseJapaneseBook(s *goquery.Selection, book *Book) {
	sNext2 := s.Next().Next()
	book.Price = parsePrice(sNext2.First())
	book.Publisher = parsePublisher(sNext2.Next().First())
}

func parseNonJapaneseBook(doc *goquery.Document, book *Book) {
	doc.Find(".pricebox").Each(func(idx int, s *goquery.Selection) {
		li := s.Find("li")
		price := parsePrice(li.First())
		if idx == 0 || price < book.Price {
			book.Price = price
		}
		if idx == 0 {
			book.Publisher = parsePublisher(li.Next().First())
		}
	})
}

func parsePrice(s *goquery.Selection) int {
	match := rPrice.FindStringSubmatch(s.Text())
	if len(match) != 2 {
		log.Fatal(fmt.Errorf("Error: pattern match error for book price"))
	}
	priceStr := strings.ReplaceAll(match[1], ",", "")
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		log.Fatal(err)
	}
	return price
}

func parsePublisher(s *goquery.Selection) string {
	return strings.TrimSpace(s.Children().Text())
}

func sanitize(str string) string {
	s1 := strings.TrimSpace(str)
	s2 := moji.Convert(s1, moji.ZE, moji.HE)
	s3 := moji.Convert(s2, moji.ZS, moji.HS)
	return s3
}

func loadHTMLFromFile(cache string) (io.Reader, error) {
	content, err := ioutil.ReadFile(cache)
	return bytes.NewReader(content), err
}

func loadHTMLFromWeb(url, cache string) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	ioutil.WriteFile(cache, body, 0644)
	return bytes.NewReader(body), nil
}

// Book defines type of book in kinokuniya
type Book struct {
	Title     string
	Author    string
	Price     int
	Publisher string
	Isbn      string
	Err       error
}

func (book Book) String() string {
	return fmt.Sprintf("%v\n%v\n%v\n%v\nISBN: %v",
		book.Title, book.Author, book.Price, book.Publisher, book.Isbn)
}

func toString(books []Book) string {
	booksStr := make([]string, len(books))
	for idx, book := range books {
		booksStr[idx] = book.String()
	}
	return strings.Join(booksStr, "\n\n")
}

func save(books []Book, filepath string) error {
	return ioutil.WriteFile(filepath, []byte(toString(books)), 0644)
}

func scrape(urls []string, cacheFilePath string) ([]Book, error) {
	if !Exists(cacheFilePath) {
		os.MkdirAll(cacheFilePath, os.ModePerm)
	}
	ch := make(chan *Book, len(urls))
	for _, url := range urls {
		go scrapeEach(url, cacheFilePath, ch)
	}

	books := make([]Book, len(urls))
	for idx := range urls {
		books[idx] = *(<-ch)
		if books[idx].Err != nil {
			return nil, books[idx].Err
		}
	}
	return books, nil
}

func main() {
	urls := os.Args[1:]
	if len(os.Args) < 2 {
		return
	}

	const cacheFilePath = ".cache"
	books, err := scrape(urls, cacheFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Printf("%v", toString(books))
	if err := save(books, "./book_info.txt"); err != nil {
		log.Fatalln(err)
	}
}
