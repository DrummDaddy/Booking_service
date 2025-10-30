package repositories

import (
	"context"
	"errors"
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

func (br *BookingRepository) FindByPaymentID(ctx context.Context, paymentID string) (*models.Booking, error) {
	var booking models.Booking

	err := br.collection.FindOne(ctx, bson.M{"payment_id": paymentID}).Decode(&booking)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("booking with the given payment ID not found")

		}
		return nil, err
	}

	return &booking, nil
}

func (br *BookingRepository) UpdatePaymentID(ctx context.Context, bookingID primitive.ObjectID, paymentID string) error {
	_, err := br.collection.UpdateOne(ctx, bson.M{"_id": bookingID}, bson.M{"$set": bson.M{"payment_id": paymentID}})

	return err
}

func (br *BookingRepository) CreateIndexes(ctx context.Context) error {
	indexModel := mongo.IndexModel{
		Keys: bson.M{"payment_id": 1},
	}
	_, err := br.collection.Indexes().CreateOne(ctx, indexModel)

	return err
}
