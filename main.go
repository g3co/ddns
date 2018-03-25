package main

import (
	"net/http"
	"log"
	"encoding/json"
	"net/url"
	"strconv"
	"io/ioutil"
	"fmt"
	"bytes"
	"time"
	"os"
)

type records struct {
	Type      string   `json:"type"`
	Content   string   `json:"content"`
	RecordId  int	   `json:"record_id"`
	Domain    string   `json:"domain"`
}

type apiResponse struct {
	Success   string	`json:"success"`
	Domain    string   	`json:"domain"`
	Records   []records	`json:"records"`
}

type apiRecordSetResponse struct {
	Success   string	`json:"success"`
	Domain    string   	`json:"domain"`
	Record    records	`json:"record"`
}

type config struct {
	Token		string	`json:"token"`
	Domain		string	`json:"domain"`
	ApiUrl		string	`json:"apiUrl"`
	IpApi		string	`json:"ipApi"`
	CheckTime	int64	`json:"checkTime"`
}


func main() {

	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	config := config{}
	err := decoder.Decode(&config)
	if err != nil {
		log.Println(fmt.Sprintf("Get config finished with error: %v", err))
	}

	do(config)
}

func do (cfg config) {
	client := &http.Client{}

	oldIp := ""

	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	log.Println("Program was tarted")

	for {
		duration := time.Duration(cfg.CheckTime)*time.Second

		if oldIp != "" {
			time.Sleep(duration)
		}

		IP, err := getIP(cfg, client)
		if err != nil {
			log.Println(fmt.Sprintf("GetIP finished with error: %v", err))
			continue
		}

		if (oldIp == IP) {
			continue
		}

		oldIp = IP

		records, _ := getRecords(cfg, client)

		for _, record := range records {
			if (record.Type == "A") {
				go setRecord(record, IP, cfg, client)
			}
		}
	}

	return
}

func getIP (cfg config, client *http.Client) (ip string, err error) {

	req, err := http.NewRequest("GET", cfg.IpApi, nil)
	if err != nil {
		return
	}
	respIp, err := client.Do(req)
	defer respIp.Body.Close()
	if err != nil {
		return
	}

	rawIp, err := ioutil.ReadAll(respIp.Body);
	if err != nil {
		return
	}
	ip = string(rawIp)

	log.Println(fmt.Sprintf("Current IP: %v", ip))

	return
}

func getRecords(cfg config, client *http.Client) (records []records, err error) {
	u, _ := url.ParseRequestURI(cfg.ApiUrl)

	u.Path = fmt.Sprintf("%v%v", u.Path, "list")
	u.RawQuery = fmt.Sprintf("domain=%v", cfg.Domain)
	req, err := http.NewRequest("GET", fmt.Sprintf("%v", u), nil)
	if err != nil {
		log.Println(fmt.Sprintf("API request has errors: %v", err))
		return
	}

	req.Header.Add("PddToken", cfg.Token)
	resp, err := client.Do(req)
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

func setRecord(record records, ip string, cfg config, client *http.Client) (err error) {
	log.Println(fmt.Sprintf("Record ID: %v start updating", record.RecordId))

	form := url.Values{}
	form.Set("domain", cfg.Domain)
	form.Add("record_id", strconv.Itoa(record.RecordId))
	form.Add("content", ip)

	u, _ := url.ParseRequestURI(cfg.ApiUrl)
	u.Path = fmt.Sprintf("%v%v", u.Path, "edit")

	req, _ := http.NewRequest("POST", fmt.Sprintf("%v", u), bytes.NewBufferString(form.Encode()))
	req.Header.Add("PddToken", cfg.Token)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

	response, err := client.Do(req)
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