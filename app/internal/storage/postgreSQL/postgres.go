// Database layout
package postgreSQL

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v4"
	"orderInfo/app/internal/models"
	"orderInfo/app/internal/storage"
	"time"
)

type DataBase struct {
	DB *pgx.Conn
}

type Storage interface {
	AddOrder(ctx context.Context, order models.Order) error
	ShowData(ctx context.Context, uid string) (models.Order, error)
	CashReturner(ctx context.Context) ([]models.Order, error)
}

func OpenDB(ctx context.Context, url string) (*DataBase, error) {
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(ctx); err != nil {
		return nil, err
	}

	return &DataBase{DB: conn}, err
}

func (db *DataBase) AddOrder(ctx context.Context, order models.Order) error {
	const op = "storage.postgres.AddOrder"

	// add order in transaction
	tx, err := db.DB.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	row := db.DB.QueryRow(ctx, `INSERT INTO "Delivery"(name, phone, zip, city, address, region, email) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip, order.Delivery.City, order.Delivery.Address,
		order.Delivery.Region, order.Delivery.Email)

	var deliveryId int
	if err := row.Scan(&deliveryId); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	row = db.DB.QueryRow(ctx, `INSERT INTO "Payment"(transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total, custom_fee) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
		order.Payment.Transaction, order.Payment.RequestId, order.Payment.Currency, order.Payment.Provider,
		order.Payment.Amount, time.Unix(order.Payment.PaymentDt, 0), order.Payment.Bank, order.Payment.DeliveryCost,
		order.Payment.GoodsTotal, order.Payment.CustomFee)

	var paymentId int
	if err := row.Scan(&paymentId); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = db.DB.Exec(ctx, `INSERT INTO "Order"(order_uid, track_number, entry, delivery_id, payment_id, locale, internal_signature, customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		order.Uid, order.TrackNumber, order.Entry, deliveryId, paymentId, order.Locale,
		order.InternalSignature, order.CustomerId, order.DeliveryService, order.Shardkey, order.SmId,
		order.DateCreated, order.OofShard)
	if err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	for i := range order.Items {
		_, err = db.DB.Exec(ctx, `INSERT INTO "Items"(order_id, chrt_id, track_number, price, rid, name, sale, size, total_price, nm_id, brand, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.Uid, order.Items[i].ChrtId, order.Items[i].TrackNumber, order.Items[i].Price, order.Items[i].Rid,
			order.Items[i].Name, order.Items[i].Sale, order.Items[i].Size, order.Items[i].TotalPrice, order.Items[i].NmId,
			order.Items[i].Brand, order.Items[i].Status)
		if err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (db *DataBase) ShowData(ctx context.Context, uid string) (models.Order, error) {
	const op = "storage.postgres.ShowData"
	var t int
	var dbTime time.Time

	var order models.Order
	if err := db.DB.QueryRow(ctx, fmt.Sprintf(`SELECT * FROM "Order" o JOIN "Payment" p ON o.payment_id = p.id JOIN "Delivery" d ON o.delivery_id = d.id WHERE order_uid = '%s'`, uid)).
		Scan(&order.Uid, &order.TrackNumber, &order.Entry, &t, &t, &order.Locale,
			&order.InternalSignature, &order.CustomerId, &order.DeliveryService, &order.Shardkey,
			&order.SmId, &order.DateCreated, &order.OofShard, &t, &order.Payment.Transaction, &order.Payment.RequestId,
			&order.Payment.Currency, &order.Payment.Provider, &order.Payment.Amount, &dbTime,
			&order.Payment.Bank, &order.Payment.DeliveryCost, &order.Payment.GoodsTotal, &order.Payment.CustomFee,
			&t, &order.Delivery.Name, &order.Delivery.Phone, &order.Delivery.Zip, &order.Delivery.City,
			&order.Delivery.Address, &order.Delivery.Region, &order.Delivery.Email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Order{}, fmt.Errorf("%s: %w", op, storage.ErrUidNotFound)
		}

		return models.Order{}, fmt.Errorf("%s: %w", op, err)
	}

	order.Payment.PaymentDt = dbTime.Unix()

	rows, _ := db.DB.Query(ctx, fmt.Sprintf(`SELECT * FROM "Items" WHERE id = '%s'`, order.Uid))

	i := 0
	for rows.Next() {
		rows.Scan(&t, &order.Uid, &order.Items[i].ChrtId, &order.Items[i].TrackNumber,
			&order.Items[i].Price, &order.Items[i].Rid, &order.Items[i].Name, &order.Items[i].Sale, &order.Items[i].Size,
			&order.Items[i].TotalPrice, &order.Items[i].NmId, &order.Items[i].Brand, &order.Items[i].Status)

		i++
	}

	return order, nil
}

func (db *DataBase) CashReturner(ctx context.Context) ([]models.Order, error) {
	const op = "storage.postgres.CashReturner"
	var orders []models.Order
	var t int
	var dbTime time.Time

	rows, err := db.DB.Query(ctx, `SELECT DISTINCT * FROM "Order" o JOIN "Payment" p ON o.payment_id = p.id JOIN "Delivery" d ON o.delivery_id = d.id`)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrNoRecords)
		}

		return nil, fmt.Errorf("%s: %w", op, err)
	}

	i := 0
	for rows.Next() {
		orders = append(orders, models.Order{})
		rows.Scan(&orders[i].Uid, &orders[i].TrackNumber, &orders[i].Entry, &t, &t, &orders[i].Locale,
			&orders[i].InternalSignature, &orders[i].CustomerId, &orders[i].DeliveryService, &orders[i].Shardkey,
			&orders[i].SmId, &orders[i].DateCreated, &orders[i].OofShard, &t, &orders[i].Payment.Transaction, &orders[i].Payment.RequestId,
			&orders[i].Payment.Currency, &orders[i].Payment.Provider, &orders[i].Payment.Amount, &dbTime,
			&orders[i].Payment.Bank, &orders[i].Payment.DeliveryCost, &orders[i].Payment.GoodsTotal, &orders[i].Payment.CustomFee,
			&t, &orders[i].Delivery.Name, &orders[i].Delivery.Phone, &orders[i].Delivery.Zip, &orders[i].Delivery.City,
			&orders[i].Delivery.Address, &orders[i].Delivery.Region, &orders[i].Delivery.Email)
		orders[i].Payment.PaymentDt = dbTime.Unix()

		itemRows, _ := db.DB.Query(ctx, fmt.Sprintf(`SELECT * FROM "Items" WHERE id = '%s'`, orders[i].Uid))

		j := 0
		for itemRows.Next() {
			rows.Scan(&t, &orders[i].Uid, &orders[i].Items[j].ChrtId, &orders[i].Items[j].TrackNumber,
				&orders[i].Items[j].Price, &orders[i].Items[j].Rid, &orders[i].Items[j].Name, &orders[i].Items[j].Sale, &orders[i].Items[j].Size,
				&orders[i].Items[j].TotalPrice, &orders[i].Items[j].NmId, &orders[i].Items[j].Brand, &orders[i].Items[j].Status)

			j++
		}

		i++
	}

	return orders, nil
}
