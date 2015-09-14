package main

import ( 
        "github.com/julienschmidt/httprouter"
        "github.com/unrolled/render"
        "gopkg.in/mgo.v2"
        "github.com/nicolasbarbe/kafka"
        "encoding/json" 
        "net/http"
        "strconv"
        "time"
        "fmt"
        "log"
)


/** Types **/

// Discussion representats a started discussion
type Discussion struct {
  Id            string             `json:"id"            bson:"_id"`
  Title         string             `json:"title"         bson:"title"`
  Description   string             `json:"description"   bson:"description"`
  Initiator     string             `json:"initiator"     bson:"initiator"`
  CreatedAt     time.Time          `json:"createdAt"     bson:"createdAt"`
}


type NormalizedUser struct {
  Id            string             `json:"id"            bson:"_id"`
  FirstName     string             `json:"firstName"     bson:"firstName"`
  LastName      string             `json:"lastName"      bson:"lastName"`
  MemberSince   time.Time          `json:"memberSince"   bson:"memberSince"`
}


// Controller embeds the logic of the microservice
type Controller struct {
  mongo         *mgo.Database
  producer      *kafka.Producer
  renderer      *render.Render
}

// func (this *Controller) ListDiscussions(response http.ResponseWriter, request *http.Request, params httprouter.Params ) {
//   var discussions []Discussion
//   if err := this.mongo.C(discussionsCollection).Find(nil).Limit(100).All(&discussions) ; err != nil {
//     log.Print(err)
//   }
//   this.renderer.JSON(response, http.StatusOK, discussions)
// }

// func (this *Controller) ShowDiscussion(response http.ResponseWriter, request *http.Request, params httprouter.Params ) {

//   id, err := strconv.Atoi(params.ByName("id"))
//   if err != nil {
//     this.renderer.JSON(response, http.StatusNotFound, "Identifier format is not recognized")
//     return
//   }

//   var discussion Discussion
//   if err := this.mongo.C(discussionsCollection).FindId(id).One(&discussion) ; err != nil {
//     this.renderer.JSON(response, http.StatusNotFound, "Discussion not found")
//     return
//   }
   
//   this.renderer.JSON(response, http.StatusOK, discussion)
// }

func (this *Controller) StartDiscussion(response http.ResponseWriter, request *http.Request, params httprouter.Params ) {

  // build discusion from the request
  discussion := new(Discussion)
  err := json.NewDecoder(request.Body).Decode(discussion)
  if err != nil {
    this.renderer.JSON(response, 422, "Request body is not a valid JSON")
    return
  }
  
  // check that initiator exists
  count, err := this.mongo.C(usersCollection).FindId(discussion.Initiator).Limit(1).Count()
  if err != nil {
      this.renderer.JSON(response, 500, err)
      return
  }
  if count <= 0 {
    this.renderer.JSON(response, 422, "Initiator of the discussion does not exist")
    return
  }  
  
  // persits discussion
  if err := this.mongo.C(discussionsCollection).Insert(discussion) ; err != nil {
    log.Printf("Cannot create document in collection %s : %s", discussionsCollection, err)
    this.renderer.JSON(response, 500, "Cannot save discussion")
    return
  }

  // create message
  body, err := json.Marshal(discussion)
  if err != nil {
    log.Printf("Cannot marshall document : %s", err)
    this.renderer.JSON(response, 500, "Cannot process the request")
    return
  }

  message := append([]byte(fmt.Sprintf("%02d%v",len(discussionStarted), discussionStarted)), body ...)

  // send message
  err = this.producer.SendMessageToTopic(message, discussionsTopic)
  if err != nil {
    this.renderer.JSON(response, http.StatusInternalServerError, "Failed to send the message: " + err.Error())
    return
  } 

  // render the response
  this.renderer.JSON(response, 200, "ok")
}


func (this *Controller) ConsumeUsers(message []byte) {
  idx, _    := strconv.Atoi(string(message[:2]))
  eventType := string(message[2:idx+2])
  body      := message[idx+2:]

  if eventType != userCreated {
    log.Printf("Message with type %v is ignored. Type %v was expected", eventType, userCreated)
    return
  }

  // unmarshal user from event body
  var normalizedUser NormalizedUser
  if err := json.Unmarshal(body, &normalizedUser); err != nil {
    log.Print("Cannot unmarshal user")
    return
  }

  // save user Id
  if err := this.mongo.C(usersCollection).Insert(normalizedUser) ; err != nil {
    log.Printf("Cannot save document in collection %s : %s", usersCollection, err)
    return
  }
}



