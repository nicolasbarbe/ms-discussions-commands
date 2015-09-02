package main

import (
  "github.com/albrow/negroni-json-recovery"
  "github.com/codegangsta/negroni"
  "github.com/julienschmidt/httprouter"
)


func main() {

  router := httprouter.New()

  router.GET("/api/v1/discussions", Index)
  router.GET("/api/v1/discussions/:id", Show)

  router.POST("/api/v1/commands/discussions/start", CreateStartDiscussionCommand)

 
  app := negroni.New(negroni.NewLogger())
  app.Use(recovery.JSONRecovery(true))
  app.UseHandler(router)

  app.Run(":8080")
}