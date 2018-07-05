package main

import (
  "net/http"
  "encoding/json"
  "fmt"
  httptransport "github.com/go-openapi/runtime/client"
  apiclient "github.com/mluvii/publicapi-go/client"
  "github.com/go-openapi/strfmt"
  "github.com/mluvii/publicapi-go/client/webhooks"
  "github.com/mluvii/publicapi-go/models"
  "net/url"
  "io/ioutil"
  "errors"
  "os"
)

type EventPayload struct {
  EventType string
  Data json.RawMessage
}

type EventDataSessionStarted struct {
  Id int64
  Channel string
  Source string
  Started string
}

type EventDataSessionEnded struct {
  Id int64
  Channel string
  Source string
  Started string
  Ended string
}

func obtainToken(authKey string) (string, error) {
  resp, err := http.PostForm("https://applocal.mluvii.com/MasSrv/Service/Login",
    url.Values{"authKey": {authKey}})

  if err != nil {
    return "", err
  }

  defer resp.Body.Close()

  if resp.StatusCode == http.StatusOK {
    bodyBytes, _ := ioutil.ReadAll(resp.Body)
    return string(bodyBytes), nil
  }

  return "", errors.New("auth failed")
}

func postOrPutWebhook(client *apiclient.Mluviiapi, model models.PublicAPIWebhookModelsWebhookModel) (bool, error) {
  postParams := webhooks.NewAPIV1WebhooksPostParams()
  postParams.Model = &model

  _, err := client.Webhooks.APIV1WebhooksPost(postParams)
  if _, conflict := err.(*webhooks.APIV1WebhooksPostConflict); conflict {
    putParams := webhooks.NewAPIV1WebhooksByIDPutParams()
    putParams.ID = err.(*webhooks.APIV1WebhooksPostConflict).Payload
    putParams.Model = postParams.Model
    _, err := client.Webhooks.APIV1WebhooksByIDPut(putParams)
    return err == nil, err
  }

  return err == nil, err
}

func subscribeToEvents(client *apiclient.Mluviiapi) {
  callbackUrl := "http://go:isawesome@localhost:5000/mluviiwebhook"
  model := models.PublicAPIWebhookModelsWebhookModel{
    CallbackURL: &callbackUrl,
    EventTypes:  []string{"SessionStarted", "SessionForwarded", "SessionEnded", "UserStatusChanged"},
  }

  if ok, err := postOrPutWebhook(client, model); !ok {
    panic(err)
  }
}

func decodeMluviiEvent(decoder *json.Decoder) (string, interface{}, error) {
  var ep EventPayload
  err := decoder.Decode(&ep)
  if err != nil {
    return "", nil, err
  }

  switch ep.EventType {
  case "SessionStarted":
    var data EventDataSessionStarted
    err := json.Unmarshal(ep.Data, &data)
    return ep.EventType, data, err
  case "SessionEnded":
    var data EventDataSessionEnded
    err := json.Unmarshal(ep.Data, &data)
    return ep.EventType, data, err
  default:
    var data interface{}
    err := json.Unmarshal(ep.Data, &data)
    return ep.EventType, data, err
  }
}

func processMluviiEvent(w http.ResponseWriter, r *http.Request) {
  user, pass, _ := r.BasicAuth()
  if user != "go" || pass != "isawesome" {
    http.Error(w, "Unauthorized.", 401)
    return
  }

  decoder := json.NewDecoder(r.Body)
  eventType, eventData, err := decodeMluviiEvent(decoder)
  if err != nil {
    panic(err)
  }

  fmt.Println(eventType)
  switch ed := eventData.(type) {
  case EventDataSessionEnded:
    fmt.Printf("session %v: %+v\n", ed.Id, ed)
  default:
    fmt.Printf("%+v\n", ed)
  }
}

func main() {
  token, err := obtainToken(os.Args[1])
  if err != nil {
    panic(err)
  }

  println(token)

  transport := httptransport.New("127.0.0.1:44301", "", nil)
  client := apiclient.New(transport, strfmt.Default)
  bearerTokenAuth := httptransport.BearerToken(token)
  transport.DefaultAuthentication = bearerTokenAuth

  subscribeToEvents(client)

  http.HandleFunc("/mluviiwebhook", processMluviiEvent)
  if err := http.ListenAndServe(":5000", nil); err != nil {
    panic(err)
  }
}
