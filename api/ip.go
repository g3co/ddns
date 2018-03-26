package api

import (
	"net/http"
	"io/ioutil"
	"log"
	"fmt"
)

func (a *Api) getIP () (ip string, err error) {

	req, err := http.NewRequest("GET", a.Cfg.IpApi, nil)
	if err != nil {
		return
	}
	respIp, err := a.client.Do(req)
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