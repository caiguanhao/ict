package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/caiguanhao/ict/html"
)

func main() {
	address := flag.String("address", "127.0.0.1:12345", "address to listen to")
	billAcceptor := flag.String("ba", "", "bill acceptor device, optional")
	billAcceptorProtocol := flag.String("bap", "ICT106U", "bill acceptor protocol (ICT002U, ICT104U, ICT106U)")
	uca := flag.String("uca", "", "uca device, optional")
	mh := flag.String("mh", "", "mini hopper device, optional")
	serve := flag.Bool("serve", false, "whether to serve index.html")
	flag.Parse()

	if *billAcceptor != "" {
		startBillAcceptor(*billAcceptor, *billAcceptorProtocol)
	}

	if *uca != "" {
		startUCA(*uca)
	}

	if *mh != "" {
		startMiniHopper(*mh)
	}

	if *serve {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			w.Write([]byte(html.Index()))
		})
		log.Println("listening", "http://"+*address)
	} else {
		log.Println("listening", *address)
	}
	log.Fatal(http.ListenAndServe(*address, logRequest(http.DefaultServeMux)))
}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
