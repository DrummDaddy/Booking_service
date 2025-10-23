package repositories

import (
	"context"

	"github.com/DrummDaddy/Booking_service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EventRepository struct {
	coolection *mongo.Collection
}

func NewEventRepository(db *mongo.Database) *EventRepository {
	return &EventRepository{
		coolection: db.Collection("events"),
	}
}

func (er *EventRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Event, error) {
	var event models.Event

	err := er.coolection.FindOne(ctx, bson.M{"_id": id}).Decode(&event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}
