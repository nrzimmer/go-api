package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
	"github.com/urfave/negroni"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleApiGet)

	n := negroni.New()

	log := negroni.NewLogger()
	log.SetFormat("{{.Status}} {{.Hostname}} {{.Method}} {{.Request.RequestURI}}")

	recovery := negroni.NewRecovery()

	n.Use(log)
	n.Use(recovery)
	n.UseHandler(mux)

	fmt.Println("Listening...")
	http.ListenAndServe(":3000", n)
}

func handleApiGet(w http.ResponseWriter, r *http.Request) {
	// Get used parameters from URL
	q, q_ok := r.URL.Query()["q"]
	author, author_ok := r.URL.Query()["author"]
	title, title_ok := r.URL.Query()["title"]

	// Initialize an empty query string
	query := ""

	// If there is a q parameter, use it
	// If not, try to search with author and title
	if q_ok && len(q) == 1 {
		query = "q=" + url.QueryEscape(q[0])
	} else {
		if author_ok && len(author) == 1 {
			query = "author=" + url.QueryEscape(author[0])
		}
		// If there is an title query, we need to append an & to the author query string
		if title_ok && len(title) == 1 {
			if query != "" {
				query = query + "&"
			}
			query = query + "title=" + url.QueryEscape(title[0])
		}
	}

	// If query is empty return an error response
	if query == "" {
		http.Error(w, "Invalid Url Parameters", http.StatusBadRequest)
		return
	}

	// If we found a query string, use it
	status, response := OpenLibraryRequest(query)

	// If the OpenLibrary server responded with an error, return it to the client
	if status < 200 || status >= 300 {
		http.Error(w, "OpenLibrary server returned an Error", status)
		return
	}

	// Validate the received JSON (expected to be an JSON instance)
	_, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "OpenLibrary returned an Invalid JSON", http.StatusInternalServerError)
		return
	}

	// TODO
	// Parse the response here and remove unused fields

	// Write the JSON back to the Client
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(response))
}

func OpenLibraryRequest(q string) (int, string) {
	// Create a Resty Client
	client := resty.New()

	// Query the openlibrary API
	fmt.Printf("[resty]   Requesting: https://openlibrary.org/search.json?%s\n", q)
	resp, err := client.R().EnableTrace().Get("https://openlibrary.org/search.json?" + q)
	fmt.Printf("[resty]   %d openlibrary.org GET /search.json?%s\n", resp.StatusCode(), q)

	// If we got an error, return it, if not, return the response
	if err == nil {
		return resp.StatusCode(), resp.String()
	} else {
		return resp.StatusCode(), err.Error()
	}
}