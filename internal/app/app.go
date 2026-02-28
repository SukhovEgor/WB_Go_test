package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"


	"test-task/pkg/models"
	"test-task/internal/storage"

	"github.com/IBM/sarama"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/gorilla/mux"
)

type App struct {
	repository storage.Repository
	consumer   sarama.ConsumerGroup
	stopChan   chan struct{}
}

func NewApp(connStr string) (*App, error) {
	app := &App{}
	err := app.repository.InitRepository(connStr)
	if err != nil {
		log.Printf("Unable to connect to database: %v", err)
		return nil, err
	}

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = false
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(
		[]string{"kafka:9092"},
		"orders-consumer-group",
		config,
	)
	if err != nil {
		return nil, err
	}

	app.consumer = consumerGroup

	go app.runConsumer()

	return app, nil
}

func (a *App) runConsumer() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        <-a.stopChan
        cancel()
    }()

    // Самый простой потребитель без группы (не использует ConsumerGroup)
    consumer, err := sarama.NewConsumer([]string{"kafka:9092"}, nil)
    if err != nil {
        log.Printf("failed to create consumer: %v", err)
        return
    }
    defer consumer.Close()

    partitionConsumer, err := consumer.ConsumePartition("orders", 0, sarama.OffsetNewest)
    if err != nil {
        log.Printf("failed to consume partition: %v", err)
        return
    }
    defer partitionConsumer.Close()

    log.Println("simple kafka consumer started (partition 0)")

    for {
        select {
        case msg, ok := <-partitionConsumer.Messages():
            if !ok {
                log.Println("messages channel closed")
                return
            }

            var order models.Order
            if err := json.Unmarshal(msg.Value, &order); err != nil {
                log.Printf("unmarshal error: %v", err)
                continue
            }

            if err := a.repository.InsertToDB(&order); err != nil {
                log.Printf("store error: %v", err)
                // можно добавить retry или логирование
            }

            log.Printf("processed order %s from offset %d", order.OrderUID, msg.Offset)

        case <-ctx.Done():
            log.Println("consumer stopped by context")
            return

        case err := <-partitionConsumer.Errors():
            log.Printf("consumer error: %v", err)
        }
    }
}

func (a *App) HomeHandler(w http.ResponseWriter, r *http.Request) {
	html, err := os.ReadFile("frontend/index.html")
	if err != nil {
		log.Printf("Error reading index.html: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "internal error")
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(html)
}

func (a *App) GetOrderById(w http.ResponseWriter, r *http.Request) {
	orderUid := mux.Vars(r)["order_uid"]
	log.Printf("Searching : %v", orderUid)

	order, exist, err := a.repository.FindOrderById(orderUid)

	if !exist {
		fmt.Fprintf(w, "Order %v does not exist\n", orderUid)
		return
	} else if err != nil {
		log.Printf("Finding order by id is failed: %v", err)
		return
	}

	json_data, err := json.MarshalIndent(order, "", "\t")
	if err != nil {
		log.Printf("Failed to create json: %v", err)
	}
	fmt.Fprintf(w, "%s\n", json_data)
}

/* func (a *App) HandleGetOrderByID(uid string) (interface{}, error) {
	uid = strings.Trim(uid, `"`)
	log.Printf("HandleSearching : %v", uid)
	order, exist, err := a.repository.FindOrderById(uid)
	if err != nil {
		log.Printf("DB fetch error: %v", err)
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("Order %s is not found", uid)
	}
	return order, nil
}

func (a *App) HandleCreateOrders(data string) (interface{}, error) {
	orderCount, err := strconv.Atoi(data)
	if err != nil {
		log.Printf("Parse error: %v", err)
		return nil, err
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	ordersAdded := 0

	var orders []models.Order
	for i := 0; i < orderCount; i++ {
		order, err := createRandomOrder(rng)

		if err != nil {
			return nil, err
		}

		if err := a.repository.InsertToDB(&order); err != nil {
			log.Printf("DB inserting error: %v", err)
			return nil, err
		}
		ordersAdded++
		orders = append(orders, order)
	}

	return orders, nil
} */

/* func (a *App) CreateOrders(w http.ResponseWriter, r *http.Request) {
	orderCount := 2

	var orders []models.Order
	if err := json.Unmarshal([]byte(msg), &orders); err != nil {
		response := map[string]interface{}{
			"error":   true,
			"message": msg,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		log.Printf("Error while creating response: %v", err)
	}
} */ 

func (a *App) CreateOrders(w http.ResponseWriter, r *http.Request) {

	amount := 2
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var orders []models.Order
	for i := 0; i < amount; i++ {
		order, err := createRandomOrder(rng)
		if err != nil {
			log.Printf("Failed to generate order #%d: %v", i+1, err)
			continue
		}

		if err := a.repository.InsertToDB(&order); err != nil {
			log.Printf("Failed to insert order #%d: %v", i+1, err)
			continue
		}

		orders = append(orders, order)
	}

	json_data, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		log.Printf("Error making json: %v", err)
	}

	fmt.Fprintf(w, "%s\n", json_data)
}

func createRandomOrder(rng *rand.Rand) (models.Order, error) {
	var order models.Order
	var delivery models.Delivery
	var payment models.Payment

	if err := gofakeit.Struct(&order); err != nil {
		return order, err
	}
	if err := gofakeit.Struct(&delivery); err != nil {
		return order, err
	}
	if err := gofakeit.Struct(&payment); err != nil {
		return order, err
	}

	itemCount := rng.Intn(10) + 1
	items := make([]models.Item, 0, itemCount)
	var goodsTotal float64

	for i := 0; i < itemCount; i++ {
		var item models.Item
		if err := gofakeit.Struct(&item); err != nil {
			return order, err
		}

		quantity := rng.Intn(5) + 1

		item.Sale = rng.Intn(51)

		item.TotalPrice = float64(item.Price*quantity) * (1 - float64(item.Sale)/100.0)

		goodsTotal += item.TotalPrice
		items = append(items, item)
	}

	payment.GoodsTotal = goodsTotal
	payment.Amount = payment.DeliveryCost + goodsTotal + payment.CustomFee

	order.Delivery = delivery
	order.Payment = payment
	order.Items = items

	return order, nil
}

func (a *App) Close() {
	a.repository.Close()
	log.Println("Stopping consumer...")
	close(a.stopChan)

	time.Sleep(3 * time.Second)

	if err := a.consumer.Close(); err != nil {
		log.Printf("error closing consumer: %v", err)
	}
}
