package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/DrummDaddy/Booking_service/internal/handlers"
	"github.com/DrummDaddy/Booking_service/internal/repositories"
	"github.com/DrummDaddy/Booking_service/internal/services"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("booking")

	bookingRepo := repositories.NewBookingRepository(db)
	eventRepo := repositories.NewEventRepository(db)
	ticketRepo := repositories.NewTicketRepository(db)

	// Это заглушка надо будет поставить конфиг платежного сервиса
	paymentService := services.NewPaymentService("https://some-api", "demo")

	bookingServise := services.NewBookingService(bookingRepo, eventRepo, ticketRepo, paymentService)
	bookingHandler := &handlers.BookingHandler{Service: bookingServise}

	http.HandleFunc("/api/bookings", bookingHandler.CreateBooking)
	http.HandleFunc("api/payments", bookingHandler.CreatePayment)
	http.HandleFunc("/api/payments/webhook", bookingServise.HandlerWebhook)

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
