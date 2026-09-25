package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Container struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string             `bson:"name" json:"name"`
	Type string             `bson:"type" json:"type"`
}

type ScryfallCard struct {
	ID        string            `bson:"id" json:"id"`
	Name      string            `bson:"name" json:"name"`
	TypeLine  string            `bson:"type_line" json:"type_line"`
	ImageURIs map[string]string `bson:"image_uris" json:"image_uris"`
}

type Card struct {
	ScryfallCard `bson:",inline"`
	ContainedIn  []InventoryOwnership `bson:"contained_in" json:"contained_in"`
}

type InventoryOwnership struct {
	ContainerID primitive.ObjectID `bson:"container_id" json:"container_id"`
	Quantity    int                `bson:"quantity" json:"quantity"`
}
