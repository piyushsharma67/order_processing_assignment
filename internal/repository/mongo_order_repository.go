package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"order_processing/internal/domain"
)

type MongoOrderRepository struct {
	collection *mongo.Collection
}

type orderDocument struct {
	ID         string             `bson:"_id"`
	CustomerID string             `bson:"customer_id"`
	Items      []domain.OrderItem `bson:"items"`
	Status     domain.OrderStatus `bson:"status"`
	Total      float64            `bson:"total"`
	CreatedAt  time.Time          `bson:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at"`
}

func NewMongoOrderRepository(ctx context.Context, uri, database string) (*MongoOrderRepository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	collection := client.Database(database).Collection("orders")
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "status", Value: 1}},
		Options: options.Index().SetName("status_idx"),
	}
	if _, err := collection.Indexes().CreateOne(ctx, indexModel); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	return &MongoOrderRepository{collection: collection}, nil
}

func (r *MongoOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	_, err := r.collection.InsertOne(ctx, toDocument(order))
	return err
}

func (r *MongoOrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var doc orderDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	return fromDocument(&doc), nil
}

func (r *MongoOrderRepository) Update(ctx context.Context, order *domain.Order) error {
	result, err := r.collection.ReplaceOne(ctx, bson.M{"_id": order.ID}, toDocument(order))
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return domain.ErrOrderNotFound
	}
	return nil
}

func (r *MongoOrderRepository) List(ctx context.Context, listFilter OrderListFilter) ([]*domain.Order, error) {
	filter := bson.M{}
	if listFilter.CustomerID != "" {
		filter["customer_id"] = listFilter.CustomerID
	}
	if listFilter.Status != nil {
		filter["status"] = *listFilter.Status
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []orderDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	orders := make([]*domain.Order, 0, len(docs))
	for i := range docs {
		orders = append(orders, fromDocument(&docs[i]))
	}
	return orders, nil
}

func (r *MongoOrderRepository) ListByStatus(ctx context.Context, status domain.OrderStatus) ([]*domain.Order, error) {
	return r.List(ctx, OrderListFilter{Status: &status})
}

func toDocument(order *domain.Order) orderDocument {
	items := make([]domain.OrderItem, len(order.Items))
	copy(items, order.Items)

	return orderDocument{
		ID:         order.ID,
		CustomerID: order.CustomerID,
		Items:      items,
		Status:     order.Status,
		Total:      order.Total,
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}

func fromDocument(doc *orderDocument) *domain.Order {
	items := make([]domain.OrderItem, len(doc.Items))
	copy(items, doc.Items)

	return &domain.Order{
		ID:         doc.ID,
		CustomerID: doc.CustomerID,
		Items:      items,
		Status:     doc.Status,
		Total:      doc.Total,
		CreatedAt:  doc.CreatedAt,
		UpdatedAt:  doc.UpdatedAt,
	}
}
