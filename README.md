# ServiceNow-payload-loadTest
[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)

**Note :** Make sure to download and install golang
### How to use this repo
1. Clone Repo by entering this command `git clone https://github.com/shaik80/ServiceNow-payload-loadTest.git`
2. Then go inside the repository and run `go run main.go`

### How to run

There are 2 ways:

1. Directly using by main.go (for development)
    *  Run `go run main.go -h`, It shows the parameter to pass in command
    * To send data to serviceNow run `go run main.go -numberOfNodes=1 -url=https://86e66aab97d5.ngrok.io/request -username=admin -password=password`
2. By using build file (for testing)
    * [click here](https://github.com/shaik80/ServiceNow-payload-loadTest/raw/main/ServiceNowLoadTest "ServiceNow load test build file") to download build file
    * run `./main -h`, It shows the parameter to pass in command
        ```
        -numberOfNodes int
            number of nodes
        -password string
            serviceNow password
        -url string
            serviceNow url
        -username string
            serviceNow username
        ```
    * To send data to serviceNow run `./main -numberOfNodes=1 -url=https://86e66aab97d5.ngrok.io/request -username=admin -password=password`
