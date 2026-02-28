package main

import (

	"fmt"
	"log"
	"net/http"
	"time"


	//"test-task/pkg/models"

	"github.com/IBM/sarama"
	"github.com/gorilla/mux"
)

type TestProducer struct {
	producer sarama.SyncProducer
}

func MakeTestProducer() (*TestProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true          // для SyncProducer обязательно
	config.Producer.RequiredAcks = sarama.WaitForAll // самый надёжный вариант
	config.Producer.Retry.Max = 5
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	producer, err := sarama.NewSyncProducer([]string{"kafka:9092"}, config)
	if err != nil {
		return nil, err
	}

	return &TestProducer{producer: producer}, nil
}

/* func (t *TestProducer) ProduceRandomOrder() (*models.Order, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	order := models.createRandomOrder(rng)

	orderJSON, err := json.Marshal(order)
	if err != nil {
		log.Printf("failed to marshal order: %v", err)
		return nil, err
	}

	msg := &sarama.ProducerMessage{
		Topic: "orders",
		Value: sarama.ByteEncoder(orderJSON),
	}

	partition, offset, err := t.producer.SendMessage(msg)
	if err != nil {
		log.Printf("failed to send message: %v", err)
		return nil, err
	}

	log.Printf("Produced order %s → partition %d, offset %d", order.OrderUID, partition, offset)
	return order, nil
} */

func (t *TestProducer) Close() error {
	return t.producer.Close()
}

func ensureTopic() error {
	config := sarama.NewConfig()
	config.Admin.Timeout = 10 * time.Second

	admin, err := sarama.NewClusterAdmin([]string{"kafka:9092"}, config)
	if err != nil {
		return err
	}
	defer admin.Close()

	topics, err := admin.ListTopics()
	if err != nil {
		return err
	}

	if _, exists := topics["orders"]; !exists {
		topicDetail := &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
		err = admin.CreateTopic("orders", topicDetail, false)
		if err != nil {
			log.Printf("Failed to create topic 'orders': %v", err)
			return err
		}
		log.Println("Topic 'orders' created")
	} else {
		log.Println("Topic 'orders' already exists")
	}

	return nil
}

func main() {
	// Пытаемся создать топик с несколькими попытками
	for i := 0; i < 5; i++ {
		err := ensureTopic()
		if err == nil {
			break
		}
		log.Printf("Attempt %d failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	producer, err := MakeTestProducer()
	if err != nil {
		log.Fatalf("Cannot create producer: %v", err)
	}
	defer producer.Close()

	fmt.Println("Producer is up")

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "producer is fine")
	})
/* 	r.HandleFunc("/produce", func(w http.ResponseWriter, r *http.Request) {
		order, err := producer.ProduceRandomOrder()
		if err != nil {
			fmt.Fprintf(w, "Failed to produce: %v", err)
		} else {
			fmt.Fprintf(w, "Successfully produced order with id: %v", order.OrderUID)
		}
	}) */
	
	log.Println("Starting HTTP server on :8082")
	if err := http.ListenAndServe(":8082", r); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}
