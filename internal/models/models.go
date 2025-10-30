package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingStatus string

const (
	BookingStatusReserved  BookingStatus = "reserved"
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusExpired   BookingStatus = "expired"
)

type Booking struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserID  primitive.ObjectID `json:"user_id" bson:"user_id"`
	EventID primitive.ObjectID `json:"event_id" bson:"event_id"`
	Status  BookingStatus      `json:"status" bson:"status"`
	Tickets []BookingTicket    `json:"tickets" bson:"tickets"`

	PaymentID   string  `json:"payment_id" bson:"payment_id"`
	Subtotal    float64 `json:"subtotal" bson:"subtotal"`
	ServiceFree float64 `json:"service_free" bson:"service_free"`
	TotalAmount float64 `json:"total_amount" bson:"total_amount"`
	Currency    string  `json:"currency" bson:"currency"`

	ReservedUntil time.Time `json:"reserved_until" bson:"reserved_until"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type BookingTicket struct {
	TicketTypeID   primitive.ObjectID `json:"ticket_type_id" bson:"ticket_type_id"`
	TicketTypeName string             `json:"ticket_type_name" bson:"ticket_type_name"`
	Quantity       int                `json:"quantity" bson:"quantity"`
	UnitPrice      float64            `json:"unit_price" bson:"unit_price"`
	TotalPrice     float64            `json:"total_price" bson:"total_price"`
	Seats          []Seat             `json:"seats,omitempty" bson:"seats,omitempty"`
}

type Seat struct {
	Sector string `json:"sector" bson:"sector"`
	Row    string `json:"row" bson:"row"`
	Number string `json:"number" bson:"number"`
	SeatID string `json:"seat_id" bson:"seat_id"`
}

type BookingRequest struct {
	EventID string            `json:"event_id"`
	Tickets []TicketSelection `json:"tickets"`
	UserID  string            `json:"-"`
}

type TicketSelection struct {
	TicketID string `json:"ticket_id"`
	Quantity int    `json:"quantity"`
	Seats    []Seat `json:"seats,omitempty"`
}

type BookingResponse struct {
	BookingID     string          `json:"booking_id"`
	Status        BookingStatus   `json:"status"`
	ReservedUntil time.Time       `json:"reserved_until"`
	TotalAmount   float64         `json:"total_amount"`
	Tickets       []BookingTicket `json:"tickets"`
}

type Event struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Date        time.Time          `json:"date" bson:"date"`
	TicketTypes []TicketType       `json:"ticket_type" bson:"ticket_type"`
}

type TicketType struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	Quantity  int                `json:"quantity" bson:"quantity"`
	SoldCount int                `json:"sold_count" bson:"sold_count"`
	Price     float64            `json:"price" bson:"price"`
}
