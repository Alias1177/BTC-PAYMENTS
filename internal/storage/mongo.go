package storage

import (
	"chi/BTC-PAYMENTS/pkg/logger"
	"chi/BTC-PAYMENTS/pkg/models"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoRepository реализует интерфейс models.Repository для MongoDB
type MongoRepository struct {
	client         *mongo.Client
	db             *mongo.Database
	collection     *mongo.Collection
	transactionTTL time.Duration
	logger         *logger.Logger
}

// NewMongoRepository создает новый экземпляр MongoDB репозитория
func NewMongoRepository(uri, dbName, collectionName string, logger *logger.Logger) (*MongoRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Создание клиента MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к MongoDB: %w", err)
	}

	// Проверка соединения
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения к MongoDB: %w", err)
	}

	// Получение ссылки на базу данных и коллекцию
	db := client.Database(dbName)
	collection := db.Collection(collectionName)

	// Создание индексов для оптимизации запросов
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "invoice_id", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "order_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: 1}},
		},
	}

	_, err = collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания индексов: %w", err)
	}

	logger.Info("MongoDB репозиторий успешно инициализирован: %s/%s", dbName, collectionName)

	return &MongoRepository{
		client:         client,
		db:             db,
		collection:     collection,
		transactionTTL: 90 * 24 * time.Hour, // хранение транзакций в течение 90 дней
		logger:         logger,
	}, nil
}

// Close закрывает соединение с MongoDB
func (r *MongoRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Disconnect(ctx)
}

// CreateTransaction создает новую запись о транзакции
func (r *MongoRepository) CreateTransaction(tx *models.Transaction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Установка времени создания и обновления
	now := time.Now()
	tx.CreatedAt = now
	tx.UpdatedAt = now

	// Вставка документа
	result, err := r.collection.InsertOne(ctx, tx)
	if err != nil {
		return fmt.Errorf("ошибка создания транзакции: %w", err)
	}

	// Сохранение ID в структуре
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		tx.ID = oid
	}

	r.logger.Info("Создана новая транзакция: %s для заказа %s", tx.InvoiceID, tx.OrderID)
	return nil
}

// UpdateTransactionStatus обновляет статус транзакции по ID инвойса
func (r *MongoRepository) UpdateTransactionStatus(invoiceID, status string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"invoice_id": invoiceID}
	update := bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	}

	// Если статус "paid" или "complete", устанавливаем дату оплаты
	if status == "paid" || status == "complete" {
		update["$set"].(bson.M)["paid_at"] = time.Now()
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса транзакции: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("транзакция с invoice_id=%s не найдена", invoiceID)
	}

	r.logger.Info("Обновлен статус транзакции %s: %s", invoiceID, status)
	return nil
}

// GetTransactionByInvoiceID получает транзакцию по ID инвойса
func (r *MongoRepository) GetTransactionByInvoiceID(invoiceID string) (*models.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var transaction models.Transaction
	filter := bson.M{"invoice_id": invoiceID}

	err := r.collection.FindOne(ctx, filter).Decode(&transaction)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("транзакция с invoice_id=%s не найдена", invoiceID)
		}
		return nil, fmt.Errorf("ошибка получения транзакции: %w", err)
	}

	return &transaction, nil
}

// ListTransactions получает список транзакций с фильтрацией и пагинацией
func (r *MongoRepository) ListTransactions(filters map[string]interface{}, page, perPage int) ([]*models.Transaction, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Формирование фильтра запроса
	filter := bson.M{}
	if status, ok := filters["status"].(string); ok && status != "" {
		filter["status"] = status
	}

	// Фильтрация по датам
	if dateFrom, ok := filters["date_from"].(time.Time); ok {
		if _, exists := filter["created_at"]; !exists {
			filter["created_at"] = bson.M{}
		}
		filter["created_at"].(bson.M)["$gte"] = dateFrom
	}

	if dateTo, ok := filters["date_to"].(time.Time); ok {
		if _, exists := filter["created_at"]; !exists {
			filter["created_at"] = bson.M{}
		}
		filter["created_at"].(bson.M)["$lte"] = dateTo
	}

	// Настройка пагинации
	skip := (page - 1) * perPage
	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(perPage)).
		SetSort(bson.D{{Key: "created_at", Value: -1}}) // Сортировка по дате создания (новые вначале)

	// Выполнение запроса
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка получения списка транзакций: %w", err)
	}
	defer cursor.Close(ctx)

	// Декодирование результатов
	var transactions []*models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, 0, fmt.Errorf("ошибка декодирования транзакций: %w", err)
	}

	// Получение общего количества документов для пагинации
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("ошибка подсчета общего количества транзакций: %w", err)
	}

	return transactions, int(total), nil
}

// UpdateTransactionPaymentInfo обновляет информацию о платеже
func (r *MongoRepository) UpdateTransactionPaymentInfo(invoiceID string, amountPaid float64, currency string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"invoice_id": invoiceID}
	update := bson.M{
		"$set": bson.M{
			"amount_paid":      amountPaid,
			"payment_currency": currency,
			"updated_at":       time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("ошибка обновления информации о платеже: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("транзакция с invoice_id=%s не найдена", invoiceID)
	}

	r.logger.Info("Обновлена информация о платеже для транзакции %s: %f %s", invoiceID, amountPaid, currency)
	return nil
}
