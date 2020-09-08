package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func getLink(w http.ResponseWriter, req *http.Request) {
	fmt.Println(req.URL)
}

func getShort(w http.ResponseWriter, req *http.Request) {
	keys, ok := req.URL.Query()["link"]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'link' is missing")
		return
	}

	link := keys[0]
	fmt.Println(link)

	w.Write([]byte("https://cyrilit.tk/s/Qwe21"))
}

func main() {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/getlink/{*}", getLink).Methods("GET")
	rtr.HandleFunc("/getshort", getShort).Methods("GET")
	http.Handle("/", rtr)

	http.ListenAndServe(":5000", nil)
}
