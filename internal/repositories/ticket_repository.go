package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TicketRepository struct {
	collection *mongo.Collection
}

func NewTicketRepository(db *mongo.Database) *TicketRepository {
	return &TicketRepository{
		collection: db.Collection("events"),
	}
}

func (tr *TicketRepository) ReserveTickets(ctx context.Context, eventID, ticketTypeID primitive.ObjectID, quantity int) error {
	_, err := tr.collection.UpdateOne(
		ctx,
		bson.M{
			"_id":              eventID,
			"ticket_types._id": ticketTypeID,
			"ticket_types.sold_count": bson.M{
				"$lte": bson.M{"&subtract": []interface{}{"$ticket_types.quantity", quantity}},
			},
		},
		bson.M{
			"$inc": bson.M{"ticket_types.$.sold_count": quantity},
		},
	)
	return err

}

func (tr *TicketRepository) ReleaseTickets(ctx context.Context, eventID, tickeTypeID primitive.ObjectID, quantity int) error {
	_, err := tr.collection.UpdateOne(
		ctx, bson.M{"_id": eventID,
			"ticket_types._id": tickeTypeID},
		bson.M{"$inc": bson.M{"ticket_types.$.sold_count": -quantity}},
	)
	return err
}

func (tr *TicketRepository) ConfirmSale(ctx context.Context, eventID, ticketTypeID primitive.ObjectID, quantity int) error {

	// Тут логику пока не продумал((

	return nil
}
