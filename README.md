# Bored Board Service

## Was Ist Das?

It's the start of a Service/REST API/Whatever you want to call it for the front of the new Bored Board to consume.

It's written in Golang, so here is some handy setup instructions:

### Local Dev Sans Docker (EXPERT MODE)

* Install Golang
* Setup a GOPATH that makes sense, and get this project setup there
* Install `dep` to manage Golang dependencies
* Run `dep ensure` from the root to get necessary dependencies setup
* Run `go run main.go` and the app should start
* Using Postman, etc... you can send a `GET` request to `http://localhost:8000/thread` and you'll get a test response if everything is working

### Docker Docker Docker (Moby wasn't in Flipper)

* Ensure you have Docker installed
* `docker-compose up` in the root
* This should get everything up and going, and you should be able to make code changes on the fly

## Can I contribute?

Yes, please. File an issue to let us know what you are working on, and then submit a PR that associates with your issue. 

## Got A Problem?

280-330-8004
