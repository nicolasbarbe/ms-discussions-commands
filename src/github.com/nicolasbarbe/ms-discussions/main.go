package main

import (
  "github.com/julienschmidt/httprouter"
  "net/http"
  "log"
)


func main() {
  router := httprouter.New()
  router.POST("/api/v1/commands/discussions/start", CreateStartDiscussionCommand)
  log.Fatal(http.ListenAndServe(":8080", router))
}