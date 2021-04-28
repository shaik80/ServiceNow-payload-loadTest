package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func main() {
	numberOfNodes := flag.Int("numberOfNodes", 0, "number of nodes")
	url := flag.String("url", "", "serviceNow url")
	username := flag.String("username", "", "serviceNow username")
	password := flag.String("password", "", "serviceNow password")
	flag.Parse()
	fmt.Println(*numberOfNodes, *url, *username, *password)
	if *numberOfNodes <= 0 {
		fmt.Println("Please enter valid number")
	} else if *url == "" {
		fmt.Println("url should not be empty")
	} else if *username == "" {
		fmt.Println("username should not be empty")
	} else if *password == "" {
		fmt.Println("password should not be empty")
	} else {
		homePage(*numberOfNodes, *url, *username, *password)
	}
}
func replaceAndAppend(res *[]interface{}, doneChannel chan bool, index int, stringJsonData string) {
	var data interface{}
	err := json.Unmarshal([]byte(ReplaceData(stringJsonData)), &data)
	if err != nil {
		log.Fatal("unexpected error")
	}
	(*res)[index] = data
	doneChannel <- true
}

func homePage(numberOfNodes int, url string, username string, password string) {

	file, _ := ioutil.ReadFile("payload.json")
	stringJsonData := string(file)

	fmt.Println("numberOfNodes", numberOfNodes)
	doneChannel := make(chan bool)
	resultSlice := make([]interface{}, numberOfNodes)
	for i := 0; i < numberOfNodes; i++ {
		go replaceAndAppend(&resultSlice, doneChannel, i, stringJsonData)
	}

	for i := 0; i < numberOfNodes; i++ {
		<-doneChannel
	}
	jsonResponse, jsonError := json.Marshal(resultSlice)
	if jsonError != nil {
		fmt.Println("Unable to encode JSON")
	}
	sendNotification(jsonResponse, url, username, password)

}
func randomNodeId() string {
	NodeId := uuid.NewV4()
	return fmt.Sprintf("%s", NodeId)
}
func randomserialNumber() string {
	serialNumber := uuid.NewV5(uuid.UUID{}, "vm")
	return fmt.Sprintf("%s", serialNumber)
}

func generateIpAddress() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	data := fmt.Sprintf("%d.%d.%d.%d", r.Intn(255), r.Intn(255), r.Intn(255), r.Intn(255))

	return data
}

func ReplaceData(data string) string {
	// node data
	ReplaceNodeId := strings.ReplaceAll(data, "1304ecea-95bd-4830-b8c9-2cbb33555695", randomNodeId())
	ReplaceIpAddress := strings.ReplaceAll(ReplaceNodeId, "10.127.75.100", generateIpAddress())
	ReplaceSerialNumber := strings.ReplaceAll(ReplaceIpAddress, "ec2f999c-e79a-0454-6a18-d9942ab11f77", randomserialNumber())
	return ReplaceSerialNumber
}

func sendNotification(data []byte, url string, username string, password string) error {

	var contentBuffer bytes.Buffer
	zip := gzip.NewWriter(&contentBuffer)
	_, err := zip.Write(data)
	if err != nil {
		return err
	}
	err = zip.Close()
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", url, &contentBuffer)
	if err != nil {
		fmt.Println("Error creating request")
		return err
	}
	request.Header.Add("Authorization", basicAuth(username, password))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Content-Encoding", "gzip")
	request.Header.Add("Accept", "*/*")

	var client http.Client
	var acceptedStatusCodes = []int32{200, 201, 202, 203, 204}

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("Error sending message ", err)
		return err
	} else {
		fmt.Println("Asset data posted to ", url, "Status ", response.Status)
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
func basicAuth(username string, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func IsAcceptedStatusCode(statusCode int32, acceptedCodes []int32) bool {
	for _, code := range acceptedCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}
