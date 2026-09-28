package repository

import (
	"accroccator/internal/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type BulkRepository struct {
	bulkCollection *mongo.Collection
}

func NewBulkRepository(db *mongo.Database) *BulkRepository {
	return &BulkRepository{
		bulkCollection: db.Collection("bulk_last_update"),
	}
}

func (r *BulkRepository) FindLastSavedBulk() (*model.BulkSyncState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var lastBulkState model.BulkSyncState
	err := r.bulkCollection.FindOne(ctx, bson.M{}).Decode(&lastBulkState)

	if err == mongo.ErrNoDocuments {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &lastBulkState, nil
}

func (r *BulkRepository) UpdateLastBulk(bulkDto model.ScryfallBulkResponse) (*model.BulkSyncState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lastBulkState, err := r.FindLastSavedBulk()

	if err != nil {
		return nil, err
	}

	if lastBulkState == nil {
		lastBulkState, err = r.CreateFirstBulk(bulkDto)

		if err != nil {
			return nil, err
		}

		return lastBulkState, nil
	}

	newBulk := model.BulkSyncState{
		ID:        bulkDto.ID,
		UpdatedAt: bulkDto.UpdatedAt,
	}
	update := bson.M{"$set": newBulk}
	_, err = r.bulkCollection.UpdateOne(ctx, bson.M{}, update)

	if err != nil {
		return nil, err
	}

	return &newBulk, nil

}

func (r *BulkRepository) CreateFirstBulk(bulkDto model.ScryfallBulkResponse) (*model.BulkSyncState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	newBulk := model.BulkSyncState{
		ID:        bulkDto.ID,
		UpdatedAt: bulkDto.UpdatedAt,
	}

	_, err := r.bulkCollection.InsertOne(ctx, newBulk)
	if err != nil {
		return nil, err
	}

	return &newBulk, nil

}
