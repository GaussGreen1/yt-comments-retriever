package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/buger/jsonparser"
)

type Comment struct {
	Author string
	Text   string
	Date   time.Time
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World")
}

func GetLastPage() []byte {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/commentThreads", nil)
	//req, err := http.NewRequest("GET", "https://developers.google.com/apis-explorer/#p/youtube/v3/youtube.commentThreads", nil)
	//req, err := http.NewRequest("GET", "https://developers.google.com/apis-explorer/#p/youtube/v3/youtube.commentThreads.list?part=snippet,replies&videoId=wtLJPvx7-ys", nil)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	q := req.URL.Query()
	q.Add("key", "AIzaSyAIDXNmChkxP2IpabW_f-8qAvjfsi9aRoo")
	q.Add("videoId", "4wLBhj_yzaU")
	q.Add("part", "snippet,replies")
	q.Add("maxResults", "100")
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	fmt.Println("Response Status: ", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)

	nextPageToken, _, _, _ := jsonparser.Get(body, "nextPageToken")
	fmt.Println(string(nextPageToken))

	i := 0

	for nextPageToken != nil {
		req, err = http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/commentThreads", nil)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		q = req.URL.Query()
		q.Add("key", "AIzaSyAIDXNmChkxP2IpabW_f-8qAvjfsi9aRoo")
		q.Add("videoId", "4wLBhj_yzaU")
		q.Add("part", "snippet,replies")
		q.Add("maxResults", "100")
		q.Add("pageToken", string(nextPageToken))
		req.URL.RawQuery = q.Encode()

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()

		fmt.Println("Response Status: ", resp.Status)
		body, _ = ioutil.ReadAll(resp.Body)
		nextPageToken, _, _, _ = jsonparser.Get(body, "nextPageToken")
		i++
		fmt.Println(i)

	}

	return body
}

func comments(w http.ResponseWriter, r *http.Request) {
	layout := "2006-01-02T15:04:05Z07:00"

	var comments []Comment
	commentJson := GetLastPage()

	jsonparser.ArrayEach(commentJson, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var thisComment Comment

		authorDisplayName, _, _, _ := jsonparser.Get(value, "snippet", "topLevelComment", "snippet", "authorDisplayName")
		textOriginal, _, _, _ := jsonparser.Get(value, "snippet", "topLevelComment", "snippet", "textOriginal")
		publishedAt, _, _, _ := jsonparser.Get(value, "snippet", "topLevelComment", "snippet", "publishedAt")

		thisComment.Author = string(authorDisplayName)
		thisComment.Text = string(textOriginal)
		thisComment.Date, _ = time.Parse(layout, string(publishedAt))

		comments = append(comments, thisComment)

	}, "items") //, "snippet", "topLevelComment", "snippet")

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].Date.Before(comments[j].Date)
	})

	for j := 0; j < len(comments); j++ {
		stringToPrint := "Author:\n"
		stringToPrint += comments[j].Author
		stringToPrint += "\n\nComment:\n"
		stringToPrint += comments[j].Text
		stringToPrint += "\n\nDate:\n"
		stringToPrint += comments[j].Date.String()
		stringToPrint += "\n\n\n"

		fmt.Fprintf(w, "%s", stringToPrint)
	}

}

func setupRoutes() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/comments", comments)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	fmt.Println("Started application")
	setupRoutes()
}
