package main

import (
	"log"
	"net/http"

	"github.com/GSA-TTS/jemison/internal/common"
	"github.com/GSA-TTS/jemison/internal/env"
)

var ThisServiceName = "pack"
var ChFinalize = make(chan string)

func main() {
	env.InitGlobalEnv(ThisServiceName)

	InitializeQueues()
	engine := common.InitializeAPI()

	log.Println("environment initialized")

	go FinalizeTimer(ChFinalize)

	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, engine)
}
