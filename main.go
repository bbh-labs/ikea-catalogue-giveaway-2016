package main

import (
	"flag"
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

	NameMinLength     = 6
	NameMaxLength     = 32
	Address1MinLength = 6
	Address1MaxLength = 64
	PostalCodeLength  = 5
	CityMinLength     = 3
	CityMaxLength     = 32
	StateMinLength    = 5
	StateMaxLength    = 15
	EmailMinLength    = 3
	EmailMaxLength    = 254
)

var States = []string{
	"Johor",
	"Kedah",
	"Kelantan",
	"Malacca",
	"Negeri Sembilan",
	"Pahang",
	"Perak",
	"Perlis",
	"Penang",
	"Sabah",
	"Sarawak",
	"Selangor",
	"Terengganu",
}

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
	if len(name) < NameMinLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Name length should be at least " + itoa(NameMinLength) + " characters long"))
		return
	} else if len(name) > NameMaxLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Name length should not exceed " + itoa(NameMaxLength) + " characters long"))
		return
	}

	// Address 1
	address1 := r.FormValue("address1")
	if len(address1) < NameMinLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Address 1 length should be at least " + itoa(Address1MinLength) + " characters long"))
		return
	} else if len(address1) > NameMaxLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Address 1 length should not exceed " + itoa(Address1MaxLength) + " characters long"))
		return
	}

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
	if len(city) < CityMinLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("City length should be at least " + itoa(CityMinLength) + " characters long"))
		return
	} else if len(city) > CityMaxLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("City length should not exceed " + itoa(CityMaxLength) + " characters long"))
		return
	}

	// State
	state := r.FormValue("state")
	if len(state) < StateMinLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("State length should be at least " + itoa(StateMinLength) + " characters long"))
		return
	} else if len(state) > StateMaxLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("State length should not exceed " + itoa(StateMaxLength) + " characters long"))
		return
	} else if !isValidState(state) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("State length is invalid"))
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
	if len(email) < EmailMinLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Email length should be at least " + itoa(EmailMinLength) + " characters long"))
		return
	} else if len(email) > EmailMaxLength {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Email length should not exceed " + itoa(EmailMaxLength) + " characters long"))
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

func isValidState(state string) bool {
	for _, v := range States {
		if state == v {
			return true
		}
	}
	return false
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
	router.HandleFunc("/success", success)
	router.HandleFunc("/duplicate", duplicate)
	router.HandleFunc("/failed", failed)
	router.HandleFunc("/outofstock", outofstock)

	// start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	n := negroni.Classic()
	n.UseHandler(router)
	n.Run(":" + port)
}
