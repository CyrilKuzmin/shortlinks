package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var domainName = "cyrilit.tk"
var log = logrus.New()

func makeLinkChecks(URL string) error {
	// if it's not a URL at all
	_, err := url.ParseRequestURI(URL)
	if err != nil {
		return err
	}
	// if it's already our URL
	if strings.HasPrefix(URL, "https://"+domainName) {
		return fmt.Errorf("It's already short")
	}
	return nil
}

func getLink(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
}

func getShort(w http.ResponseWriter, req *http.Request) {
	keys, ok := req.URL.Query()["link"]

	if !ok || len(keys[0]) < 1 {
		log.Error("Param 'link' is missing")
		w.WriteHeader(400)
		fmt.Fprintln(w, "Param 'link' is missing")
		return
	}

	link := keys[0]

	err := makeLinkChecks(link)

	if err != nil {
		log.Error(err)
		w.WriteHeader(400)
		fmt.Fprintln(w, err.Error())
		return
	}

	w.Write([]byte("https://cyrilit.tk/s/Qwe21"))
}

func main() {
	log.Formatter = &logrus.TextFormatter{
		FullTimestamp: true,
	}
	log.SetOutput(os.Stdout)

	rtr := mux.NewRouter()
	rtr.HandleFunc("/getlink/{*}", getLink).Methods("GET")
	rtr.HandleFunc("/getshort", getShort).Methods("GET")
	http.Handle("/", rtr)

	http.ListenAndServe(":5000", nil)
}
