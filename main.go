package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"./properties"

	"github.com/buger/jsonparser"
)

type Comment struct {
	Author    string
	Text      string
	Date      time.Time
	CommentID string
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

	pageTokens := []string{"", ""}

	body := callApi(stringKey, videoId, "")

	nextPageToken, _, _, _ := jsonparser.Get(body, "nextPageToken")

	i := 0
	for nextPageToken != nil {
		nextPageTokenString := string(nextPageToken)
		body = callApi(stringKey, videoId, nextPageTokenString)

		lastPageCalled = nextPageTokenString
		pageTokens[0] = pageTokens[1]
		pageTokens[1] = lastPageCalled

		nextPageToken, _, _, _ = jsonparser.Get(body, "nextPageToken")

		i++
		fmt.Println(i)

	}

	return body, pageTokens[0]
}

func appendElementsToCommentsArray(body []byte, comments []Comment) []Comment {
	layout := "2006-01-02T15:04:05Z07:00"

	jsonparser.ArrayEach(body, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		var thisComment Comment
		var textReplaced string

		authorDisplayName, _, _, _ := jsonparser.Get(value, "snippet", "topLevelComment", "snippet", "authorDisplayName")
		textOriginal, _, _, _ := jsonparser.Get(value, "snippet", "topLevelComment", "snippet", "textOriginal")
		publishedAt, _, _, _ := jsonparser.Get(value, "snippet", "topLevelComment", "snippet", "publishedAt")
		commentID, _, _, _ := jsonparser.Get(value, "snippet", "topLevelComment", "id")

		textReplaced = strings.Replace(string(textOriginal), `\n`, "\n", -1)
		textReplaced = strings.Replace(string(textReplaced), `\r`, "\r", -1)
		textReplaced = strings.Replace(string(textReplaced), `\"`, "\"", -1)
		//TODO altre robe simili formattate male da rimpiazzare?

		thisComment.Author = string(authorDisplayName)
		thisComment.Text = textReplaced
		thisComment.Date, _ = time.Parse(layout, string(publishedAt))
		thisComment.CommentID = string(commentID)

		comments = append(comments, thisComment)

	}, "items")

	return comments

}

func comments(w http.ResponseWriter, r *http.Request) {

	var comments []Comment
	body, secondToLastPageCalled := GetLastPage()

	comments = appendElementsToCommentsArray(body, comments)

	if len(comments) < 100 && secondToLastPageCalled != "" {
		lastBody := callApi(properties.ReturnKey(), os.Args[1], secondToLastPageCalled)
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
