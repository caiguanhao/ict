package main

import (
	"flag"
	"log"
	"net/http"
)

func main() {
	address := flag.String("address", "127.0.0.1:12345", "address to listen to")
	billAcceptor := flag.String("ba", "", "bill acceptor device")
	uca := flag.String("uca", "", "uca device")
	mh := flag.String("mh", "", "mini hopper device")
	flag.Parse()

	if *billAcceptor != "" {
		startBillAcceptor(*billAcceptor)
	}

	if *uca != "" {
		startUCA(*uca)
	}

	if *mh != "" {
		startMiniHopper(*mh)
	}

	log.Fatal(http.ListenAndServe(*address, logRequest(http.DefaultServeMux)))
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
