package repository

import (
	"accroccator/internal/model"
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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

func (r *CardRepository) SetOwnedCard(
	scryfallId string,
	containerInfos []model.QuantityUpdateRequest) (*model.Card, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var scryfallCard model.ScryfallCard
	err := r.scryfallCardsCollection.FindOne(ctx, bson.M{"id": scryfallId}).Decode(&scryfallCard)

	if err != nil {
		log.Printf("Error finding card with id '%s': %v", scryfallId, err)
		return nil, err
	}

	//check if the containerInfos are correct
	var quantityUpdates []model.InventoryOwnership = make([]model.InventoryOwnership, 0)
	for _, info := range containerInfos {
		foundContainer, err := r.containerRepo.FindContainerById(info.ContainerID)

		if err != nil {
			return nil, errors.New("Invalid container Id in list")
		}

		ownership := model.InventoryOwnership{
			ContainerID: foundContainer.ID,
			Quantity:    info.Quantity,
		}

		quantityUpdates = append(quantityUpdates, ownership)

	}

	var ownedCard model.Card
	err = r.ownedCardsCollection.FindOne(ctx, bson.M{"id": scryfallId}).Decode(&ownedCard)

	switch err {
	case mongo.ErrNoDocuments:
		newOwnedCard := model.Card{
			ScryfallCard: scryfallCard,
			ContainedIn:  quantityUpdates,
		}

		_, err := r.ownedCardsCollection.InsertOne(ctx, newOwnedCard)

		if err != nil {
			return nil, err
		}

		return &newOwnedCard, nil

	case nil:
		filter := bson.M{"id": scryfallId}
		update := bson.M{"$set": bson.M{"contained_in": quantityUpdates}}

		_, err = r.ownedCardsCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			return nil, err
		}

		newOwnedCard := model.Card{
			ScryfallCard: scryfallCard,
			ContainedIn:  quantityUpdates,
		}

		return &newOwnedCard, nil

	default:
		return nil, err
	}

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
