package api

import (
	"log"
	"time"
	"fmt"
	"net/http"
)

type Config struct {
	Token		string	`json:"token"`
	Domain		string	`json:"domain"`
	ApiUrl		string	`json:"apiUrl"`
	IpApi		string	`json:"ipApi"`
	CheckTime	int64	`json:"checkTime"`
}

type Api struct {
	Cfg 		Config
	client 		*http.Client
}

func (a *Api) Do () {
	a.client = &http.Client{}

	oldIp := ""

	log.SetFlags(log.Lshortfile | log.Ltime | log.Ldate)

	log.Println("Program was tarted")

	for {
		duration := time.Duration(a.Cfg.CheckTime)*time.Second

		if oldIp != "" {
			time.Sleep(duration)
		}

		ip, err := a.getIP()
		if err != nil {
			log.Println(fmt.Sprintf("GetIP finished with error: %v", err))
			continue
		}

		if (oldIp == ip) {
			continue
		}

		oldIp = ip

		records, _ := a.getRecords()

		for _, record := range records {
			if (record.Type == "A") {
				go a.setRecord(record, ip)
			}
		}
	}

	return
}

