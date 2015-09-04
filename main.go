package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	"github.com/bbhmakerlab/debug"
	"github.com/bbhmakerlab/ikea-catalogue-giveaway-2016/store"
	"github.com/bradfitz/http2"
	"github.com/lib/pq"
)

const (
	MaxEntries = 8000

	PostalCodeLength  = 5
)

func home(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/html")
	http.ServeFile(w, r, "public/main.html")
}

func success(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/html")
	http.ServeFile(w, r, "public/success.html")
}

func duplicate(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/html")
	http.ServeFile(w, r, "public/duplicate.html")
}

func failed(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/html")
	http.ServeFile(w, r, "public/failed.html")
}

func outofstock(w http.ResponseWriter, r *http.Request) {
	r.Header.Set("Content-Type", "text/html")
	http.ServeFile(w, r, "public/outofstock.html")
}

func submit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Name
	name := r.FormValue("name")
	if len(name) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Name shouldn't be empty"))
		return
	}

	// Address 1
	address1 := r.FormValue("address1")
	if len(address1) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Address 1 shouldn't be empty"))
		return
	}

	// Address 2
	address2 := r.FormValue("address2")

	// Postal Code
	postalCode := r.FormValue("postal_code")
	if len(postalCode) != 5 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Postal code length must be " + itoa(PostalCodeLength) + " digits long"))
		return
	}

	// City
	city := r.FormValue("city")
	if len(city) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("City shouldn't be empty"))
		return
	}

	// State
	state := r.FormValue("state")
	if len(state) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("State shouldn't be empty"))
		return
	}

	// Country
	country := r.FormValue("country")
	if country != "Malaysia" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Country name is invalid"))
		return
	}

	// Email
	email := r.FormValue("email")
	if len(email) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Address 1 shouldn't be empty"))
		return
	}

	// Check number of entries already submitted
	if n, err := store.CountEntries(); err != nil {
		http.Redirect(w, r, "/failed", 302)
		return
	} else if n >= MaxEntries {
		http.Redirect(w, r, "/outofstock", 302)
		return
	}

	// Insert to database
	if err := store.InsertEntry(name, address1, address2, city, state, country, postalCode, email); err != nil {
		e, ok := err.(*pq.Error)
		if ok && e.Code == "23505" { // unique violation
			http.Redirect(w, r, "/duplicate", 302)
			return
		} else {
			debug.Warn(err)
			http.Redirect(w, r, "/failed", 302)
			return
		}
	}

	http.Redirect(w, r, "/success", 302)
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func startServer() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := http.Server{Addr: ":443"}
	http2.ConfigureServer(&srv, &http2.Server{})
	go func() {
	     debug.Fatal(http.ListenAndServe(":" + port, nil))
	}()

	debug.Fatal(srv.ListenAndServeTLS("/home/website/ssl/server.crt", "/home/website/ssl/server.key"))
}


func main() {
	// parse CLI options
	flag.Parse()

	// start database
	store.Init()

	// configure server
	http.HandleFunc("/", home)
	http.HandleFunc("/submit", submit)
	http.HandleFunc("/success", success)
	http.HandleFunc("/duplicate", duplicate)
	http.HandleFunc("/failed", failed)
	http.HandleFunc("/outofstock", outofstock)
	startServer()
}
