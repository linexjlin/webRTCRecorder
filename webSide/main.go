package main

import (
	"flag"
	"net/http"
)

func main() {
	addr := flag.String("listen", ":8032", "-listen :8032")
	flag.Parse()
	fs := http.FileServer(http.Dir("./web"))
	http.Handle("/", fs)
	http.ListenAndServe(*addr, nil)
}
