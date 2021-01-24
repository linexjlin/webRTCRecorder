package main

import (
	"encoding/json"
	"fmt"
	"github.com/linexjlin/simple-log"
	"net/http"
)

type StdRsp struct {
	Success bool
	Message string
	Data    interface{}
}

func record(w http.ResponseWriter, r *http.Request) {
	var ret StdRsp

	type Req struct {
		Channel string
		SDP     string
	}

	var req Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(err)
		ret.Success = false
		ret.Message = fmt.Sprintln(err)
	} else {
		if des, err := Rec.Init(req.SDP, req.Channel); err != nil {
			log.Error(err)
			ret.Success = false
			ret.Message = fmt.Sprintln(err)
		} else {
			ret.Success = true
			ret.Data = des
		}
	}
	WriteRsp(w, ret)
}

func WriteRsp(w http.ResponseWriter, i interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if dat, err := json.Marshal(i); err != nil {
		log.Debug(err)
	} else {
		log.Debug(string(dat))
		w.Write(dat)
	}
}
