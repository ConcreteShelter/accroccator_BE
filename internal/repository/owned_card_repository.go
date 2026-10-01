package repository

import (
	"accroccator/internal/model"
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (r *CardRepository) EnsureOwnedIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	_, err := r.ownedCardsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "shop_id", Value: 1},
				{Key: "id", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *CardRepository) QueryOwned(ctx context.Context,
	scryfallCards []model.ScryfallCard,
	shopId primitive.ObjectID) ([]model.OwnedCardReference, error) {
	// GET ALL CARD IDS FROM THE PREVIOUS SLICE AND PUT THEM IN A SEPARATE SLICE
	var cardIds []string = make([]string, 0, len(scryfallCards))
	for _, card := range scryfallCards {
		cardIds = append(cardIds, card.ID)
	}

	// BULK QUERY FOR ALL THE CARDS IN THE PREVIOUS SLICE TO CHECK IF THEY ARE IN THE OWNED COLLECTION
	cursor, err := r.ownedCardsCollection.Find(ctx, bson.M{
		"id":      bson.M{"$in": cardIds},
		"shop_id": shopId,
	})
	if err != nil {
		return nil, err
	}

	var ownedCards []model.OwnedCardReference
	if err := cursor.All(ctx, &ownedCards); err != nil {
		return nil, err
	}

	return ownedCards, nil
}

func (r *CardRepository) SetOwnedCard(
	scryfallId string,
	containerInfos []model.QuantityUpdateRequest,
	shopID primitive.ObjectID) (*model.Card, error) {

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
		foundContainer, err := r.containerRepo.FindContainerById(info.ContainerID, shopID)

		if err != nil {
			return nil, errors.New("Invalid container Id in list")
		}

		ownership := model.InventoryOwnership{
			ContainerID: foundContainer.ID,
			Quantity:    info.Quantity,
		}

		quantityUpdates = append(quantityUpdates, ownership)

	}

	var ownedCard model.OwnedCardReference
	err = r.ownedCardsCollection.FindOne(ctx, bson.M{"id": scryfallId, "shop_id": shopID}).Decode(&ownedCard)

	newOwnedCard := model.Card{
		ScryfallCard: scryfallCard,
		ContainedIn:  quantityUpdates,
	}

	ownedCardReference := model.OwnedCardReference{
		ScryfallID:  scryfallId,
		ContainedIn: quantityUpdates,
		ShopID:      shopID,
	}

	switch err {
	case mongo.ErrNoDocuments:

		_, err := r.ownedCardsCollection.InsertOne(ctx, ownedCardReference)

		if err != nil {
			return nil, err
		}

		return &newOwnedCard, nil

	case nil:
		filter := bson.M{"id": scryfallId, "shop_id": shopID}
		update := bson.M{"$set": bson.M{"contained_in": quantityUpdates}}

		_, err = r.ownedCardsCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			return nil, err
		}

		return &newOwnedCard, nil

	default:
		return nil, err
	}

}

func (r *CardRepository) RemoveOwnedCard(id string, shopId primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"id": id, "shop_id": shopId}
	result, err := r.ownedCardsCollection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrCardNotFound // Or a custom error
	}

	return nil
}
