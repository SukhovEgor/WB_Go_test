package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"test-task/internal/app"
	"test-task/internal/config"
	"test-task/internal/kafka"

	"github.com/gorilla/mux"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to get config: %v", err)
	}

	/* 	connStr := "postgres://postgres:qwerty@localhost:5433/WB_ordersDB"
	   	newApp, err := app.NewApp(connStr) */
	newApp, err := app.NewApp(fmt.Sprintf("postgres://%s:%s@localhost:5433/%s",
		config.DBUser,
		config.DBPassword,
		config.DBName,
	))
	if err != nil {
		log.Fatalf("Failed to initialize")
	}
	defer newApp.Close()

	brokers := []string{"localhost:9092"}

	consumer, err := kafka.ConnectConsumer(brokers)
	if err != nil {
		log.Fatalf("Kafka consumer init error: %v", err)
	}
	defer consumer.Close()

	producer, err := kafka.ConnectProducer(brokers)
	if err != nil {
		log.Fatalf("Kafka producer init error: %v", err)
	}
	defer producer.Close()

	go kafka.DoServiceRequest(producer, consumer, ctx.Done(),
		newApp.HandleCreateOrders, "post_order", "post_order_response")

	go kafka.DoServiceRequest(producer, consumer, ctx.Done(),
		newApp.HandleGetOrderByID, "get_order_by_id", "get_order_by_id_response")

	r := mux.NewRouter()

	r.HandleFunc("/", newApp.HomeHandler)
	r.HandleFunc("/order/{order_uid}", newApp.GetOrderById).Methods("GET")
	r.HandleFunc("/add", newApp.CreateOrders).Methods("GET")

	go func() {
		log.Println("HTTP server started at :3000")
		if err := http.ListenAndServe(":3000", r); err != nil {
			log.Printf("http server error: %v", err)
			cancel()
		}
	}()

	waitForShutdown(sigchan, cancel)
	log.Println("The service has shut down.")

}

func waitForShutdown(sigchan <-chan os.Signal, cancel context.CancelFunc) {
	<-sigchan
	log.Println("A termination signal is received, and the service stops...")
	cancel()
}
