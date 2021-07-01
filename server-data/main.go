// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"errors"
	"flag"

	"gopkg.in/gookit/color.v1"
)

const (
	nodeDataFolder        = "node-data"
	complianceDataFolder  = "compliance-data"
	nodeSuccessfile       = "successnode.tmpl"
	nodefailurefile       = "failurenode.tmpl"
	complianceSuccessFile = "compliance.tmpl"
	complianceFailureFile = "complainceStatus.tmpl"
)

var nodeFile = map[string]string{
	"success": nodeSuccessfile,
	"failure": nodefailurefile,
	"large":   nodeSuccessfile,
}

var complianceFile = map[string]string{
	"success": complianceSuccessFile,
	"failure": complianceFailureFile,
	"large":   complianceFailureFile,
}

func main() {

	var err error
	numberOfElements := flag.Int("numberOfElement", 1, "number of elements")
	url := flag.String("url", "", "automate url")
	apiToken := flag.String("token", "", "automate token")
	dataType := flag.String("data", "node", "type of data to send: node or compliance")
	concurrency := flag.Int("concurrency", 1, "Number of concurrent requests to run")
	status := flag.String("status", "", "give status type: success or failure")
	large := flag.Bool("useLarge", false, "use larger json file: true or false, does not work if status is set")
	output := flag.String("o", "", "output file name eg: fileName.json")
	input := flag.String("i", "", "input file name eg: fileName.json")

	maxGoroutines := concurrency

	flag.Parse()
	if *input != "" || *output != "" {
		if *input == *output {
			err = errors.New("both fileName can not be same")
		}
	}

	if *numberOfElements <= 0 {
		err = errors.New("Please enter valid number")
	}
	if *url == "" {
		err = errors.New("url should not be empty")
	}
	if *apiToken == "" {
		err = errors.New("token should not be empty")
	}
	if *dataType != "node" && *dataType != "compliance" {
		err = errors.New("data should be either node or compliance")
	}
	if *status != "success" && *status != "failure" && *status != "" {
		err = errors.New("data should be either node or compliance or blank")
	}

	if *status == "success" || *status == "failure" {
		if *large {
			color.Warn.Println("Option for useLarge wont work when status option is set")
		}
	}

	if *status != "success" && *status != "failure" && !*large {
		err = errors.New("status and useLarge both cannot be blank and false at the same time")
	}

	// if *input != "" {
	// 	err = true
	// 	var fileName = *input + ".json"
	// 	if _, err := os.Stat("./Files/" + fileName); os.IsNotExist(err) {
	// 		color.Error.Println(err)
	// 	}
	// }
	if err != nil {
		color.Error.Println(err)
		return
	}

	makeRequest(*numberOfElements, *url, *apiToken, *dataType, *status, *large, *maxGoroutines, *input, *output)

}
