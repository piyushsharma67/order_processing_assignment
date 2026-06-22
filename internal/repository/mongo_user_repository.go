package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoUserRepository struct {
	collection *mongo.Collection
}

type userDocument struct {
	Username     string    `bson:"_id"`
	PasswordHash string    `bson:"password_hash"`
	CustomerID   string    `bson:"customer_id,omitempty"`
	CreatedAt    time.Time `bson:"created_at"`
}

func NewMongoUserRepository(ctx context.Context, uri, database string) (*MongoUserRepository, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	collection := client.Database(database).Collection("users")
	return &MongoUserRepository{collection: collection}, nil
}

func (r *MongoUserRepository) Create(ctx context.Context, username, passwordHash, customerID string) error {
	_, err := r.collection.InsertOne(ctx, userDocument{
		Username:     username,
		PasswordHash: passwordHash,
		CustomerID:   customerID,
		CreatedAt:    time.Now().UTC(),
	})
	if mongo.IsDuplicateKeyError(err) {
		return ErrUserExists
	}
	return err
}

func (r *MongoUserRepository) GetByUsername(ctx context.Context, username string) (UserRecord, error) {
	var doc userDocument
	err := r.collection.FindOne(ctx, bson.M{"_id": username}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return UserRecord{}, ErrUserNotFound
	}
	if err != nil {
		return UserRecord{}, err
	}
	return UserRecord{
		Username:     doc.Username,
		PasswordHash: doc.PasswordHash,
		CustomerID:   doc.CustomerID,
	}, nil
}

func (r *MongoUserRepository) SetCustomerID(ctx context.Context, username, customerID string) error {
	result, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": username},
		bson.M{"$set": bson.M{"customer_id": customerID}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrUserNotFound
	}
	return nil
}
