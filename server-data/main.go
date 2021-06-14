// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
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
	Status,
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

const (
	nodeDataFolder        = "node-data"
	complianceDataFolder  = "compliance-data"
	nodeSuccessfile       = "successnode.tmpl"
	nodefailurefile       = "failurenode.tmpl"
	complianceSuccessFile = "compliance.tmpl"
	complianceFailureFile = "complainceStatus.tmpl"
)

var nodeFile = map[string]string{
	"success": "successnode.tmpl",
	"failure": "failurenode.tmpl",
}

var complianceFile = map[string]string{
	"success": "compliance.tmpl",
	"failure": "complainceStatus.tmpl",
}

func main() {

	err := false
	numberOfElements := flag.Int("numberOfElement", 1, "number of elements")
	url := flag.String("url", "", "automate url")
	apiToken := flag.String("token", "", "automate token")
	dataType := flag.String("data", "node", "type of data to send: node or compliance")
	concurrency := flag.Int("concurrency", 1, "Number of concurrent requests to run")
	status := flag.String("status", "", "give status type: success or failure")

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
	if *status != "success" && *status != "failure" {
		err = true
		fmt.Println("data should be either node or compliance")
	}

	if !err {
		makeRequest(*numberOfElements, *url, *apiToken, *dataType, *status, *maxGoroutines)
	}

}
