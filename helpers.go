package main

import (
	"encoding/json"
	"strconv"
	"time"

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

func (rep *report) logcheck(col string, log bool) {
	if log {
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

func fromCtx(key ctxKey, r *http.Request) int64 {
	if v := r.Context().Value(key); v != nil {
		return v.(int64)
	}
	return 0
}

func addCookie(w http.ResponseWriter, name, value string, exp time.Time) {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  exp,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, &cookie)
}

func processCategories(validity *report, cats []int64) string {
	validity.logcheck("too much cats", len(cats) > 3)
	validity.logcheck("no cats", len(cats) == 0)
	var strcats string
	for _, cat := range cats {
		check := "SELECT categoryId FROM categories WHERE categoryId = ?"
		inDB := isInDB(check, cat)
		if !inDB {
			validity.logcheck("no such category!", true)
			break
		}
		strcats += "\"" + strconv.FormatInt(cat, 10) + "\""
	}
	return strcats
}

func reqQuery(name string, r *http.Request) string {
	switch len(r.FormValue(name)) {
	case 0:
		return "%"
	default:
		return r.FormValue(name)
	}
}
