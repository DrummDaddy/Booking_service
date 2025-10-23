package repositories

import (
	"context"
	"time"

	"github.com/DrummDaddy/Booking_service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type BookingRepository struct {
	collection *mongo.Collection
}

func NewBookingRepository(db *mongo.Database) *BookingRepository {
	return &BookingRepository{
		collection: db.Collection("bookings"),
	}
}

func (br *BookingRepository) Create(ctx context.Context, booking *models.Booking) error {
	_, err := br.collection.InsertOne(ctx, booking)
	return err
}

func (br *BookingRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Booking, error) {
	var booking models.Booking
	err := br.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&booking)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}

func (br *BookingRepository) FindByIDAndUser(ctx context.Context, id, userID primitive.ObjectID) (*models.Booking, error) {
	var booking models.Booking
	err := br.collection.FindOne(ctx, bson.M{"_id": id, "user_id": userID}).Decode(&booking)
	if err != nil {
		return nil, err
	}

	return &booking, nil
}

func (br *BookingRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status models.BookingStatus) error {
	_, err := br.collection.UpdateOne(ctx, bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"status":     status,
				"updated_at": time.Now(),
			},
		},
	)
	return err
}

func (br *BookingRepository) FindExpiredReservation(ctx context.Context) ([]models.Booking, error) {
	cursor, err := br.collection.Find(
		ctx,
		bson.M{
			"status":         models.BookingStatusReserved,
			"reserved_until": bson.M{"$lt": time.Now()},
		},
	)

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bookings []models.Booking

	if err := cursor.All(ctx, &bookings); err != nil {
		return nil, err
	}
	return bookings, nil
}
