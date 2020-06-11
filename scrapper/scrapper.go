package scrapper

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type extractedJob struct {
	id       string
	location string
	title    string
	salary   string
	summary  string
}

//Scrape indded by a term
func Scrape(term string) {
	var baseURL string = "https://kr.indeed.com/jobs?q=" + term + "&limit=50"

	var jobs []extractedJob
	c := make(chan []extractedJob)

	totalPages := getPages(baseURL)

	for i := 0; i < totalPages; i++ {
		//Page Data 추출
		go getPage(i, baseURL, c)
	}

	for i := 0; i < totalPages; i++ {
		extractedJobs := <-c
		jobs = append(jobs, extractedJobs...)
	}

	writeJobs(jobs)

	fmt.Println("Done, extracted", len(jobs))
}

func getPage(page int, url string, mainC chan<- []extractedJob) {
	var jobs []extractedJob

	c := make(chan extractedJob)

	pageURL := url + "&start=" + strconv.Itoa(page*50)
	fmt.Println("Requesting ", pageURL)

	res, err := http.Get(pageURL)
	checkErr(err)
	checkCode(res)

	//함수가 종료될때 메모리반환
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	//Card div
	searchCards := doc.Find(".jobsearch-SerpJobCard")
	searchCards.Each(func(i int, card *goquery.Selection) {
		//Card의 item을 struct로 return
		go extractJob(card, c)
	})

	for i := 0; i < searchCards.Length(); i++ {
		job := <-c
		jobs = append(jobs, job)
	}

	mainC <- jobs
}

//items to extractedJob struct
func extractJob(card *goquery.Selection, c chan<- extractedJob) {
	id, _ := card.Attr("data-jk")
	title := CleanString(card.Find(".title>a").Text())
	location := CleanString(card.Find(".sjcl").Text())
	salary := CleanString(card.Find(".salaryText").Text())
	summary := CleanString(card.Find(".summary").Text())
	c <- extractedJob{
		id:       id,
		title:    title,
		location: location,
		salary:   salary,
		summary:  summary}
}

//CleanString cleans a string empty
func CleanString(str string) string {
	//TrimSpace = string의 앞뒤 빈공간 없애기
	//Fields = string의 중간 빈공간 없애서 배열로 반환
	//join = 배열을 ,뒤의 문자로 합쳐 1개의 string으로 반환
	return strings.Join(strings.Fields(strings.TrimSpace(str)), " ")
}

func getPages(url string) int {
	pages := 0
	res, err := http.Get(url)
	checkErr(err)
	checkCode(res)

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	checkErr(err)

	//page 개수 가져오기
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		pages = s.Find("a").Length()
	})

	return pages
}

func writeJobs(jobs []extractedJob) {
	c := make(chan []string)

	//jobs.csv file로 파일명 저장
	file, err := os.Create("jobs.csv")
	checkErr(err)

	//csv file create
	w := csv.NewWriter(file)
	//함수 종료 시 flush로 할당 해제
	defer w.Flush()

	headers := []string{"Link", "Title", "Location", "Salary", "Summary"}

	wErr := w.Write(headers)
	checkErr(wErr)

	for _, job := range jobs {
		go jobSlices(job, c)
	}

	for i := 0; i < len(jobs); i++ {
		jwErr := w.Write(<-c)
		checkErr(jwErr)
	}
}

func jobSlices(job extractedJob, c chan<- []string) {
	c <- []string{"https://kr.indeed.com/viewjob?jk=" + job.id, job.title, job.location, job.salary, job.summary}
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func checkCode(res *http.Response) {
	if res.StatusCode != 200 {
		log.Fatalln("Request failed with Status : ", res.StatusCode)
	}
}
