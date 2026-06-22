package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"order_processing/internal/domain"
)

var ErrProductNotFound = errors.New("product not found")

type MongoProductRepository struct {
	collection *mongo.Collection
}

type productDocument struct {
	ID          string  `bson:"_id"`
	Name        string  `bson:"name"`
	Description string  `bson:"description"`
	Category    string  `bson:"category"`
	Price       float64 `bson:"price"`
	ImageURL    string  `bson:"image_url"`
	Stock       int     `bson:"stock"`
}

func NewMongoProductRepository(ctx context.Context, uri, database string) (*MongoProductRepository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	collection := client.Database(database).Collection("products")
	return &MongoProductRepository{collection: collection}, nil
}

func (r *MongoProductRepository) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}

func (r *MongoProductRepository) CreateMany(ctx context.Context, products []domain.Product) error {
	if len(products) == 0 {
		return nil
	}

	docs := make([]any, len(products))
	for i, product := range products {
		docs[i] = toProductDocument(product)
	}

	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

func (r *MongoProductRepository) List(ctx context.Context) ([]domain.Product, error) {
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "name", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []productDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	products := make([]domain.Product, len(docs))
	for i, doc := range docs {
		products[i] = fromProductDocument(doc)
	}
	return products, nil
}

func (r *MongoProductRepository) GetByID(ctx context.Context, id string) (*domain.Product, error) {
	var doc productDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrProductNotFound
	}
	if err != nil {
		return nil, err
	}
	product := fromProductDocument(doc)
	return &product, nil
}

func toProductDocument(product domain.Product) productDocument {
	return productDocument{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Category:    product.Category,
		Price:       product.Price,
		ImageURL:    product.ImageURL,
		Stock:       product.Stock,
	}
}

func fromProductDocument(doc productDocument) domain.Product {
	return domain.Product{
		ID:          doc.ID,
		Name:        doc.Name,
		Description: doc.Description,
		Category:    doc.Category,
		Price:       doc.Price,
		ImageURL:    doc.ImageURL,
		Stock:       doc.Stock,
	}
}
