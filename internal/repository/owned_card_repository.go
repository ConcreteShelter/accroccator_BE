package repository

import (
	"accroccator/internal/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func (r *CardRepository) QueryOwned(ctx context.Context, scryfallCards []model.ScryfallCard) ([]model.Card, error) {
	// GET ALL CARD IDS FROM THE PREVIOUS SLICE AND PUT THEM IN A SEPARATE SLICE
	var cardIds []string = make([]string, 0, len(scryfallCards))
	for _, card := range scryfallCards {
		cardIds = append(cardIds, card.ID)
	}

	// BULK QUERY FOR ALL THE CARDS IN THE PREVIOUS SLICE TO CHECK IF THEY ARE IN THE OWNED COLLECTION
	cursor, err := r.ownedCardsCollection.Find(ctx, bson.M{
		"id": bson.M{"$in": cardIds},
	})
	if err != nil {
		return nil, err
	}

	var ownedCards []model.Card
	if err := cursor.All(ctx, &ownedCards); err != nil {
		return nil, err
	}

	return ownedCards, nil
}

func (r *CardRepository) RemoveOwnedCard(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"id": id}
	result, err := r.ownedCardsCollection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrCardNotFound // Or a custom error
	}

	return nil
}
