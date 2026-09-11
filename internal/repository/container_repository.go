package repository

import (
	"accroccator/internal/model"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrContainerNotFound = errors.New("no containers with that id were found")

type ContainerRepository struct {
	containersCollection *mongo.Collection
}

func NewContainerRepository(db *mongo.Database) *ContainerRepository {
	return &ContainerRepository{
		containersCollection: db.Collection("containers"),
	}
}

func (r *ContainerRepository) FindContainerById(id string) (*model.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid ID format")
	}

	var container model.Container
	err = r.containersCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&container)

	if err == mongo.ErrNoDocuments {
		return nil, ErrContainerNotFound
	}

	if err != nil {
		return nil, err
	}

	return &container, nil

}

func (r *ContainerRepository) GetAllContainers() ([]model.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.containersCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var containers []model.Container
	if err := cursor.All(ctx, &containers); err != nil {
		return nil, err
	}

	return containers, nil
}

func (r *ContainerRepository) CreateContainer(
	dto model.CreateContainerRequest) (*model.Container, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	container := model.Container{
		Name: dto.Name,
		Type: dto.Type,
	}

	result, err := r.containersCollection.InsertOne(ctx, container)
	if err != nil {
		return nil, err
	}

	container.ID = result.InsertedID.(primitive.ObjectID)

	return &container, nil

}
