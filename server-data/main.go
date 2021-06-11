// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"time"
)

func main() {

	type comp struct {
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
	type Recipient struct {
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
	var recipients = []Recipient{
		{
			"hostname",
			"2021-06-10T11:55:52Z",
			"2021-06-10T11:47:52Z",
			"nginx::default",
			"tomcat::default",
			"yum::default",
			"c2fb9ee4-13ec-4c48-a495-8d6a52710e33",
			"ubuntu",
			"extra_who_died_on_ER",
			"grassland",
			"something",
			"network_device",
			"The Defenders",
			"nginx", "0.1.0",
			"chef-sugar",
			"c2fb9ee4-13ec-4c48-a495-8d6a52710e33",
		},
	}

	var data = []comp{
		{
			"2021-06-11T07:02:25Z",
			"DevSec Dev_Gamma",
			"chef-load-violet-waxwing-yellow",
			"2a463196-970c-3ecc-97bc-46945d0844da",
			"debian",
			"8.11",
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
			"DevSec Apache Baseline",
			"linux::harden",
			"tomcat",
			"tomcat::setup",
			"base_windows",
			"windows-hardening",
			"best.role.ever",
			"5fb3688a-90e3-4571-bed6-0067432d1878",
		},
	}

	// Create a new template and parse the letter into it.
	tmpl := template.New("node.tmpl")

	t, err := tmpl.ParseFiles("node.tmpl")
	if err != nil {
		panic(err)
	}

	// Execute the template for each recipient.
	var tpl bytes.Buffer
	for _, r := range recipients {
		err := t.Execute(&tpl, r)
		if err != nil {
			log.Println("executing template:", err)
		}
	}

	fmt.Println(tpl.String())

	// Create a new template and parse the letter into it.
	tmplc := template.New("compliance.tmpl")

	tr, errr := tmplc.ParseFiles("compliance.tmpl")
	if errr != nil {
		panic(err)
	}

	// Execute the template for each recipient.
	var tplc bytes.Buffer
	for _, r := range data {
		err := tr.Execute(&tplc, r)
		if err != nil {
			log.Println("executing template:", err)
		}
	}

	err = sendNotification(tplc, "https://a2-dev.test:2000/api/v0/events/data-collector", "6CYotJxooVPotCVryLxgPJ10zQQ=")
	if err != nil {
		fmt.Println("Error", err)
	}

	// fmt.Println(tplc.String())

}

func sendNotification(buffer bytes.Buffer, url string, token string) error {

	startTime := time.Now().UnixNano() / 1000000
	// var buffer bytes.Buffer
	// for _, message := range data {
	// 	data1, _ := json.Marshal(message)
	// 	data1 = bytes.ReplaceAll(data1, []byte("\n"), []byte("\f"))
	// 	buffer.Write(data1)
	// 	buffer.WriteString("\n")
	// }

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
	request.Header.Add("api-token", token)
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

func IsAcceptedStatusCode(statusCode int32, acceptedCodes []int32) bool {
	for _, code := range acceptedCodes {
		if code == statusCode {
			return true
		}
	}
	return false
}
