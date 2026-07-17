package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cinema-booking/internal/config"
	"cinema-booking/internal/handler"
	"cinema-booking/internal/messaging"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/repository"
	"cinema-booking/internal/service"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	// load .env config
	cfg := config.Load()

	// connect mongodb
	mongoDB := connectMongo(cfg)
	defer mongoDB.Client().Disconnect(context.Background())

	// connect redis
	redisClient := connectRedis(cfg)
	defer redisClient.Close()

	// connect rabbitmq
	rabbitConn := connectRabbitMQ(cfg)
	defer rabbitConn.Close()

	// connect firebase
	firebaseAuth := config.InitFirebase(cfg.FirebaseCredFile)

	// wire dependencies

	movieRepo := repository.NewMovieRepository(mongoDB)
	showtimeRepo := repository.NewShowtimeRepository(mongoDB)
	userRepo := repository.NewUserRepository(mongoDB)
	bookingRepo := repository.NewBookingRepository(mongoDB)

	movieService := service.NewMovieService(movieRepo)
	showtimeService := service.NewShowtimeService(showtimeRepo)
	lockService := service.NewLockService(redisClient)

	publisher, err := messaging.NewBookingPublisher(rabbitConn)
	if err != nil {
		log.Fatalf("Failed to setup booking publisher: %v", err)
	}

	consumer := messaging.NewBookingConsumer(rabbitConn, bookingRepo)
	go consumer.Start()

	bookingService := service.NewBookingService(bookingRepo, showtimeRepo, lockService, publisher)

	movieHandler := handler.NewMovieHandler(movieService)
	showtimeHandler := handler.NewShowtimeHandler(showtimeService)
	bookingHandler := handler.NewBookingHandler(bookingService)

	// middleware
	authMiddleware := middleware.AuthMiddleware(firebaseAuth, userRepo)

	// create Gin router
	r := setupRouter(movieHandler, showtimeHandler, bookingHandler, authMiddleware)

	// start server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("Cinema Booking API running on %s", addr)

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// helpers
func setupRouter(
	movieHandler *handler.MovieHandler,
	showtimeHandler *handler.ShowtimeHandler,
	bookingHandler *handler.BookingHandler,
	authMiddleware gin.HandlerFunc,
) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "cinema-booking-api",
		})
	})

	api := r.Group("/api/v1")
	{
		movies := api.Group("/movies")
		{
			movies.GET("", movieHandler.GetAll)
			movies.GET("/:id", movieHandler.GetByID)
		}

		showtimes := api.Group("/showtimes")
		{
			showtimes.GET("", showtimeHandler.GetAll)
			showtimes.GET("/:id", showtimeHandler.GetByID)
		}

		protected := api.Group("")
		protected.Use(authMiddleware)
		{
			protected.POST("/movies", movieHandler.Create)
			protected.POST("/showtimes", showtimeHandler.Create)
			protected.POST("/bookings/lock", bookingHandler.LockSeats)
			protected.POST("/bookings", bookingHandler.CreateBooking)
		}
	}

	return r
}

func connectMongo(cfg *config.Config) *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB")
	return client.Database(cfg.MongoDBName)
}

func connectRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPass,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Connected to Redis")
	return client
}

func connectRabbitMQ(cfg *config.Config) *amqp.Connection {
	conn, err := amqp.Dial(cfg.RabbitMQURL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	log.Println("Connected to RabbitMQ")
	return conn
}
