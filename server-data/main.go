// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"text/template"
	"time"

	uuid "github.com/satori/go.uuid"
)

type comp struct {
	ProfileVersion,
	ProfileSHA,
	EndPoint,
	Environment,
	NodeName,
	NodeUUID,
	PlatformName,
	PlatformRelease,
	ProfileID1,
	ProfileID2,
	ProfileID3,
	ProfileID4,
	ProfileID5,
	ProfileID6,
	ProfileID7,
	ProfileID8,
	ProfileID9,
	ProfileID10,
	ProfileID11,
	ProfileID12,
	ProfileID13,
	ProfileID14,
	ProfileName,
	Recipe1,
	Recipe2,
	Recipe3,
	Roles1,
	Roles2,
	Roles3,
	ReportUUID string
}

// Prepare some data to insert into the template.
type node struct {
	EntityUUID,
	NodeName,
	Hostnamestr,
	IpAddress,
	Hostname,
	EndTime,
	StartTime,
	Recipe1,
	Recipe2,
	Recipe3,
	ID,
	Platform,
	Roles,
	ChefEnvironment,
	Attr,
	NormalTags,
	OrganizationName,
	CookbookName,
	CookbookVersion,
	CookbookID,
	RunID string
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func String(length int) string {
	return StringWithCharset(length, charset)
}

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func generateIpAddress() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	data := fmt.Sprintf("%d.%d.%d.%d", r.Intn(255), r.Intn(255), r.Intn(255), r.Intn(255))
	return data
}

func randomserialNumber() string {
	serialNumber := uuid.NewV4()
	return fmt.Sprintf("%s", serialNumber)
}

func randInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
}

func main() {

	err := false
	numberOfElements := flag.Int("numberOfElement", 1, "number of elements")
	url := flag.String("url", "", "automate url")
	apiToken := flag.String("token", "", "automate token")
	dataType := flag.String("data", "node", "type of data to send: node or compliance")
	concurrency := flag.Int("concurrency", 1, "Number of concurrent requests to run")

	maxGoroutines := concurrency

	flag.Parse()
	if *numberOfElements <= 0 {
		err = true
		fmt.Println("Please enter valid number")
	}
	if *url == "" {
		err = true
		fmt.Println("url should not be empty")
	}
	if *apiToken == "" {
		err = true
		fmt.Println("token should not be empty")
	}
	if *dataType != "node" && *dataType != "compliance" {
		err = true
		fmt.Println("data should be either node or compliance")
	}

	if !err {
		makeRequest(*numberOfElements, *url, *apiToken, *dataType, *maxGoroutines)
	}

}

func makeRequest(numberOfElements int, endpoint string, apiToken string, dataType string, maxGoroutines int) {

	if numberOfElements < maxGoroutines {
		maxGoroutines = 1
	}
	guard := make(chan struct{}, maxGoroutines)
	doneChannel := make(chan bool)

	// endpoint := "https://ec2-13-233-86-54.ap-south-1.compute.amazonaws.com/api/v0/events/data-collector"
	// apiToken := "OTsmSLhhLLeSIrNN6AloGDykP-M="

	var nodeData = make([]node, numberOfElements)
	var complianceData = make([]comp, numberOfElements)

	if dataType == "node" {
		for i := 0; i < numberOfElements; i++ {
			nodeData[i] = node{
				randomserialNumber(),
				String(10),
				"ip-" + generateIpAddress(),
				generateIpAddress(),
				"testhost" + String(8),
				"2021-06-10T11:55:52Z",
				"2021-06-10T11:47:52Z",
				"nginx::default" + String(3),
				"tomcat::default" + String(3),
				"yum::default" + String(3),
				randomserialNumber(),
				"ubuntu" + String(2),
				"test" + String(3),
				"grow" + String(10),
				String(10),
				"network_device" + String(3),
				"The " + String(8),
				String(6),
				"0.1." + fmt.Sprintf("%d", randInt(0, 10)),
				"chef-" + String(5),
				randomserialNumber(),
			}
		}

		for _, r := range nodeData {
			guard <- struct{}{}
			go processTemplateAndSend(guard, r, doneChannel, endpoint, apiToken, "node.tmpl")
		}

	} else if dataType == "compliance" {
		for i := 0; i < numberOfElements; i++ {
			complianceData[i] = comp{
				"2.1." + fmt.Sprintf("%d", randInt(0, 10)),
				String(5),
				"2021-06-11T07:02:25Z",
				"DevSec " + String(3) + " " + String(5),
				"chef-test-violet-waxwing-yellow-" + String(6),
				randomserialNumber(),
				"debian",
				"8." + fmt.Sprintf("%d", randInt(10, 20)),
				"apache-01",
				"apache-02",
				"apache-03",
				"apache-04",
				"apache-05",
				"apache-06",
				"apache-07",
				"apache-08",
				"apache-09",
				"apache-10",
				"apache-11",
				"apache-12",
				"apache-13",
				"apache-14",
				"DevSec Apache Baseline" + String(6),
				"linux::harden",
				"tomcat" + String(5),
				"tomcat::setup" + String(5),
				"base_windows" + String(5),
				"windows-hardening" + String(5),
				"best.role.ever" + String(5),
				randomserialNumber(),
			}
		}

		for _, r := range complianceData {

			guard <- struct{}{}
			go processTemplateAndSend(guard, r, doneChannel, endpoint, apiToken, "compliance.tmpl")
		}
	}
	for i := 0; i < numberOfElements; i++ {
		<-doneChannel
	}

}

func processTemplateAndSend(guard chan struct{}, r interface{}, doneChannel chan bool, endpoint string, apiToken string, fileName string) {
	var tpl bytes.Buffer
	// Create a new template and parse the letter into it.
	tmpl := template.New(fileName)
	t, err := tmpl.ParseFiles(fileName)
	if err != nil {
		panic(err)
	}
	err = t.Execute(&tpl, r)
	if err != nil {
		log.Println("executing template:", err)
	}
	err = sendNotification(tpl.String(), endpoint, apiToken)
	if err != nil {
		fmt.Println("Error", err)
	}
	<-guard
	doneChannel <- true
}

func sendNotification(data string, url string, token string) error {

	startTime := time.Now().UnixNano() / 1000000
	var buffer bytes.Buffer
	for _, message := range data {
		data1, _ := json.Marshal(message)
		data1 = bytes.ReplaceAll(data1, []byte("\n"), []byte("\f"))
		buffer.Write(data1)
		buffer.WriteString("\n")
	}

	var contentBuffer bytes.Buffer
	zip := gzip.NewWriter(&contentBuffer)
	_, err := zip.Write(buffer.Bytes())
	if err != nil {
		return err
	}
	err = zip.Close()
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		fmt.Println("Error creating request")
		return err
	}
	request.Header.Add("api-token", token)
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Content-Encoding", "gzip")
	request.Header.Add("Accept", "*/*")

	// var client http.Client
	config := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: config}
	var acceptedStatusCodes = []int32{200, 201, 202, 203, 204}

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error sending message ", err)
		return err
	} else {
		endTime := time.Now().UnixNano() / 1000000
		fmt.Println("Asset data posted to ", url, "Status ", response.Status)
		bodyBytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Fatal(err)
		}
		bodyString := string(bodyBytes)
		fmt.Println("Message body", bodyString)
		fmt.Println("Time taken to send data to serviceNow is ", endTime-startTime, "Millisecond")

	}
	if !IsAcceptedStatusCode(int32(response.StatusCode), acceptedStatusCodes) {
		return errors.New(response.Status)
	}
	err = response.Body.Close()
	if err != nil {
		fmt.Println("Error closing response body", err)
	}
	return nil
}

func IsAcceptedStatusCode(statusCode int32, acceptedCodes []int32) bool {
	for _, code := range acceptedCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}
