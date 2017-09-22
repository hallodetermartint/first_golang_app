package main

// Legg til kontroll på home url for at den ikke skal innehole "/" på slutten

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"io"
	"strings"
	"os"
)

// Payload struct to hold all data to be marshaled as json
type Payload struct {
	Project string 			`json:"full_name"`
	Owner struct {
		Login string 		`json:"login"`
	}
	TopCommitter Committer
	Language []string
}

// Committer struct to hold top contributor in the project
type Committer struct {
	Login string			`json:"login"`
	Commits int				`json:"contributions"`
}

func main()	{
	//port is required for heroku deployment
	port := os.Getenv("PORT")
	// Set up listenAndServe with handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/", defaultHandler)
	http.ListenAndServe(":"+port, mux)
}

// Handler for different cases
func defaultHandler(w http.ResponseWriter, r *http.Request) {
	myURL := r.URL.String()

	switch {
	// URL only contains one backslash
	case len(myURL) == 1 && strings.Contains(myURL, "/"):
		io.WriteString(w, "No github address specified,\n" +
							 "please input the URL in the following format:\n" +
							 "github.com/{organisation or username}/{repository}")

	// URL contains github address
	case strings.Contains(myURL, "github.com"):

		myURLFormatted := urlFormat(myURL)
		if myURLFormatted == "invalid" {
			fmt.Fprintln(w, "Unexpected input")
		} else {
			io.WriteString(w, myURLFormatted + "\n")

			resp, err := http.Get(myURLFormatted)
			if err != nil {
				fmt.Fprintln(w, err)
				fmt.Fprintf(w, "something went wrong while trying to \"get\" body from %v", myURL)
			}

			defer resp.Body.Close()
			// Handles statuscodes 404, 403 and 200
			if resp.StatusCode == 404 {
			fmt.Fprintln(w, "404: Repository not found")
			} else if resp.StatusCode == 403 {
			fmt.Fprintln(w, "403: Forbidden. No access")
			} else if resp.StatusCode == 200 {
			fmt.Fprintln(w, "200: OK. found repository")
			presentData(myURLFormatted, w)
			} else {
			fmt.Fprintln(w, "Unexpected input")
			}
		}

	// Default message for inputs without dedicated cases
	default:
		io.WriteString(w, "Unexpected input")
	}

}

func urlFormat(url string) string{
	urlSplit := strings.Split(url, "github.com/")
	if !(strings.Contains(urlSplit[1], "/")) {
		return "invalid"
	}
	urlSecondSplit := strings.Split(urlSplit[1], "/")
	urlNew := "https://api.github.com/repos/" + urlSecondSplit[0] + "/" + urlSecondSplit[1]
	return urlNew
}

func presentData(myURL string, w http.ResponseWriter) []byte{
	res := make(map[string]interface{})
	committerArray := [1]Committer{}
	payload := new(Payload)

	urlHome := 			myURL
	urlLanguage := 		myURL + "/languages"
	urlContributors := 	myURL + "/contributors"

	// URL for home needs to end without slash
	strings.TrimSuffix(urlHome, "/")
	fmt.Println(urlHome)

	bodyLanguage := 	getBody(urlLanguage)
	bodyHome := 		getBody(urlHome)
	bodyContributors := getBody(urlContributors)

	// Unmarshal owner & name
	err := json.Unmarshal(bodyHome, &payload)
	if err != nil {
		fmt.Println(err)
		fmt.Println("something went wrong while trying to unmarshal bodyHome. Remove later")
	}

	// Unmarshal language
	err = json.Unmarshal(bodyLanguage, &res)
	if err != nil {
		fmt.Println(err)
		fmt.Println("something went wrong while trying to unmarshal bodyLanguage. Remove later")
	}

	for key := range res {
		payload.Language = append(payload.Language, key)
	}

	// Unmarshal top contributor
	err = json.Unmarshal(bodyContributors, &committerArray)
	if err != nil {
		fmt.Println(err)
		fmt.Println("something went wrong while trying to unmarshal bodyContributors")
	}
	payload.TopCommitter.Commits = committerArray[0].Commits
	payload.TopCommitter.Login = committerArray[0].Login

	testJSON, err := json.MarshalIndent(payload, " ", "    ")
	if err != nil {
		fmt.Println(err)
		fmt.Println("something went wrong while trying to marshal struct")
	}
	fmt.Fprintf(w, "%s\n", testJSON)
	fmt.Printf("%s\n", testJSON)
	return testJSON
}
func getBody(url string) []byte {
	// GET request
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("something went wrong while trying to \"get\" body from %v", url)
	}
	defer resp.Body.Close()

	// Read body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("something went wrong while trying to read response body from %v. Remove later", url)
	}
	return body
}