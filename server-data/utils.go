package main

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"text/template"
	"time"

	uuid "github.com/satori/go.uuid"
)

var counter = 0

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

func String(length int) string {
	return StringWithCharset(length, charset)
}

func getFileArr(path string) []string {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	var fileNames []string
	for _, f := range files {
		fileNames = append(fileNames, f.Name())
	}
	return fileNames
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

func pickStatus() string {
	status := []string{"success", "failure"}
	val := randInt(0, 2)
	return status[val]
}

func pickSuccessOrFailureFile(filearr []string) string {
	val := randInt(0, len(filearr))
	return filearr[val]
}

func makeRequest(numberOfElements int, endpoint string, apiToken string, dataType string, status string, maxGoroutines int) {

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
				pickStatus(),
				randomserialNumber(),
				String(10),
				"ip-" + generateIpAddress(),
				generateIpAddress(),
				"testhost" + String(8),
				currentTimestamp(),
				currentTimestamp(),
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
				"0.1." + fmt.Sprintf("%d", randInt(0, 11)),
				"chef-" + String(5),
				randomserialNumber(),
			}
		}

		for _, r := range nodeData {
			guard <- struct{}{}
			// nodefiles := getFileArr(nodeDataFolder)
			file := pickProperFile(dataType, status)
			fmt.Println("file used for node", file)
			go processTemplateAndSend(guard, r, doneChannel, endpoint, apiToken, nodeDataFolder, file)
		}

	} else if dataType == "compliance" {
		for i := 0; i < numberOfElements; i++ {
			complianceData[i] = comp{
				"2.1." + fmt.Sprintf("%d", randInt(0, 11)),
				String(5),
				currentTimestamp(),
				"DevSec " + String(3) + " " + String(5),
				"chef-test-violet-waxwing-yellow-" + String(6),
				randomserialNumber(),
				"debian",
				"8." + fmt.Sprintf("%d", randInt(10, 21)),
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
			// complianceFiles := getFileArr(complianceDataFolder)
			// file := pickSuccessOrFailureFile(complianceFiles)
			file := pickProperFile(dataType, status)
			fmt.Println("file used for compliance", file)
			go processTemplateAndSend(guard, r, doneChannel, endpoint, apiToken, complianceDataFolder, file)
		}
	}
	for i := 0; i < numberOfElements; i++ {
		<-doneChannel
	}

}

func pickProperFile(dataType string, status string) string {
	if dataType == "node" {
		return nodeFile[status]
	} else {
		return complianceFile[status]
	}
}

func currentTimestamp() string {
	// randomTime := rand.Int63n(time.Now().Unix()-94608000) + 94608000
	timeNow := time.Now().Unix()
	now := time.Unix(timeNow, 0)
	loc, err := time.LoadLocation("UTC")
	if err != nil {
		fmt.Println(err)
		return "2021-06-14T07:02:25Z"
	}
	const DateTimeFormat = "2006-01-02T15:04:05Z"
	return now.In(loc).Format(DateTimeFormat)
}

func processTemplateAndSend(guard chan struct{}, r interface{}, doneChannel chan bool, endpoint string, apiToken string, folder string, fileName string) {
	var tpl bytes.Buffer
	tmpl := template.New(fileName)
	t, err := tmpl.ParseFiles(folder + "/" + fileName)
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
	counter++
	fmt.Println("Total done", counter)
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

	// fmt.Println(data)

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
