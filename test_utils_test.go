package gocd

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

const testUsername = "admin"
const testPassword = "badger"
const testAPIToken = "api_token"

func DummyRequestBodyValidator(body string) error {
	return nil
}

func newTestAPIClient(route string, handler func(http.ResponseWriter, *http.Request)) (Client, *httptest.Server) {
	newTestServerHandler := http.NewServeMux()
	newTestServerHandler.HandleFunc(route, handler)
	newAPIServer := httptest.NewServer(newTestServerHandler)

	//return New(newAPIServer.URL, testUsername, testPassword), newAPIServer
	return NewTokenBased(newAPIServer.URL, testAPIToken), newAPIServer
}

func serveFileAsJSON(t *testing.T, method string, filepath string, apiVersion int, requestBodyValidator func(string) error) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		// log.Println("Doing AcceptHeaderCheck")
		if apiVersion > 0 {
			AcceptHeaderCheck(t, apiVersion, request)
		}
		// log.Println("Doing BasicAuthCheck")
		AuthCheck(t, request)
		// log.Println("Doing RequestMethodCheck with " + method)
		RequestMethodCheck(t, request, method)
		// log.Println("Doing RequestBodyCheck")
		RequestBodyCheck(t, request, requestBodyValidator)

		contents, err := ioutil.ReadFile(filepath)
		if err != nil {
			log.Fatal(err)
		}
		if apiVersion > 0 {
			writer.Header().Add("Content-Type", fmt.Sprintf("application/vnd.go.cd.v%d+json; charset=utf-8", apiVersion))
		}
		writer.Write(contents)
	}
}

func serveFileAsXML(t *testing.T, method, filepath string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		AuthCheck(t, request)
		RequestMethodCheck(t, request, method)

		contents, err := ioutil.ReadFile(filepath)
		if err != nil {
			log.Fatal(err)
		}
		writer.Header().Add("Content-Type", "application/xml; charset=utf-8")
		writer.Write(contents)
	}
}

func AcceptHeaderCheck(t *testing.T, apiVersion int, request *http.Request) {
	// Accept Header check
	acceptHeader := request.Header.Get("Accept")
	if acceptHeader != fmt.Sprintf("application/vnd.go.cd.v%d+json", apiVersion) {
		log.Fatalf("We did not receive Accept: application/vnd.go.cd.v%d+json header in the request", apiVersion)
	}
}

func AuthCheck(t *testing.T, request *http.Request) {
	// BasicAuth check
	username, password, ok := request.BasicAuth()
	if ok {
		if username != testUsername && password != testPassword {
			log.Fatalf("Invalid username / password combination")
		}
	} else if request.Header.Get("Authorization") != "bearer "+testAPIToken {
		log.Fatalf("Invalid api token, expected \"%s\" but was \"%s\"", "bearer "+testAPIToken, request.Header.Get("Authorization"))
	}
}

func RequestMethodCheck(t *testing.T, request *http.Request, method string) {
	if request.Method != method {
		log.Fatalf("Expected HTTP method is %s while client sent %s", method, request.Method)
	}
}

func RequestBodyCheck(t *testing.T, request *http.Request, requestBodyValidator func(string) error) {
	bytes, err := ioutil.ReadAll(request.Body)
	// log.Printf("%v\n", request.Body)
	if err != nil {
		log.Fatalf("%v\n", err)
	} else if err := requestBodyValidator(string(bytes)); err != nil {
		log.Fatalf("%v\n", err)
	}
}
