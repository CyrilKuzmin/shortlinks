package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var domain = os.Getenv("DOMAIN")
var log = logrus.New()
var letterRunes = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
var shortlen = 4

var client = redis.NewClient(&redis.Options{
	Addr:     os.Getenv("REDIS_ADDR"),
	Password: os.Getenv("REDIS_PW"),
	DB:       0,
})

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func makeLinkChecks(URL string) error {
	// if it's not a URL at all
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return err
	}
	// if it's already our URL
	if strings.HasPrefix(URL, domain) {
		return fmt.Errorf("It's already short")
	}
	return nil
}

func getOriginal(w http.ResponseWriter, req *http.Request) {
	slink := trimFirstRune(req.URL.Path)
	val, err := client.Get(slink).Result()
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		log.Error(err)
		w.WriteHeader(400)
		fmt.Fprintln(w, "{\"error\":\""+err.Error()+"\"}")
		return
	}
	http.Redirect(w, req, val, 302)
}

func getShort(w http.ResponseWriter, req *http.Request) {
	keys, ok := req.URL.Query()["link"]
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	log.Print(keys)
	if !ok || len(keys[0]) < 1 {
		log.Error("Param 'link' is missing")
		w.WriteHeader(400)
		fmt.Fprintln(w, "{\"error\":\"Param 'link' is missing\"}")
		return
	}

	link := keys[0]

	err := makeLinkChecks(link)

	if err != nil {
		log.Errorf("Link checks failed: %v", err)
		w.WriteHeader(400)
		fmt.Fprintln(w, "{\"error\":\""+err.Error()+"\"}")
		return
	}
	val, err := client.Get(link).Result()

	if err == nil {
		w.Write([]byte("{\"shortlink\":\"" + domain + val + "\"}"))
		return
	}

	slink := RandStringRunes(shortlen)

	err = client.Set(slink, link, 0).Err()
	if err != nil {
		log.Errorf("Cannot set slink-link kv: %v", err)
		w.WriteHeader(400)
		fmt.Fprintln(w, "{\"error\":\""+err.Error()+"\"}")
		return
	}

	err = client.Set(link, slink, 0).Err()
	if err != nil {
		log.Errorf("Cannot set link-slink kv: %v", err)
		w.WriteHeader(400)
		fmt.Fprintln(w, "{\"error\":\""+err.Error()+"\"}")
		return
	}

	w.Write([]byte("{\"shortlink\":\"" + domain + slink + "\"}"))
}

func main() {
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
	log.SetOutput(os.Stdout)
	log.Printf("Statrting app on %v... Redis: %v", domain, os.Getenv("REDIS_ADDR"))
	rtr := mux.NewRouter()
	rtr.HandleFunc("/{*}", getOriginal).Methods("GET")
	rtr.HandleFunc("/s/", getShort).Methods("GET")
	http.Handle("/", rtr)

	http.ListenAndServe(":5000", nil)
}
