package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

func main() {
	handleRequest()
}

func handleRequest() {

	myRoute := mux.NewRouter().StrictSlash(true)
	// Declaration of api route
	// homePage is func homePage which is called in handle func
	myRoute.HandleFunc("/", homePage).Methods("POST")
	// Declaration of server and port
	log.Println("listening on", 8081)
	err := http.ListenAndServe(":8081", myRoute)
	if err != nil {
		log.Fatal(err)
	}

}

func replaceAndAppend(res *[]interface{}, doneChannel chan bool, index int, stringJsonData string) {
	var data interface{}
	err := json.Unmarshal([]byte(ReplaceData(stringJsonData)), &data)
	if err != nil {
		log.Fatal("unexpected error")
	}
	// res = append(res, data)
	(*res)[index] = data
	doneChannel <- true
}

func homePage(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var t map[string]interface{}
	err := decoder.Decode(&t)

	if err != nil {
		panic(err)
	}

	numberOfNodes := t["id"].(int)
	url := t["url"]
	username := t["username"]
	password := t["password"]

	file, _ := ioutil.ReadFile("payload.json")
	stringJsonData := string(file)

	// var res []interface{}

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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}

// func homePage(w http.ResponseWriter, r *http.Request) {
// 	numberOfNodes, _ := strconv.Atoi(mux.Vars(r)["id"])
// 	file, _ := ioutil.ReadFile("payload.json")
// 	stringJsonData := string(file)

// 	var res []interface{}

// 	for i := 0; i < numberOfNodes; i++ {
// 		var data interface{}
// 		err := json.Unmarshal([]byte(ReplaceData(stringJsonData)), &data)
// 		if err != nil {
// 			log.Fatal("unexpected error")
// 			break
// 		}
// 		res = append(res, data)
// 	}

// 	jsonResponse, jsonError := json.Marshal(res)
// 	if jsonError != nil {
// 		fmt.Println("Unable to encode JSON")
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write(jsonResponse)

// }

func randomNodeId() string {
	NodeId := uuid.NewV4()
	fmt.Println("Your UUIDv4 is: %s", NodeId)
	return fmt.Sprintf("%s", NodeId)
}
func randomserialNumber() string {
	serialNumber := uuid.NewV5(uuid.UUID{}, "vm")
	fmt.Println("Your UUIDv4 is: %s", serialNumber)
	return fmt.Sprintf("%s", serialNumber)
}

func generateIpAddress() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	data := fmt.Sprintf("%d.%d.%d.%d", r.Intn(255), r.Intn(255), r.Intn(255), r.Intn(255))
	fmt.Println("your Ipaddress %s", data)
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
		fmt.Println("Error sending message %v", err)
		return err
	} else {
		fmt.Println("Asset data posted to %v, Status %v", url, response.Status)
	}
	if !IsAcceptedStatusCode(int32(response.StatusCode), acceptedStatusCodes) {
		return errors.New(response.Status)
	}
	err = response.Body.Close()
	if err != nil {
		fmt.Println("Error closing response body %v", err)
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
