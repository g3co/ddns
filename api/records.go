package api

import (
	"net/url"
	"fmt"
	"net/http"
	"log"
	"io/ioutil"
	"encoding/json"
	"strconv"
	"bytes"
)

type record struct {
	Type      string   `json:"type"`
	Content   string   `json:"content"`
	RecordId  int	   `json:"record_id"`
	Domain    string   `json:"domain"`
}

type apiResponse struct {
	Success   string		`json:"success"`
	Domain    string   		`json:"domain"`
	Records   []record    	`json:"records"`
}

type apiRecordSetResponse struct {
	Success 	string	`json:"success"`
	Domain  	string	`json:"domain"`
	Record  	record	`json:"record"`
}

func (a *Api) getRecords() (records []record, err error) {
	u, _ := url.ParseRequestURI(a.Cfg.ApiUrl)

	u.Path = fmt.Sprintf("%v%v", u.Path, "list")
	u.RawQuery = fmt.Sprintf("domain=%v", a.Cfg.Domain)
	req, err := http.NewRequest("GET", fmt.Sprintf("%v", u), nil)
	if err != nil {
		log.Println(fmt.Sprintf("API request has errors: %v", err))
		return
	}

	req.Header.Add("PddToken", a.Cfg.Token)
	resp, err := a.client.Do(req)
	defer resp.Body.Close()

	if err != nil {
		log.Println(err)
		return
	}

	var t apiResponse
	rawResp, err := ioutil.ReadAll(resp.Body);
	if err != nil {
		log.Println(fmt.Sprintf("Can't read API response: %v", err))
		return
	}

	err = json.Unmarshal(rawResp, &t)
	if err != nil {
		log.Println(fmt.Sprintf("Can't parse API response: %v", err))
		return
	}

	records = t.Records
	return
}

func (a *Api) setRecord(record record, ip string) (err error) {
	log.Println(fmt.Sprintf("Record ID: %v start updating", record.RecordId))

	form := url.Values{}
	form.Set("domain", a.Cfg.Domain)
	form.Add("record_id", strconv.Itoa(record.RecordId))
	form.Add("content", ip)

	u, _ := url.ParseRequestURI(a.Cfg.ApiUrl)
	u.Path = fmt.Sprintf("%v%v", u.Path, "edit")

	req, _ := http.NewRequest("POST", fmt.Sprintf("%v", u), bytes.NewBufferString(form.Encode()))
	req.Header.Add("PddToken", a.Cfg.Token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	response, err := a.client.Do(req)
	defer response.Body.Close()

	if err != nil {
		log.Println(fmt.Sprintf("can't send API request: %v", err))
		return
	}

	responseData, _ := ioutil.ReadAll(response.Body);
	var recordResponse apiRecordSetResponse
	err = json.Unmarshal(responseData, &recordResponse)
	if err == nil {
		if recordResponse.Success == "ok" {
			log.Println(fmt.Sprintf("Record ID: %v had been updated", recordResponse.Record.RecordId))
			return
		}
	}

	return
}
