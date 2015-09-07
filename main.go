package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/bbhmakerlab/debug"
	"github.com/bbhmakerlab/ikea-catalogue-giveaway-2016/store"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
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

func submit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "http://www.malaysia-ikea.com")

	// Name
	name := r.FormValue("name")
	if len(name) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Name is empty"))
		return
	}

	// Address 1
	address1 := r.FormValue("address1")
	if len(address1) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Address 1 is empty"))
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
		w.Write([]byte("City is empty"))
		return
	}

	// State
	state := r.FormValue("state")
	if len(state) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("State is empty"))
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
		w.Write([]byte("Email is empty"))
		return
	}

	// Check number of entries already submitted
	if n, err := store.CountEntries(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	} else if n >= MaxEntries {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("out of stock"))
		return
	}

	// Insert to database
	if err := store.InsertEntry(name, address1, address2, city, state, country, postalCode, email); err != nil {
		e, ok := err.(*pq.Error)
		if ok && e.Code == "23505" { // unique violation
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("duplicate"))
			return
		} else {
			debug.Warn(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
}

// Check number of entries already submitted
func count(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "http://www.malaysia-ikea.com")

	if n, err := store.CountEntries(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.Write([]byte(fmt.Sprintf("%d", n)))
	}
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func main() {
	// parse CLI options
	flag.Parse()

	// start database
	store.Init()

	// configure server
	router := mux.NewRouter()
	router.HandleFunc("/", home)
	router.HandleFunc("/submit", submit)
	router.HandleFunc("/count", count)

	// start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":" + port)
}
