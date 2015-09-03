package main

import ( 
        "github.com/julienschmidt/httprouter"
        "github.com/unrolled/render"
        "github.com/nicolasbarbe/mongo"
        "github.com/nicolasbarbe/kafka"
        "encoding/json" 
        "net/http"
        "time"
        "fmt"
)


/** Types **/

// Discussion represents a discussion and is used both in the database and REST API
type Discussion struct {
  Id            uint32     `json:"id"            bson:"_id"`
  Title         string     `json:"title"         bson:"title"`
  Description   string     `json:"description"   bson:"description"`
  Initiator     string     `json:"initiator"     bson:"initiator"`
  CreatedAt     time.Time  `json:"createdAt"     bson:"createdAt"`
}

// DiscussionStarted is an event representing a started discussion
type DiscussionStarted struct {
  Id            string    `json:"id"`
}

// Controller embeds the logic of the microservice
type Controller struct {
  mongo         *mongo.Mongo
  producer      *kafka.Producer
  renderer      *render.Render
}

func (this *Controller) CreateStartDiscussionCommand(response http.ResponseWriter, request *http.Request, params httprouter.Params ) {

  // build discusion from the request
  discussion := new(Discussion)
  err := json.NewDecoder(request.Body).Decode(discussion)
  if err != nil {
    this.renderer.JSON(response, 422, "Request body is not a valid JSON")
    return
  }
  
  // check that initiator exists
  if(!this.mongo.Exists(usersCollection, discussion.Initiator)) {
    this.renderer.JSON(response, 422, "Initiator of the discussion does not exist")
    return
  }

  // persits discussion
  this.mongo.Create(discussionsCollection, discussion)

  // create and send event
  err = this.producer.SendEventToTopic(DiscussionStarted{ Id: fmt.Sprint(discussion.Id) }, discussionsTopic)
  if err != nil {
    this.renderer.JSON(response, http.StatusInternalServerError, "Failed to send the message: " + err.Error())
    return
  } 

  // render the response
  this.renderer.JSON(response, 200, "ok")
}






