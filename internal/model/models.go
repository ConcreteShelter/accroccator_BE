package model

import "go.mongodb.org/mongo-driver/bson/primitive"

var PlaceholderShopID primitive.ObjectID

func init() {
	var err error
	PlaceholderShopID, err = primitive.ObjectIDFromHex("111111111111111111111111")
	if err != nil {
		panic("wrong hex!")
	}
}

type Container struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name   string             `bson:"name" json:"name"`
	Type   string             `bson:"type" json:"type"`
	ShopID primitive.ObjectID `bson:"shop_id" json:"shop_id"`
}

type ScryfallCard struct {
	ID              string            `bson:"id" json:"id"`
	Name            string            `bson:"name" json:"name"`
	TypeLine        string            `bson:"type_line" json:"type_line"`
	ImageURIs       map[string]string `bson:"image_uris" json:"image_uris"`
	Set             string            `bson:"set" json:"set"`
	SetName         string            `bson:"set_name" json:"set_name"`
	CollectorNumber string            `bson:"collector_number" json:"collector_number"`
	Lang            string            `bson:"lang" json:"lang"`
	ReleasedAt      string            `bson:"released_at" json:"released_at"`
}

type Card struct {
	ScryfallCard `bson:",inline"`
	ContainedIn  []InventoryOwnership `bson:"contained_in" json:"contained_in"`
}

type OwnedCardReference struct {
	ScryfallID  string               `bson:"id" json:"id"`
	ContainedIn []InventoryOwnership `bson:"contained_in" json:"contained_in"`
	ShopID      primitive.ObjectID   `bson:"shop_id" json:"shop_id"`
}

type InventoryOwnership struct {
	ContainerID primitive.ObjectID `bson:"container_id" json:"container_id"`
	Quantity    int                `bson:"quantity" json:"quantity"`
}

type CardNameDoc struct {
	NameID string `bson:"_id"`
}
