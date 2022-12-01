package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
)

const perPage = 100

type User struct {
	AvatarUrl string `json:"avatar_url"`
	HtmlUrl   string `json:"html_url"`
	Login     string `json:"login"`
}

var selfUser User
var followers, following []User
var wg sync.WaitGroup
var token string

func queryHelper(page int, queryName string, isSelf bool) {
	client := http.Client{}
	var url string
	if isSelf {
		url = "https://api.github.com/user"
	} else {

		url = fmt.Sprintf("https://api.github.com/user/%s?per_page=%d&page=%d", queryName, perPage, page)
	}
	req, err := http.NewRequest("GET", url, nil)
	bearer := "Bearer " + token
	if err != nil {
		log.Fatal(err)
	}

	req.Header = http.Header{
		"Content-Type":  {"application/json"},
		"Authorization": {bearer},
		"accept":        {"application/vnd.github+json"},
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if isSelf {
		json.Unmarshal(data, &selfUser)
	} else {
		var tmp []User
		json.Unmarshal(data, &tmp)
		if queryName == "following" {
			following = append(following, tmp...)
		} else if queryName == "followers" {
			followers = append(followers, tmp...)
		}
	}

	// fmt.Println(string(data))
	if len(data) >= perPage {
		queryHelper(page+1, queryName, false)
	}
}

func queryFollowing(page int) {
	defer wg.Done()
	queryHelper(1, "following", false)
}

func queryFollower(page int) {
	defer wg.Done()
	queryHelper(1, "followers", false)
}

func querySelfUser() {
	defer wg.Done()
	queryHelper(1, "", true)
}

func generateMD() {
	f, err := os.Create("./README.md")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("## %s\n<img src='%s' width='120' />\n", selfUser.Login, selfUser.AvatarUrl))
	f.WriteString(fmt.Sprintf("## Followers <kbd>%d</kbd>\n%s\n", len(followers), formatTable(followers)))
	f.WriteString(fmt.Sprintf("## Following <kbd>%d</kbd>\n%s\n", len(following), formatTable(following)))
}

func formatTable(arr []User) string {
	res := "<table>\n"
	n := len(arr)
	rows := n / 4
	if n%4 != 0 {
		rows += 1
	}
	for i := 0; i < rows; i++ {
		tmp := ""
		for j := 0; j < 4 && i*4+j < n; j++ {
			tmp += fmt.Sprintf("<td width='150' align='center'>\n%s\n</td>\n", formatUser(arr[i*4+j]))
		}
		res += fmt.Sprintf("<tr>%s</tr>", tmp)
	}
	res += "</table>"
	return res
}

func formatUser(u User) string {
	return fmt.Sprintf(`<a href="%s">
<img src="%s" width="50">
<br />
%s
</a>`, u.HtmlUrl, u.AvatarUrl, u.Login)
}

func main() {
	token = os.Getenv("TOKEN")
	wg.Add(3)
	go queryFollower(1)
	go queryFollowing(1)
	go querySelfUser()
	wg.Wait()
	// fmt.Println(selfUser, followers, following)
	fmt.Println(formatUser(selfUser))
	generateMD()
}
