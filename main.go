package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"./properties"

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

func callApi(stringKey string, videoId string, pageToken string) []byte {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/youtube/v3/commentThreads", nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	q := req.URL.Query()
	q.Add("key", stringKey)
	q.Add("videoId", videoId)
	q.Add("part", "snippet,replies")
	q.Add("maxResults", "100")
	if pageToken != "" {
		q.Add("pageToken", pageToken)
	}
	req.URL.RawQuery = q.Encode()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	fmt.Println("Response Status: ", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	return body
}

func GetLastPage() ([]byte, string) {
	stringKey := properties.ReturnKey()
	videoId := os.Args[1]
	lastPageCalled := ""

	body := callApi(stringKey, videoId, "")

	nextPageToken, _, _, _ := jsonparser.Get(body, "nextPageToken")

	i := 0
	for nextPageToken != nil {
		nextPageTokenString := string(nextPageToken)
		body = callApi(stringKey, videoId, nextPageTokenString)

		lastPageCalled = nextPageTokenString
		nextPageToken, _, _, _ = jsonparser.Get(body, "nextPageToken")
		i++
		fmt.Println(i)

	}

	return body, lastPageCalled
}

func appendElementsToCommentsArray(body []byte, comments []Comment) []Comment {
	layout := "2006-01-02T15:04:05Z07:00"

	jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var thisComment Comment

		authorDisplayName, _, _, _ := jsonparser.Get(value, "snippet", "topLevelComment", "snippet", "authorDisplayName")
		textOriginal, _, _, _ := jsonparser.Get(value, "snippet", "topLevelComment", "snippet", "textOriginal")
		publishedAt, _, _, _ := jsonparser.Get(value, "snippet", "topLevelComment", "snippet", "publishedAt")

		thisComment.Author = string(authorDisplayName)
		thisComment.Text = string(textOriginal)
		thisComment.Date, _ = time.Parse(layout, string(publishedAt))

		comments = append(comments, thisComment)

	}, "items") //, "snippet", "topLevelComment", "snippet")

	return comments

}

func comments(w http.ResponseWriter, r *http.Request) {

	var comments []Comment
	body, lastPageCalled := GetLastPage()

	comments = appendElementsToCommentsArray(body, comments)

	if len(comments) < 100 && lastPageCalled != "" {
		lastBody := callApi(properties.ReturnKey(), os.Args[1], lastPageCalled)
		comments = appendElementsToCommentsArray(lastBody, comments)
	}

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].Date.Before(comments[j].Date)
	})

	maxValue := 0
	if len(comments) > 100 {
		maxValue = 100
	} else {
		maxValue = len(comments)
	}

	for j := 0; j < maxValue; j++ {
		stringToPrint := "Author:\n" + comments[j].Author + "\n\nComment:\n" + comments[j].Text +
			"\n\nDate:\n" + comments[j].Date.String() + "\n\n\n"

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
