package main

import (
	"log"
	"os"
	"test-vehcile-monitoring/common/logger"
)

func main() {
	log.Println("main start")
	defer log.Println("main closed")

	Init()

	makeConnectionDB()

	// Todo  init jaeger(tracer)

	// Log level setting
	log, err := logger.New("test", 1, os.Stdout)
	if err != nil {
		panic(err) // Check for error
	}

	// Critically log critical
	log.Critical("This is Critical!")
	log.CriticalF("%+v", err)
}

func Init() {

}

func makeConnectionDB() {

}
