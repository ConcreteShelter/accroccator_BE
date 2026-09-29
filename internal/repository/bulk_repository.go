package repository

import (
	"accroccator/internal/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BulkRepository struct {
	bulkCollection       *mongo.Collection
	scryfallCardsStaging *mongo.Collection
}

func NewBulkRepository(db *mongo.Database) *BulkRepository {
	return &BulkRepository{
		bulkCollection:       db.Collection("bulk_last_update"),
		scryfallCardsStaging: db.Collection("scryfall_cards_staging"),
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

	newBulk := model.BulkSyncState{
		ID:        bulkDto.ID,
		UpdatedAt: bulkDto.UpdatedAt,
	}

	update := bson.M{"$set": newBulk}
	opts := options.Update().SetUpsert(true)
	_, err := r.bulkCollection.UpdateOne(ctx, bson.M{}, update, opts)

	if err != nil {
		return nil, err
	}

	return &newBulk, nil

}

func (r *BulkRepository) InsertBatches(batch []any) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := r.scryfallCardsStaging.InsertMany(ctx, batch)
	if err != nil {
		return err
	}

	return nil
}

func (r *BulkRepository) DropStaging() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := r.scryfallCardsStaging.Drop(ctx)

	if err != nil {
		return err
	}

	return nil

}

func (r *BulkRepository) CreateStagingIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err := r.scryfallCardsStaging.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},   // field "id", ascending (the "1 (asc)" in Compass)
			Options: options.Index().SetUnique(true), // the "unique" checkbox
		},
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index(),
		},
	})

	if err != nil {
		return err
	}

	return nil
}

func (r *BulkRepository) BuildStagingNames() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$name"}}}},
		{{Key: "$out", Value: "card_names_staging"}},
	}

	cursor, err := r.scryfallCardsStaging.Aggregate(ctx, pipeline)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	return nil

}

func (r *BulkRepository) SwapStaging() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cardStagingName := r.scryfallCardsStaging.Database().Name()

	cmd := bson.D{
		{Key: "renameCollection", Value: cardStagingName},
		{Key: "to", Value: "accroccator_db.scryfall_cards"},
		{Key: "dropTarget", Value: true},
	}

	adminDB := r.scryfallCardsStaging.Database().Client().Database("admin")
	err := adminDB.RunCommand(ctx, cmd).Err()

	if err != nil {
		return err
	}

	return nil
}
