package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cinema-booking/internal/config"
	"cinema-booking/internal/handler"
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

	// wire dependencies
	movieRepo := repository.NewMovieRepository(mongoDB)
	showtimeRepo := repository.NewShowTimeRepository(mongoDB)

	movieService := service.NewMovieService(movieRepo)
	showtimeService := service.NewShowtimeService(showtimeRepo)

	movieHandler := handler.NewMovieHandler(movieService)
	showtimeHandler := handler.NewShowtimeHandler(showtimeService)

	// create Gin router
	r := setupRouter(movieHandler, showtimeHandler)

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
			movies.POST("", movieHandler.Create)
			movies.GET("", movieHandler.GetAll)
			movies.GET("/:id", movieHandler.GetByID)
		}

		showtimes := api.Group("/showtimes")
		{
			showtimes.POST("", showtimeHandler.Create)
			showtimes.GET("", showtimeHandler.GetAll)
			showtimes.GET("/:id", showtimeHandler.GetByID)
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
