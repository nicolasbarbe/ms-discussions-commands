package main

import ( 
        "github.com/julienschmidt/httprouter"
        "github.com/unrolled/render"
        "github.com/nicolasbarbe/kafka"
        "encoding/json" 
        "net/http"
        "strconv"
        "strings"
        "time"
        "sort"
        "fmt"
        "os"
)

type Discussion struct {
  Id            uint32     `json:"id"`
  Title         string     `json:"title"`
  Description   string     `json:"description"`
  Initiator     string     `json:"initiator"`
  CreatedAt     time.Time  `json:"createdAt"`
}

type DiscussionStarted struct {
  Id         string    `json:"id"`
}

var producer = kafka.NewProducer(strings.Split(os.Getenv("KAFKA_BROKERS"), ","))
var renderer = render.New(render.Options{
    IndentJSON: true,
})


/** Sample data */
var discussions = make([]Discussion, 0)
var initiators  =[]string{"nbarbe"}

func Index(response http.ResponseWriter, request *http.Request, params httprouter.Params ) {
  renderer.JSON(response, http.StatusOK, discussions)
}

func Show(response http.ResponseWriter, request *http.Request, params httprouter.Params ) {

  idx, err := strconv.Atoi(params.ByName("id"))
  if err != nil {
    renderer.JSON(response, http.StatusNotFound, "Identifier format is not recognized")
    return
  }

  if idx >= len(discussions) {
    renderer.JSON(response, http.StatusNotFound, "Discussion not found")
    return
  }
   
  renderer.JSON(response, http.StatusOK,  discussions[idx])
}


func CreateStartDiscussionCommand(response http.ResponseWriter, request *http.Request, params httprouter.Params ) {

  // build discusion from the request
  discussion := new(Discussion)
  err := json.NewDecoder(request.Body).Decode(discussion)
  if err != nil {
    renderer.JSON(response, 422, "Request body is not a valid JSON")
    return
  }
  discussion.Id = uint32(len(discussions))

  // check that initiator exists
  found, _ := find(discussion.Initiator, initiators)

  if !found {
    renderer.JSON(response, 422, "Initiator of the discussion does not exist")
    return
  }

  // persits discussion
  discussions = append(discussions, *discussion)

  // create event
  event := DiscussionStarted{ Id: fmt.Sprint(discussion.Id) }

  // send event
  err = producer.SendEventToTopic(event, "discussions")
  if err != nil {
    renderer.JSON(response, http.StatusInternalServerError, "Failed to send the message: " + err.Error())
    return
  } 

  // render the response
  renderer.JSON(response, 200, "ok")
}

func addInitiator(initiator string) {
  found, idx := find(initiator, initiators)
  if found {
   return
  }

  initiators = append(initiators, "")
  copy(initiators[idx+1:], initiators[idx:])
  initiators[idx] = initiator
}

func find( value string, slice []string ) (bool, int) {
  idx := sort.SearchStrings(slice, value)
  return idx < len(slice) && slice[idx] == value, idx 
}








