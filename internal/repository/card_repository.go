package repository

import (
	"accroccator/internal/model"
	card_mapping "accroccator/internal/service/mapping"
	"context"
	"errors"
	"log"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	// "go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrCardNotFound = errors.New("no cards with that id were found")

type CardRepository struct {
	ownedCardsCollection    *mongo.Collection
	scryfallCardsCollection *mongo.Collection
	containerRepo           *ContainerRepository
	cardNamesCollection     *mongo.Collection
}

// Constructor
func NewCardRepository(db *mongo.Database, containerRepo *ContainerRepository) *CardRepository {
	return &CardRepository{
		ownedCardsCollection:    db.Collection("owned_cards"),
		scryfallCardsCollection: db.Collection("scryfall_cards"),
		containerRepo:           containerRepo,
		cardNamesCollection:     db.Collection("card_names"),
	}
}

// FIND BY ID
func (r *CardRepository) FindCardByID(id string, shopId primitive.ObjectID) ([]model.Card, error) {
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

	ownedCards, err := r.QueryOwned(ctx, scryfallCards, shopId)
	if err != nil {
		return nil, err
	}

	result := card_mapping.MapOwned(scryfallCards, ownedCards)

	return result, nil

}

// FIND BY NAME
func (r *CardRepository) FindCardsByName(name string, shopId primitive.ObjectID) ([]model.Card, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"name": name,
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

	ownedCards, err := r.QueryOwned(ctx, scryfallCards, shopId)
	if err != nil {
		return nil, err
	}

	result := card_mapping.MapOwned(scryfallCards, ownedCards)

	return result, nil

}

func (r *CardRepository) GetAllCards(page int, pageSize int, shopId primitive.ObjectID) ([]model.Card, error) {
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

	ownedCards, err := r.QueryOwned(ctx, scryfallCards, shopId)
	if err != nil {
		return nil, err
	}

	result := card_mapping.MapOwned(scryfallCards, ownedCards)

	return result, nil
}

func (r *CardRepository) SearchCardNames(queryString string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id": bson.M{
			"$regex":   "^" + regexp.QuoteMeta(queryString),
			"$options": "i",
		},
	}

	opts := options.Find().SetLimit(20)

	cursor, err := r.cardNamesCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cardNameDocs []model.CardNameDoc
	err = cursor.All(ctx, &cardNameDocs)
	if err != nil {
		return nil, err
	}

	stringResults := make([]string, 0, len(cardNameDocs))
	for _, cardNameDoc := range cardNameDocs {
		stringResults = append(stringResults, cardNameDoc.NameID)
	}

	if len(cardNameDocs) < 20 {
		difToFill := 20 - len(cardNameDocs)

		opts = options.Find().SetLimit(int64(difToFill))

		filter = bson.M{
			"_id": bson.M{
				"$regex":   regexp.QuoteMeta(queryString),
				"$options": "i",
				"$nin":     stringResults,
			},
		}

		cursor, err := r.cardNamesCollection.Find(ctx, filter, opts)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		var remainingCardNames []model.CardNameDoc

		err = cursor.All(ctx, &remainingCardNames)
		if err != nil {
			return nil, err
		}

		for _, cardNameDoc := range remainingCardNames {
			stringResults = append(stringResults, cardNameDoc.NameID)
		}
	}

	return stringResults, nil

}

// func (r *CardRepository) AdvancedCardSearch(filters model.CardAdvancedSearchRequest, shopId primitive.ObjectID) ([]model.Card, error) {
// 	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	filter := bson.M{
// 		"name": bson.M{
// 			"$regex":   filters.Name,
// 			"$options": "i",
// 		},
// 	}

// 	if len(filters.Keywords) > 0 {
// 		filter["keywords"] = bson.M{"$all": filters.Keywords}
// 	}

// 	return []model.Card{}, nil
// }
