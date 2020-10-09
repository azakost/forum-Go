package main

import (
	"encoding/json"

	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

func err(e error) {
	if e != nil {
		panic(e)
	}
}

func structBody(r *http.Request, data interface{}) {
	body, readError := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	err(readError)
	unmarshalError := json.Unmarshal(body, &data)
	err(unmarshalError)
}

type report []string

func (rep *report) regcheck(col, str, regex string) {
	if !regexp.MustCompile(regex).MatchString(str) {
		*rep = append(*rep, col)
	}
}

func (rep *report) errcheck(col string, err error, str string) {
	if strings.Contains(err.Error(), str) {
		*rep = append(*rep, col)
	}
}

func returnJSON(d interface{}, w http.ResponseWriter) {
	js, jsonError := json.Marshal(d)
	err(jsonError)
	w.Header().Set("Content-Type", "application/json")
	_, writeError := w.Write(js)
	err(writeError)
}
