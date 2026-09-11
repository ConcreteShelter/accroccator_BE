package repository

import (
	"accroccator/internal/model"
	card_mapping "accroccator/internal/service/mapping"
	"context"
	"errors"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrCardNotFound = errors.New("no cards with that id were found")

type CardRepository struct {
	ownedCardsCollection    *mongo.Collection
	scryfallCardsCollection *mongo.Collection
}

// Constructor
func NewCardRepository(db *mongo.Database) *CardRepository {
	return &CardRepository{
		ownedCardsCollection:    db.Collection("owned_cards"),
		scryfallCardsCollection: db.Collection("scryfall_cards"),
	}
}

// FIND BY ID
func (r *CardRepository) FindCardByID(id string) ([]model.Card, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.scryfallCardsCollection.Find(ctx, bson.M{"id": id})

	if err != nil {
		log.Printf("Error searching cards with name '%s': %v", id, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var scryfallCards []model.ScryfallCard
	if err := cursor.All(ctx, &scryfallCards); err != nil {
		return nil, err
	}

	if len(scryfallCards) == 0 {
		return nil, ErrCardNotFound
	}

	ownedCards, err := r.QueryOwned(ctx, scryfallCards)
	if err != nil {
		return nil, err
	}

	result := card_mapping.MapOwned(scryfallCards, ownedCards)

	return result, nil

}

// FIND BY NAME
func (r *CardRepository) FindCardsByName(name string) ([]model.Card, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"name": bson.M{
			"$regex":   name,
			"$options": "i",
		},
	}

	cursor, err := r.scryfallCardsCollection.Find(ctx, filter)

	if err != nil {
		log.Printf("Error searching cards with name '%s': %v", name, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var scryfallCards []model.ScryfallCard
	if err := cursor.All(ctx, &scryfallCards); err != nil {
		return nil, err
	}

	if len(scryfallCards) == 0 {
		return nil, ErrCardNotFound
	}

	ownedCards, err := r.QueryOwned(ctx, scryfallCards)
	if err != nil {
		return nil, err
	}

	result := card_mapping.MapOwned(scryfallCards, ownedCards)

	return result, nil

}

func (r *CardRepository) GetAllCards(page int, pageSize int) ([]model.Card, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	skip := (page - 1) * pageSize

	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.scryfallCardsCollection.Find(ctx, bson.M{"digital": false}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// GET ALL CARDS FROM CARDS COLLECTION
	var scryfallCards []model.ScryfallCard
	if err := cursor.All(ctx, &scryfallCards); err != nil {
		return nil, err
	}

	ownedCards, err := r.QueryOwned(ctx, scryfallCards)
	if err != nil {
		return nil, err
	}

	result := card_mapping.MapOwned(scryfallCards, ownedCards)

	return result, nil
}

func (r *CardRepository) SetCardOwnedQuantity(id string, newQuantity int, containerId string) (*model.Card, error) {
	// CERCHIAMO LA CARTA SCRYFALL
	// 		SE NON C'È ERRORE
	//		SE C'È CERCHIAMOLA IN OWNED
	// 			SE C'È FACCIAMO UN UPDATE DELLA QUANTITY
	// 			SE NON C'È LA AGGIUNGIAMO IN REPO SETTANDO IL NUMERO

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// CERCHIAMO LA CARTA SCRYFALL
	var scryfallCard model.ScryfallCard
	err := r.scryfallCardsCollection.FindOne(ctx, bson.M{"id": id}).Decode(&scryfallCard)

	if err != nil {
		log.Printf("Error finding card with id '%s': %v", id, err)
		return nil, err
	}

	//	SE C'È CERCHIAMOLA IN OWNED

	if containerId != "" {

	}

	var ownedCard model.Card
	err = r.ownedCardsCollection.FindOne(ctx, bson.M{"id": id}).Decode(&ownedCard)

	filter := bson.M{"id": id}
	update := bson.M{"$set": bson.M{"quantity": newQuantity}}

	if err == nil {

		if newQuantity == 0 {
			err := r.RemoveOwnedCard(id)
			return nil, err
		}
		_, err := r.ownedCardsCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return nil, err
		}

		ownedCard.OwnerShip.ContainerQuantity = newQuantity
		return &ownedCard, nil
	} else if err == mongo.ErrNoDocuments {
		newOwnedCard := model.Card{
			ScryfallCard: scryfallCard,
			OwnerShip: model.OwnerShip{
				ContainerQuantity: newQuantity,
			},
		}

		insertResult, err := r.ownedCardsCollection.InsertOne(ctx, newOwnedCard)
		newID := insertResult.InsertedID.(primitive.ObjectID)
		newOwnedCard.OwnerShip.ContainerID = &newID
		if err != nil {
			return nil, err
		}

		return &newOwnedCard, nil
	} else {
		return nil, err
	}

}

func (r *CardRepository) AdvancedCardSearch(filters model.CardAdvancedSearchRequest) ([]model.Card, error) {
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"name": bson.M{
			"$regex":   filters.Name,
			"$options": "i",
		},
	}

	if len(filters.Keywords) > 0 {
		filter["keywords"] = bson.M{"$all": filters.Keywords}
	}

	return []model.Card{}, nil
}
