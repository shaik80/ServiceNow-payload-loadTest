# ServiceNow-payload-loadTest
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)

**Note :** Make sure to download and install golang

### How to use this repo
1. Clone Repo by entering this command `git clone https://github.com/shaik80/ServiceNow-payload-loadTest.git`

### How to run

There are 2 ways:

1. Use the build file
    * [click here](https://github.com/shaik80/ServiceNow-payload-loadTest/raw/chef-server/server-data/ServiceNowLoadTest "Project root") Go here after cloning the repository
    * run `./ServiceNowLoadTest -h`, It shows the parameter to pass in command
        ```
        -concurrency int
            Number of concurrent requests to run (default 1)
        -data string
            type of data to send: node or compliance (default "node")
        -numberOfElement int
            number of elements (default 1)
        -status string
            give status type: success or failure
        -token string
            automate token
        -url string
            automate url
        -useLarge
            use larger json file: true or false, does not work if status is set
        ```
    * If u want to send Node Data to automate run `./ServiceNowLoadTest -url https://ec2-65-1-113-189.ap-south-1.compute.amazonaws.com/api/v0/events/data-collector -token WC3OTNWT0-y576gTS5OudtAKIR8= -numberOfElement=50000 -concurrency=200 -data=node -useLarge=true`,
    * If u want to send Complaince Data to automate run `./ServiceNowLoadTest -url https://ec2-65-1-113-189.ap-south-1.compute.amazonaws.com/api/v0/events/data-collector -token WC3OTNWT0-y576gTS5OudtAKIR8= -numberOfElement=50000 -concurrency=200 -data=complaince -useLarge=true`,
    * useLarge will auto select the larger data of complaince or node
    * status success or failure will work if useLarge is not set.

**Note :** if you have multiple files downloaded then replace file name with `ServiceNowLoadTest` then u can follow secound point
