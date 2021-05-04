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
	"math"
	"math/rand"
	"net/http"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func main() {
	err := false
	numberOfNodes := flag.Int("numberOfNodes", 0, "number of nodes")
	url := flag.String("url", "", "serviceNow url")
	username := flag.String("username", "", "serviceNow username")
	password := flag.String("password", "", "serviceNow password")
	batchSize := flag.Int("batchSize", 1, "Batch size to send number of nodes")
	flag.Parse()
	if *numberOfNodes <= 0 {
		err = true
		fmt.Println("Please enter valid number")
	}
	if *url == "" {
		err = true
		fmt.Println("url should not be empty")
	}
	if *username == "" {
		err = true
		fmt.Println("username should not be empty")
	}
	if *password == "" {
		err = true
		fmt.Println("password should not be empty")
	}
	if *batchSize <= 0 {
		err = true
		fmt.Println("Please enter valid number")
	}

	if !err {
		batchSizeRequest(*numberOfNodes, *url, *username, *password, *batchSize)
	}
}
func replaceAndAppend(res *[]map[string]interface{}, doneChannel chan bool, index int, stringJsonData string) {
	var data map[string]interface{}
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

	doneChannelReplaceAppend := make(chan bool)
	resultSlice := make([]map[string]interface{}, numberOfNodes)
	for i := 0; i < numberOfNodes; i++ {
		go replaceAndAppend(&resultSlice, doneChannelReplaceAppend, i, stringJsonData)
	}

	for i := 0; i < numberOfNodes; i++ {
		<-doneChannelReplaceAppend
	}
	// jsonResponse, jsonError := json.Marshal(resultSlice)
	// if jsonError != nil {
	// 	fmt.Println("Unable to encode JSON")
	// }
	fmt.Println(":::::: Node size per request :::::: ", numberOfNodes)
	sendNotification(resultSlice, url, username, password)
	// doneChannel <- true
}

func batchSizeRequest(numberOfNodes int, url string, username string, password string, batchSize int) {
	// doneChannelBatchSize := make(chan bool)

	numberOfBatchRequest := math.Trunc(float64(numberOfNodes / batchSize))

	singleRequest := numberOfNodes % batchSize

	for i := 0; i < int(numberOfBatchRequest); i++ {
		homePage(batchSize, url, username, password)
	}
	// for i := 0; i < int(numberOfBatchRequest); i++ {
	// 	<-doneChannelBatchSize
	// }
	if singleRequest != 0 {
		homePage(singleRequest, url, username, password)
	}
	// if singleRequest != 0 {
	// 	<-doneChannelBatchSize
	// }
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

func sendNotification(data []map[string]interface{}, url string, username string, password string) error {

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
		endTime := time.Now().UnixNano() / 1000000
		fmt.Println("Asset data posted to ", url, "Status ", response.Status)
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
