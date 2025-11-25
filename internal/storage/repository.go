package storage

import (
	"context"
	"fmt"
	"log"

	"test-task/internal/cache"
	"test-task/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool  *pgxpool.Pool
	cache cache.Cache
}

const cacheCapacity = 10

func (repository *Repository) InitRepository(connStr string) error {

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Printf("Unable to parse config: %v", err)
		return err
	}

	log.Printf("InitRepository")

	repository.pool, err = pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Printf("Unable to connect to database: %v", err)
		return err
	}

	repository.cache = *cache.CreateCache(cacheCapacity)

	orders, err := repository.GetOrders(cacheCapacity)
	if err != nil {
		log.Printf("Unable to init cache: %v", err)
		return err
	}
	for i := 0; i < len(orders); i++ {
		repository.cache.Add(&orders[i])
	}

	return nil

}

func (repository *Repository) GetOrders(quantity int) ([]models.Order, error) {
	ctx := context.Background()

	conn, err := repository.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT order_uid FROM orders LIMIT $1", quantity)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close()

	uidsSet := make(map[string]struct{})
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, fmt.Errorf("scan uid: %w", err)
		}
		uidsSet[uid] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration: %w", err)
	}

	var orders []models.Order
	for uid := range uidsSet {
		order, found, err := repository.FindOrderById(uid)
		if err != nil {
			log.Printf("error finding %s: %v", uid, err)
			continue
		}
		if !found {
			log.Printf("order not found: %s", uid)
			continue
		}
		orders = append(orders, order)
	}

	return orders, nil
}

func (repository *Repository) InsertToDB(order *models.Order) error {
	conn, err := repository.pool.Acquire(context.Background())
	if err != nil {
		log.Printf("Unable to get connection from the Pool: %v", err)
		return err
	}
	defer conn.Release()

	tx, err := conn.Begin(context.Background())
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		return err
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), insertOrder,
		order.OrderUID, order.TrackNumber, order.Entry,
		order.Locale, order.InternalSignature, order.CustomerID,
		order.DeliveryService, order.Shardkey, order.SmID,
		order.DateCreated, order.OofShard)
	if err != nil {
		log.Printf("Error inserting order: %v", err)
		return err
	}

	delivery := &order.Delivery
	_, err = tx.Exec(context.Background(), insertDelivery,
		order.OrderUID, delivery.Name, delivery.Phone,
		delivery.Zip, delivery.City, delivery.Address,
		delivery.Region, delivery.Email)
	if err != nil {
		log.Printf("Error inserting delivery: %v", err)
		return err
	}

	payment := &order.Payment
	_, err = tx.Exec(context.Background(), insertPayment,
		order.OrderUID, payment.Transaction, payment.RequestID,
		payment.Currency, payment.Provider, payment.Amount,
		payment.PaymentDt, payment.Bank, payment.DeliveryCost,
		payment.GoodsTotal, payment.CustomFee)
	if err != nil {
		log.Printf("Error inserting payment: %v", err)
		return err
	}

	for i := 0; i < len(order.Items); i++ {
		item := &order.Items[i]
		_, err = tx.Exec(context.Background(), insertItem,
			order.OrderUID, item.ChrtID, item.TrackNumber,
			item.Price, item.Rid, item.Name, item.Sale,
			item.Size, item.TotalPrice, item.NmID,
			item.Brand, item.Status,
		)
		if err != nil {
			log.Printf("Error inserting items: %v", err)
			return err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		log.Fatalf("Error committing transaction: %v", err)
		return err
	}

	log.Println("Insert is completed")
	return nil

}

func (repository *Repository) FindOrderById(orderUid string) (order models.Order, exist bool, err error) {
	cacheOrder, exist, err := repository.cache.Get(orderUid)
	if exist {
		log.Printf("Have found in the cache")
		return *cacheOrder, true, nil
	}
	log.Printf("Have found in the DB")
	return repository.selectFromDB(orderUid)
}

func (repository *Repository) selectFromDB(orderUid string) (order models.Order, exist bool, err error) {
	exist = true

	conn, err := repository.pool.Acquire(context.Background())
	if err != nil {
		log.Printf("Unable to get connection from the Pool: %v", err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(context.Background(), pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	err = tx.QueryRow(context.Background(), "SELECT * FROM orders WHERE order_uid = $1", orderUid).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry,
		&order.Locale, &order.InternalSignature, &order.CustomerID,
		&order.DeliveryService, &order.Shardkey, &order.SmID,
		&order.DateCreated, &order.OofShard,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			exist = false
			err = nil
			log.Printf("Order does not exist = %v\n", orderUid)
			return
		}
		log.Printf("Error of query: %v", err)
		return
	}

	err = tx.QueryRow(context.Background(), "SELECT * FROM deliveries WHERE order_uid = $1", orderUid).Scan(
		&order.Delivery.OrderUID, &order.Delivery.Name, &order.Delivery.Phone,
		&order.Delivery.Zip, &order.Delivery.City, &order.Delivery.Address,
		&order.Delivery.Region, &order.Delivery.Email,
	)
	if err != nil && err != pgx.ErrNoRows {
		log.Printf("Query of delivery is failed : %v", err)
		return
	}

	err = tx.QueryRow(context.Background(), "SELECT * FROM payments WHERE order_uid = $1", orderUid).Scan(
		&order.Payment.OrderUID, &order.Payment.Transaction, &order.Payment.RequestID,
		&order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount,
		&order.Payment.PaymentDt, &order.Payment.Bank, &order.Payment.DeliveryCost,
		&order.Payment.GoodsTotal, &order.Payment.CustomFee,
	)
	if err != nil && err != pgx.ErrNoRows {
		log.Printf("Query of payment is failed : %v", err)
		return
	}

	rows, err := tx.Query(context.Background(), "SELECT * FROM items WHERE order_uid = $1", orderUid)
	if err != nil {
		log.Printf("Query of items is failed: %v", err)
		return
	}
	defer rows.Close()
	items, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Item])
	if err != nil {
		log.Printf("Collecting items is failed: %v", err)
		return
	}
	order.Items = items

	return
}

func (repository *Repository) Close() {
	repository.pool.Close()
}
