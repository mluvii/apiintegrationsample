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
)

type EventPayload struct {
  EventType string
  Data map[string]interface{}
}

func postOrPutWebhook(client *apiclient.Mluviiapi, model models.PublicAPIWebhookModelsWebhookModel) (bool, error) {
  postParams := webhooks.NewAPIV1WebhooksPostParams();
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
  callbackUrl := "http://localhost:5000/mluviiwebhook"
  model := models.PublicAPIWebhookModelsWebhookModel{
    &callbackUrl,
    []string{"SessionStarted", "SessionForwarded", "SessionEnded", "UserStatusChanged"},
  }

  if ok, err := postOrPutWebhook(client, model); !ok {
    panic(err)
  }
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
  transport := httptransport.New("127.0.0.1:44301", "", nil)
  client := apiclient.New(transport, strfmt.Default)

  subscribeToEvents(client)

  http.HandleFunc("/mluviiwebhook", processMluviiEvent)
  if err := http.ListenAndServe(":5000", nil); err != nil {
    panic(err)
  }
}
