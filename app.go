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

func DetailInsert(detail Detail, db *DB) bool {

	fmt.Println("insert :", detail.Title)

	// insert to details
	res, err := db.Exec("INSERT INTO details (title) VALUES(?)", detail.Title)
	if err != nil {
		fmt.Println("Exec err:", err.Error())
		return false
	} else {
		detail_id, err := res.LastInsertId()
		if err != nil {
			fmt.Println("Error:", err.Error())
			return false
		}
	}

	// insert to means
	for _, mean := range detail.Means {
		res, err := db.Exec("INSERT INTO means (detail_id, form, description) VALUES(?, ?,?)", detail_id, mean.Form, mean.Description)
		if err != nil {
			fmt.Println("Exec err:", err.Error())
			return false
		} else {
			mean_id, err := res.LastInsertId()
			if err != nil {
				fmt.Println("Error:", err.Error())
				return false
			}
		}
		for _, synonym := range mean.Synonyms {
			res, err := db.Exec("INSERT INTO synonyms (detail_id, mean_id, style, text) VALUES(?, ?,?,?)", detail_id, mean_id, synonym.Style, synonym.Text)
			if err != nil {
				fmt.Println("Exec err:", err.Error())
				return false
			}

		}
	}
	return true

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

	detail := Detail{title, means}
	//DetailInsert(detail)
	fmt.Println("insert :", detail.Title)
	return detail
}

func main() {
	db, err := sql.Open("mysql", "dbuser:1861dleae@(mydbinstance.cfjiimohdkcd.ap-northeast-1.rds.amazonaws.com:3306)/mydb")
	if err != nil {
		panic("Error opening DB:" + err.Error())
	}
	defer db.Close()

	wg := new(sync.WaitGroup)
	m := new(sync.Mutex)
	url := "/list/a"
	for _, url1 := range GetUrls(url) {
		go func() {
			for _, detail_url := range GetUrls(url1) {
				go func() {
					wg.Add(1)
					detail := GetDetail(detail_url, wg, m)
					DetailInsert(detail, db)
					wg.Done()
				}()
			}
		}()
	}
	wg.Wait()
}
