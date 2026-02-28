package main

import (
	"fmt"
	"log"
	"net/http"

	"test-task/internal/app"
	"test-task/internal/config"

	"github.com/gorilla/mux"
)

func main() {

/* 	//ctx, cancel := context.WithCancel(context.Background())
	defer cancel()*/

/* 	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
  */
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to get config: %v", err)
	}

	/* 	connStr := "postgres://postgres:qwerty@localhost:5432/WB_ordersDB"
	   	newApp, err := app.NewApp(connStr) */
	newApp, err := app.NewApp(fmt.Sprintf("postgres://%s:%s@db:5432/%s",
		config.DBuser,
		config.DBpassword,
		config.DBname,
	))
	if err != nil {
		log.Fatalf("Failed to initialize")
	}
	defer newApp.Close()

	r := mux.NewRouter()

	r.HandleFunc("/", newApp.HomeHandler)
	r.HandleFunc("/order/{order_uid}", newApp.GetOrderById).Methods("GET")
	//r.HandleFunc("/add", newApp.CreateOrders).Methods("GET")

	log.Println("The service has shut down.")
http.ListenAndServe(":8080", r)
}
