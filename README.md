<!-- 
Copyright 2017 Jeffry Hesse

Licensed under the Apache License, Version 2.0 (the "License"); 
you may not use this file except in compliance with the License. 
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software 
distributed under the License is distributed on an "AS IS" BASIS, 
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. 
See the License for the specific language governing permissions and 
limitations under the License.  
-->

# Bored Board Service

[![Build Status](https://travis-ci.org/DarthHater/bored-board-service.svg?branch=master)](https://travis-ci.org/DarthHater/bored-board-service)

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
