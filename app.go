package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"sync"
)

type Detail struct {
	Title string
	Means []Mean
}
type Mean struct {
	Form        string
	Description string
	Synonyms    []Synonym
}

type Synonym struct {
	Style string
	Text  string
}

func DetailInsert() {
	db, err := sql.Open("mysql", "dbuser:1861dleae@/mydb")
	if err != nil {
		panic("Error opening DB:", err.Error())
	}
	defer db.Close()

	someParam := "value"

	res, err := db.Exec(`INSERT INTO foo VALUES("bar", ?))`, someParam)
	if err != nil {
		println("Exec err:", err.Error())
	} else {
		id, err := res.LastInsertId()
		if err != nil {
			println("LastInsertId:", id)
		} else {
			println("Error:", err.Error())
		}
	}

}

func GetUrls(url string) []string {
	url = "http://www.thesaurus.com" + url
	doc, _ := goquery.NewDocument(url)
	urls := []string{}
	doc.Find("div.result_list a").Each(func(_ int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		urls = append(urls, url)
	})
	return urls
}

func GetDetail(url string, wg *sync.WaitGroup, m *sync.Mutex) Detail {
	m.Lock()
	defer m.Unlock()

	url = "http://www.thesaurus.com" + url
	doc, _ := goquery.NewDocument(url)
	title := doc.Find("h1").Text()
	means := []Mean{}
	doc.Find("div.synonyms").Each(func(_ int, s *goquery.Selection) {
		form := s.Find("div.synonym-description em").Text()
		description := s.Find("div.synonym-description strong").Text()
		synonyms := []Synonym{}
		doc.Find("div.relevancy-list li a").Each(func(_ int, s *goquery.Selection) {
			style, _ := s.Attr("style")
			text := s.Find("span.text").Text()
			synonyms = append(synonyms, Synonym{style, text})
		})
		means = append(means, Mean{form, description, synonyms})
	})

	wg.Done()
	detail := Detail{title, means}
	fmt.Println(detail)
	return detail
}

func main() {
	wg := new(sync.WaitGroup)
	m := new(sync.Mutex)
	url := "/list/a"
	for _, url1 := range GetUrls(url) {
		fmt.Println("in for url1")
		for _, url2 := range GetUrls(url1) {
			wg.Add(1)
			go GetDetail(url2, wg, m)
		}
	}
	wg.Wait()
}
