package main

import (
	"flag"
	"github.com/linexjlin/webRTCRecorder/serverSide/recorder"
	"net/http"
)

var Rec *recorder.Recorder

func main() {
	addr := flag.String("listen", ":8025", "-listen :8025")
	flag.Parse()
	Rec = recorder.NewRecorder()
	http.HandleFunc("/API/Record", record)
	http.ListenAndServe(*addr, nil)
}
