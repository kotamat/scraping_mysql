package main

import (
	"database/sql"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
	"strconv"
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

func DetailInsert(detail Detail, db *sql.DB) bool {

	fmt.Println("insert :", detail.Title)

	var detail_id, mean_id int64

	// insert to details
	res, err := db.Exec("INSERT INTO details (title) VALUES(?)", detail.Title)
	if err != nil {
		fmt.Println("Exec err:", err.Error())
		return false
	} else {
		id, err := res.LastInsertId()
		if err != nil {
			fmt.Println("Error:", err.Error())
			return false
		}
		detail_id = id
	}

	// insert to means
	for _, mean := range detail.Means {
		res, err := db.Exec("INSERT INTO means (detail_id, form, description) VALUES(?, ?,?)", detail_id, mean.Form, mean.Description)
		if err != nil {
			fmt.Println("Exec err:", err.Error())
			return false
		} else {
			id, err := res.LastInsertId()
			if err != nil {
				fmt.Println("Error:", err.Error())
				return false
			}
			mean_id = id
		}
		for _, synonym := range mean.Synonyms {
			_, err := db.Exec("INSERT INTO synonyms (detail_id, mean_id, style, text) VALUES(?, ?,?,?)", detail_id, mean_id, synonym.Style, synonym.Text)
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
	base_url := "/list/"
	for _, abc := range "abcdefghijklmnopqrstuvwxyz" {
		for i := 1; ; i++ {
			url := base_url + string(abc) + "/" + strconv.Itoa(i)
			urls := GetUrls(url)
			if len(urls) < 1 {
				break
			}
			for _, url1 := range urls {
				wg.Add(1)
				go func(url1 string) {
					fmt.Println("url1 :", url1)
					for _, detail_url := range GetUrls(url1) {
						wg.Add(1)
						go func(detail_url string) {
							detail := GetDetail(detail_url, wg, m)
							DetailInsert(detail, db)
							wg.Done()
						}(detail_url)
					}
					wg.Done()
				}(url1)
			}
		}
	}
	wg.Wait()
}
