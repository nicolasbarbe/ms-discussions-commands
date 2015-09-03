package main

import (
  "github.com/julienschmidt/httprouter"
  "github.com/unrolled/render"
  "github.com/nicolasbarbe/kafka"
  "github.com/nicolasbarbe/mongo"
  "net/http"
  "strings"
  "log"
  "os"
)

/** Constants **/

const discussionsCollection = "discussions"
const usersCollection       = "users"
const discussionsTopic      = "discussions"


/** Main **/

func main() {

  // create kafka producer
  var producer = kafka.NewProducer(strings.Split(os.Getenv("KAFKA_BROKERS"), ","))

  // create http renderer
  var renderer = render.New(render.Options{
      IndentJSON: true,
  })

  // create mongodb session
  mongo := mongo.NewMongo(os.Getenv("MONGODB_CS"), os.Getenv("MONGODB_DB"))

  // create controller
  controller := Controller {
    mongo    : mongo,
    producer : producer,
    renderer : renderer,
  }

  // create routes
  router := httprouter.New()
  router.POST("/api/v1/commands/discussions/start", controller.CreateStartDiscussionCommand)
  
  log.Fatal(http.ListenAndServe(":8080", router))

  defer func() {
    producer.Close()
    mongo.Close()
  }()
}