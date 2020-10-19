package main

import (
	"encoding/json"
	"strconv"
	"time"
	"unicode"

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

func readBody(r *http.Request, data interface{}) {
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

func ctx(key ctxKey, r *http.Request) interface{} {
	if v := r.Context().Value(key); v != nil {
		return v
	}
	return nil
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

func getCats(c string) []interface{} {

	// Divide string to []string
	cats := strings.FieldsFunc(c, func(c rune) bool {
		return !unicode.IsNumber(c)
	})

	// Get all categories from DB
	var categ []struct {
		ID   int64
		Name string
	}
	catQuery := `SELECT categoryId, name FROM categories`
	sliceFromDB(&categ, catQuery, nil)

	var res []interface{}
	for _, c := range cats {
		for _, k := range categ {
			id, _ := strconv.ParseInt(c, 10, 64)
			if k.ID == id {
				res = append(res, k)
				break
			}
		}
	}
	return res
}
