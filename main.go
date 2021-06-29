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

const (
	COMPLIANCE = "compliance"
	ATTRIBUTES = "attributes"
	CLIENT_RUN = "client_run"
	NODE       = "node"
	REPORT     = "report"
	charset    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func main() {
	err := false
	numberOfNodes := flag.Int("numberOfNodes", 0, "number of nodes")
	url := flag.String("url", "", "serviceNow url")
	username := flag.String("username", "", "serviceNow username")
	password := flag.String("password", "", "serviceNow password")
	batchSize := flag.Int("batchSize", 1, "Batch size to send number of nodes")
	dataType := flag.String("data", "all", "type of data to send: node or compliance or all")
	concurrency := flag.Int("concurrency", 1, "Number of concurrent requests to run")

	maxGoroutines := concurrency

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
	if *dataType != "all" && *dataType != "node" && *dataType != "compliance" {
		err = true
		fmt.Println("data should be either node or compliance or all")
	}

	if !err {
		batchSizeRequest(*numberOfNodes, *url, *username, *password, *batchSize, *dataType, *maxGoroutines)
	}
}
func replaceAndAppend(res *[]map[string]interface{}, doneChannel chan bool, index int, stringJsonData string, datatype string) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(ReplaceData(stringJsonData)), &data)
	if err != nil {
		log.Fatal("unexpected error")
	}
	(*res)[index] = map[string]interface{}{}
	if datatype == COMPLIANCE {
		(*res)[index][REPORT] = data[REPORT]
	} else if datatype == NODE {
		(*res)[index][ATTRIBUTES] = data[ATTRIBUTES]
		(*res)[index][CLIENT_RUN] = data[CLIENT_RUN]
		(*res)[index][NODE] = data[NODE]
	} else {
		(*res)[index] = data
	}
	doneChannel <- true
}

func homePage(numberOfNodes int, url string, username string, password string, datatype string, doneChannel chan bool, guard chan struct{}) {

	file, _ := ioutil.ReadFile("payload.json")
	stringJsonData := string(file)

	doneChannelReplaceAppend := make(chan bool)
	resultSlice := make([]map[string]interface{}, numberOfNodes)
	for i := 0; i < numberOfNodes; i++ {
		go replaceAndAppend(&resultSlice, doneChannelReplaceAppend, i, stringJsonData, datatype)
	}

	for i := 0; i < numberOfNodes; i++ {
		<-doneChannelReplaceAppend
	}

	fmt.Println("Node size per request :::::: ", numberOfNodes)
	sendNotification(resultSlice, url, username, password)
	// time.Sleep(time.Second * 2)
	<-guard
	doneChannel <- true
}

func batchSizeRequest(numberOfNodes int, url string, username string, password string, batchSize int, datatype string, maxGoroutines int) {
	doneChannelBatchSize := make(chan bool)
	numberOfBatchRequest := int(math.Trunc(float64(numberOfNodes / batchSize)))
	singleRequest := numberOfNodes % batchSize
	if singleRequest != 0 {
		numberOfBatchRequest++
	}

	if numberOfBatchRequest < maxGoroutines {
		fmt.Println("Since number Of batches " + fmt.Sprintf("%d", numberOfBatchRequest) + " is less than concurrency " + fmt.Sprintf("%d", maxGoroutines) + ". So using comcurrency of 1.")
		maxGoroutines = 1
	}

	guard := make(chan struct{}, maxGoroutines)

	for i := 0; i < numberOfBatchRequest; i++ {
		guard <- struct{}{}
		if i == numberOfBatchRequest-1 && singleRequest != 0 {
			go homePage(singleRequest, url, username, password, datatype, doneChannelBatchSize, guard)
		} else {
			go homePage(batchSize, url, username, password, datatype, doneChannelBatchSize, guard)
		}
	}
	for i := 0; i < numberOfBatchRequest; i++ {
		<-doneChannelBatchSize
	}
}

func randomNodeId() string {
	NodeId := uuid.NewV4()
	return fmt.Sprintf("%s", NodeId)
}
func randomserialNumber() string {
	serialNumber := uuid.NewV4()
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
	ReplaceHostName := strings.ReplaceAll(ReplaceSerialNumber, "VA1IHGDLOY08", String(12))
	ReplaceReportID := strings.ReplaceAll(ReplaceHostName, "9049e316-ebed-4572-9368-9a26c781e8da", randomserialNumber())
	return ReplaceReportID
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

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}
