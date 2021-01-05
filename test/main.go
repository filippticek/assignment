package main

import (
	"bufio"
	"bytes"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func client(reqID int, method string, urlValue string, data string) {
	defer myWaitGroup1.Done()
	myClient := &http.Client{
		Transport:     nil,
		CheckRedirect: nil,
		Jar:           nil,
		Timeout:       0,
	}
	log.Println("ReqID: " + strconv.Itoa(reqID) + " " + method + " " + urlValue + " " + data)
	var url = "http://localhost:8080/" + urlValue
	req, err := http.NewRequest(method, url, strings.NewReader(data))
	if data != " " {
		req.Header.Add("Content-Type", "application/json")
	}
	if method == "GET" {
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		log.Fatal(err)
	}
	resp, err := myClient.Do(req)
	if err != nil {
		log.Println(strconv.Itoa(reqID) + err.Error())
		return
	}
	var response = "RespID: " + strconv.Itoa(reqID) + " ReqMethod" + method + " " + resp.Status
	if resp.StatusCode == http.StatusOK {
		response += " Content"
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		response += buf.String()
	}

	log.Println(response)
}

var myWaitGroup1 sync.WaitGroup

func main() {

	file, err := os.Open("operations")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	var i int
	for scanner.Scan() {
		var split = make([]string, 4)
		split = strings.Split(scanner.Text(), " ")
		var method = split[0]
		var urlValue = split[1][1:]
		var data = split[2]
		myWaitGroup1.Add(1)
		go client(i, method, urlValue, data)
		i++
	}
	myWaitGroup1.Wait()
}
