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
	"regexp"
	"time"
)

type records struct {
	Type string
	Content string
	RecordId int `json:"record_id"`
}

type test_struct struct {
	Success string
	Domain string
	Records []records
}


func main() {
	token := "***"
	domain := "site.net"

	apiUrl := "https://pddimp.yandex.ru/api2/admin/dns/"

	client := &http.Client{}

	oldIp := ""

	for {
		duration := time.Duration(300)*time.Second
		time.Sleep(duration)

		req, _ := http.NewRequest("GET", "http://ipv4.internet.yandex.net/internet/api/v0/ip", nil)
		respIp, _ := client.Do(req)
		ipResponseData, _ := ioutil.ReadAll(respIp.Body);
		ipString := string(ipResponseData)
		re := regexp.MustCompile(`[^"]+`)
		ip_address := re.FindStringSubmatch(ipString)

		fmt.Println(ip_address)

		respIp.Body.Close()

		if (oldIp == ip_address[0]) {
			continue
		}

		oldIp = ip_address[0]

		u, _ := url.ParseRequestURI(apiUrl)
		u.Path = fmt.Sprintf("%v%v", u.Path, "list")

		u.RawQuery = fmt.Sprintf("domain=%v", domain)

		req, err := http.NewRequest("GET", fmt.Sprintf("%v", u), nil)

		if err != nil {
			panic(err)
		}
		req.Header.Add("PddToken", token)

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		decoder := json.NewDecoder(resp.Body)
		var t test_struct
		err = decoder.Decode(&t)

		if err != nil {
			panic(err)
		}

		resp.Body.Close()

		for _, element := range t.Records {
			if (element.Type == "A") {
				log.Println(element.RecordId)

				form := url.Values{}
				form.Set("domain", domain)
				form.Add("record_id", strconv.Itoa(element.RecordId))
				form.Add("content", ip_address[0])

				u, _ := url.ParseRequestURI(apiUrl)
				u.Path = fmt.Sprintf("%v%v", u.Path, "edit")

				req, _ := http.NewRequest("POST", fmt.Sprintf("%v", u), bytes.NewBufferString(form.Encode()))
				req.Header.Add("PddToken", token)
				req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Add("Content-Length", strconv.Itoa(len(form.Encode())))

				response, err := client.Do(req)
				if err != nil {
					panic(err)
				}

				responseData, _ := ioutil.ReadAll(response.Body);
				responseString := string(responseData)
				fmt.Println(responseString)
			}
		}
	}
}