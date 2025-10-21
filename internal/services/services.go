package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DrummDaddy/Booking_service/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BookingService struct {
	bookingRepo    *BookingRepository
	eventRepo      *EventRepository
	ticketRepo     *TicketRepository
	paymentService *PaymentService
	cache          *RedisCache
	reservationTTL time.Duration
}

func NewBookingService(
	bookingRepo *BookingRepository,
	eventRepo *EventRepository,
	ticketRepo *TicketRepository,
	paymentService *PaymentService,
) *BookingService {
	return &BookingService{
		bookingRepo:    bookingRepo,
		eventRepo:      eventRepo,
		ticketRepo:     ticketRepo,
		paymentService: paymentService,
		reservationTTL: 15 * time.Minute,
	}
}

func (bs *BookingService) CreatingBooking(ctx context.Context, req *models.BookingRequest) (*models.BookingResponse, error) {
	if err := bs.validateBookingRequest(ctx, req); err != nil {
		return nil, err
	}

	eventObjID, _ := primitive.ObjectIDFromHex(req.EventID)
	event, err := bs.eventRepo.FindByID(ctx, eventObjID)
	if err != nil {
		return nil, errors.New("event not found")
	}
	reservedTickets, err := bs.reservedTickets(ctx, event, req.Tickets)
	if err != nil {
		return nil, err
	}

	booking := &models.Booking{
		ID:            primitive.NewObjectID(),
		USerID:        primitive.ObjectIDFromHex(req.UserID),
		EventID:       eventObjID,
		Status:        models.BookingStatusReserved,
		Tickets:       reservedTickets,
		Subtotal:      bs.calculateSubtotal(reservedTickets),
		ServiceFree:   bs.calculateServiceFee(reservedTickets),
		TotalAmount:   0,
		Currency:      "RUB",
		ReservedUntil: time.Now().Add(bs.reservationTTL),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	booking.TotalAmount = booking.Subtotal + booking.ServiceFree

	if err := bs.bookingRepo.Create(ctx, booking); err != nil {
		bs.relaseTickets(ctx, eventObjID, reservedTickets)
		return nil, err
	}

	response := &models.BookingResponse{
		BookingID:     booking.ID.Hex(),
		Status:        booking.Status,
		ReservedUntil: booking.ReservedUntil,
		TotalAmount:   booking.TotalAmount,
		Tickets:       booking.Tickets,
	}
	return response, nil
}

func (bs *BookingService) validateBookingRequest(ctx context.Context, req *models.BookingRequest) error {
	if req.UserID == "" {
		return errors.New("user ID is required")
	}
	if req.EventID == "" {
		return errors.New("event ID is required")
	}

	if len(req.Tickets) == 0 {
		return errors.New("at least one ticket required")
	}

	totalTickets := 0
	for _, ticket := range req.Tickets {
		if ticket.Quantity <= 0 {
			return errors.New("invalid ticket quantity")
		}
		totalTickets += ticket.Quantity
	}
	if totalTickets > 10 {
		return errors.New("maximum 10 tickets per order")
	}

	return nil
}

func (bs *BookingService) reserveTickets(ctx context.Context, event *models.Event, ticketSelections []models.TicketSelection) ([]models.BookingTicket, error) {
	var bookingTickets []models.BookingTicket

	for _, selection := range ticketSelections {
		ticketTypeID, _ := primitive.ObjectIDFromHex(selection.TicketTypeID)
		var ticketType *models.TicketType
		for i, tt := range event.TcketTypes {
			if tt.ID == ticketTypeID {
				ticketType = &event.TicketTypes[i]
				break
			}
		}
		if ticketType == nil {
			return nil, fmt.Errorf("ticket typr %s not found", selection.TicketTypeID)
		}

		avalible := ticketType.Quantity - ticketType.SoldCount
		if avalible < selection.Quantity {
			return nil, fmt.Errorf("not enough tickets avalible for %s", &ticketType.Name)
		}
		if err := bs.ticketRepo.ReserveTickets(ctx, event.ID, ticketTypeID, selection.Quantity); err != nil {
			return nil, err
		}

		bookingTicket := models.BookingTicket{
			TicketTypeID:   ticketTypeID,
			TicketTypeName: ticketType.Name,
			Quantity:       selection.Quantity,
			UnitPrice:      ticketType.Price,
			TotalPrice:     ticketType.Price * float64(selection.Quantity),
			Seats:          selection.Seats,
		}

		bookingTickets = append(bookingTickets, bookingTicket)
	}
	return bookingTickets, nil
}

func (bs *BookingService) calculateSubtotal(tickets []models.BookingTicket) float64 {
	var subtotal float64
	for _, ticket := range tickets {
		subtotal += ticket.TotalPrice
	}
	return subtotal
}

func (bs *BookingService) calculateServiceFee(tickets []models.BookingTicket) float64 {
	var serviceFee float64
	for _, ticket := range tickets {
		fee := ticket.TotalPrice * 0.1
		if fee < 50*float64(ticket.Quantity) {
			fee = 50 * float64(ticket.Quantity)
		}
		serviceFee += fee
	}

	return serviceFee
}

func (bs *BookingService) releaseTickets(ctx context.Context, eventID primitive.ObjectID, tickets []models.BookingTicket) {
	for _, ticket := range tickets {
		bs.ticketRepo.ReleaseTickets(ctx, eventID, ticket.TicketTypeID, ticket.Quantity)
	}
}

func (bs *BookingService) Getbooking(ctx context.Context, bookingID string, userID string) (*models.Booking, error) {
	bookingObjID, _ := primitive.ObjectIDFromHex(bookingID)
	userObjID, _ := primitive.ObjectIDFromHex(userID)

	booking, err := bs.bookingRepo.FindByIDAndUser(ctx, bookingObjID, userObjID)
	if err != nil {
		return nil, err
	}

	return booking, nil
}

func (bs *BookingService) ConfirmBooking(ctx context.Context, bookingID string, paymentID string) error {
	bookingObjID, _ := primitive.ObjectIDFromHex(bookingID)

	booking, err := bs.bookingRepo.FindByID(ctx, bookingObjID)
	if err != nil {
		return err
	}

	if booking.Status != models.BookingStatusReserved {
		return errors.New(" booking is not in reserved status")
	}

	if time.Now().After(booking.ReservedUntil) {
		bs.CancelBooking(ctx, bookingID, "reservation expires")
	}

	if err := bs.bookingRepo.UpdateStatus(ctx, bookingObjID, models.BookingStatusConfirmed); err != nil {
		return err
	}

	for _, ticket := range booking.Tickets {
		if err := bs.ticketRepo.ConfirmSale(ctx, booking.EventID, ticket.TicketTypeID, ticket.Quantity); err != nil {
			return err
		}
	}
	return nil
}

func (bs *BookingService) CancelBooking(ctx context.Context, bookingID string, reason string) error {
	bookingObjID, _ := primitive.ObjectIDFromHex(bookingID)

	booking, err := bs.bookingRepo.FindByID(ctx, bookingObjID)
	if err != nil {
		return err
	}

	bs.releaseTickets(ctx, booking.EventID, booking.Tickets)

	if err := bs.bookingRepo.UpdateStatus(ctx, bookingObjID, models.BookingStatusCancelled); err != nil {
		return err
	}

	return nil
}
