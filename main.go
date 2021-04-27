package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
)

func main() {
	fmt.Println(len(os.Args), os.Args)
	if len(os.Args) < 5 {
		fmt.Println("Please pass all the arguments: numberOfNodes, url, username, password")
		return
	}
	// handleRequest()
	homePage(os.Args)
}

// func handleRequest() {

// 	myRoute := mux.NewRouter().StrictSlash(true)
// 	// Declaration of api route
// 	// homePage is func homePage which is called in handle func
// 	myRoute.HandleFunc("/{id}", homePage).Methods("GET")
// 	// Declaration of server and port
// 	log.Println("listening on", 8081)
// 	err := http.ListenAndServe(":8081", myRoute)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// }

func homePage(args []string) {
	numberOfNodes, _ := strconv.Atoi(args[1])
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

	fmt.Println(string(jsonResponse))

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

// func homePage(w http.ResponseWriter, r *http.Request) {
// 	numberOfNodes, _ := strconv.Atoi(mux.Vars(r)["id"])
// 	file, _ := ioutil.ReadFile("payload.json")
// 	stringJsonData := string(file)

// 	// var res []interface{}

// 	doneChannel := make(chan bool)
// 	resultSlice := make([]interface{}, numberOfNodes)
// 	for i := 0; i < numberOfNodes; i++ {
// 		go replaceAndAppend(&resultSlice, doneChannel, i, stringJsonData)
// 	}

// 	for i := 0; i < numberOfNodes; i++ {
// 		<-doneChannel
// 	}
// 	jsonResponse, jsonError := json.Marshal(resultSlice)
// 	if jsonError != nil {
// 		fmt.Println("Unable to encode JSON")
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write(jsonResponse)

// }

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
