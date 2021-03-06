package main

import (
  "fmt"
  "net/http"
  "strings"
)

// Response is a struct representing the outcome of an attack.
type Response struct {
  Passed bool
  Report string
  Expected string
  Actual string
  AttackConfig AttackConfig
}

func checkHTTPResponse(httpResponse *http.Response, config AttackConfig) (bool, string, string, string) {
  if strings.Trim(httpResponse.Status, " ") != strings.Trim(config.ExpectedStatus, " ") {
    reason := fmt.Sprintf("Invalid status code of %s detected. Expected %s", httpResponse.Status, config.ExpectedStatus)
    return false, reason, config.ExpectedStatus, httpResponse.Status
  }

  return true, "", "", ""
}

func checkHTTPResponses(httpResponses []*http.Response, config AttackConfig) (bool, string, string, string) {
  for _,httpResp := range httpResponses {

    if httpResp == nil {
      return false, "Error occurred during HTTP request", "A valid HTTP Response", "No HTTP Response"
    }

    passed, reason, expected, actual := checkHTTPResponse(httpResp, config)

    if !passed {
      return passed, reason, expected, actual
    }
  }

  return true, "", "", ""
}

func dispatchMultipleHTTPRequests(endpoint string, c chan *http.Response, count int, method string) {
  for i := 0; i < count; i++ {
    if method == "" {
      SendRandomHTTPRequest(endpoint, c)
    } else {
      SendHTTPRequest(endpoint, c, method)
    }
  }
}

func dispatchConcurrentHTTPRequests(concurrentCount int, endpoint string, c chan *http.Response, count int, method string) {
  for i:=0; i < concurrentCount; i++ {
    go dispatchMultipleHTTPRequests(endpoint, c, count, method)
  }
}

func collectConcurrentHTTPResponses(c chan *http.Response, expectedCount int) []*http.Response {
  responses := []*http.Response{}

  for len(responses) < (expectedCount) {
    responses = readResponseFromChannel(responses, c)

    if responses == nil {
      return nil
    }
  }

  return responses
}

func readResponseFromChannel(responses []*http.Response, c chan *http.Response) []*http.Response {
  response := <- c

  if response == nil {
    return nil
  }

  defer response.Body.Close()
  return append(responses, response)
}

// RunHTTPSpam fires off the requested number of concurrent messages at an endpoint and tests response.
func RunHTTPSpam(endpointConfig EndpointConfig, attackConfig AttackConfig, responseChannel chan Response) error {
  c := make(chan *http.Response)

  endpoint := BuildNetworkPath(endpointConfig.Protocol, endpointConfig.Host, endpointConfig.Port, endpointConfig.Path)

  messageCount := attackConfig.Concurrents * attackConfig.MessagesPerConcurrent

  dispatchConcurrentHTTPRequests(attackConfig.Concurrents, endpoint, c, attackConfig.MessagesPerConcurrent, attackConfig.Method)
  responses := collectConcurrentHTTPResponses(c, messageCount)

  if len(responses) == 0 {
    responseChannel <- Response{AttackConfig: attackConfig, Passed: false, Report: "Error occurred during HTTP Spam."}
    return nil
  }

  passed, reason, expected, actual := checkHTTPResponses(responses, attackConfig)

  if !passed {
    responseChannel <- Response{Expected: expected, Actual: actual, AttackConfig: attackConfig, Passed: false, Report: fmt.Sprintf("Failure during HTTP Spam. %s", reason)}
    return nil
  }

  responseChannel <- Response{AttackConfig: attackConfig, Passed: true}
  return nil
}

// RunCorruptHTTP fires off a Corrupted HTTP request at the specific endpoint.
func RunCorruptHTTP(endpointConfig EndpointConfig, attackConfig AttackConfig, responseChannel chan Response) error {
  c := make(chan string)
  endpoint := BuildNetworkPath("", endpointConfig.Host, endpointConfig.Port, "")

  go SendCorruptHTTPData(endpoint, c)
  rawResponse := <- c

  if rawResponse == "" {
    responseChannel <- Response{AttackConfig: attackConfig, Passed: false, Report: "Expected valid response but got empty String."}
    return nil
  }

  if !strings.Contains(rawResponse, attackConfig.ExpectedStatus) {
    responseChannel <- Response{AttackConfig: attackConfig, Passed: false, Report: fmt.Sprintf("Expected Status = %s | Actual Response = %s", attackConfig.ExpectedStatus, rawResponse)}
  }

  responseChannel <- Response{AttackConfig: attackConfig, Passed: true, Report: fmt.Sprintf("Corrupt HTTP Test passed for endpoint %s", endpointConfig.Name)}
  return nil
}

// RunURLQuery hits an endpoint with a set of dodgy values in parameters.
func RunURLQuery(endpointConfig EndpointConfig, attackConfig AttackConfig, responseChannel chan Response) error {
  c := make(chan *http.Response)

  endpoint := BuildNetworkPath(endpointConfig.Protocol, endpointConfig.Host, endpointConfig.Port, endpointConfig.Path)

  params := strings.Split(attackConfig.Parameters, ",")

  fakeValues := []string{"hello", "&hello", "@", "\""}

  for _,param := range params {
    for _,value := range fakeValues {
      attackPoint := fmt.Sprintf("%s?%s=%s", endpoint, param, value)
      go SendHTTPRequest(attackPoint, c, "GET")
    }

  }

  responses := collectConcurrentHTTPResponses(c, len(params) * len(fakeValues))

  if len(responses) == 0 {
    responseChannel <- Response{AttackConfig: attackConfig, Passed: false, Report: "Web Error Occurred."}
    return nil
  }

  passed, reason, expected, actual := checkHTTPResponses(responses, attackConfig)

  if !passed {
    responseChannel <- Response{Expected: expected, Actual: actual, AttackConfig: attackConfig, Passed: false, Report: reason}
    return nil
  }

  responseChannel <- Response{AttackConfig: attackConfig, Passed: true, Report: fmt.Sprintf("URL Query Spam passed for endpoint %s", endpointConfig.Name)}
  return nil
}
