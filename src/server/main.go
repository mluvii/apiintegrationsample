package main

import (
  "net/http"
  "encoding/json"
  "fmt"
  httptransport "github.com/go-openapi/runtime/client"
  apiclient "github.com/mluvii/publicapi-go/client"
)

type EventPayload struct {
  EventType string
  Data map[string]interface{}
}

func processMluviiEvent(w http.ResponseWriter, r *http.Request) {
  decoder := json.NewDecoder(r.Body)
  var ep EventPayload
  err := decoder.Decode(&ep)
  if err != nil {
    panic(err)
  }
  fmt.Println(ep.EventType)
  fmt.Println(ep.Data)
}

func main() {
  http.HandleFunc("/mluviiwebhook", processMluviiEvent)
  if err := http.ListenAndServe(":5000", nil); err != nil {
    panic(err)
  }
}
