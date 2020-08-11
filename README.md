# yt-comments-retriever
Using Go and the YouTube API, the purpose of this little project is to retrieve old comments from videos without having to endlessly scroll the webpage. 

In its current implementation, it retrieves the first 100 comments on the video. Other features may be implemented in the future.

## Running the software

The `properties.go` file needs to be edited to include a [Google API Key](https://console.developers.google.com/apis/credentials)  :

```
package properties

func ReturnKey() string{
	return "YOUR_API_KEY_HERE"
}
```

You will then be able to run the software with the following command:

```
go run main.go YOUR_YOUTUBE_VIDEO_ID_HERE
```



For example, to retrieve the first 100 comments in this video:

https://www.youtube.com/watch?v=gCWjioIR5MM

We use:

```
go run main.go gCWjioIR5MM
```

You will then be able to read the first 100 comments of the video by accessing the following page from your browser:
http://localhost:8080/comments
