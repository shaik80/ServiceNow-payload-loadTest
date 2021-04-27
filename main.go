package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

func main() {
	handleRequest()
}

type Payload struct {
	res []interface{}
}

func handleRequest() {

	myRoute := mux.NewRouter().StrictSlash(true)
	// Declaration of api route
	// homePage is func homePage which is called in handle func
	myRoute.HandleFunc("/{id}", homePage).Methods("GET")
	// Declaration of server and port
	err := http.ListenAndServe(":8081", myRoute)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Fatal("Server starting on port 8081......")
	}

}
func homePage(w http.ResponseWriter, r *http.Request) {
	numberOfNodes, _ := strconv.Atoi(mux.Vars(r)["id"])
	file, _ := ioutil.ReadFile("payload.json")
	stringJsonData := string(file)

	var data []interface{}
	res := &Payload{}

	for i := 0; i < numberOfNodes; i++ {
		_ = json.Unmarshal([]byte(ReplaceData(stringJsonData)), &data)
		res.res = append(res.res, data)
	}

	jsonResponse, jsonError := json.Marshal(res.res)
	if jsonError != nil {
		fmt.Println("Unable to encode JSON")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)

}
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
