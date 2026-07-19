package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"cinema-booking/internal/config"
	"cinema-booking/internal/model"
	"cinema-booking/internal/repository"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	db := client.Database(cfg.MongoDBName)
	movieRepo := repository.NewMovieRepository(db)
	showtimeRepo := repository.NewShowtimeRepository(db)

	movies := []*model.Movie{
		{
			Title:       "The Avengers",
			Description: "Earth's mightiest heroes must come together to stop Loki and his alien army from enslaving humanity.",
			Duration:    143,
			PosterURL:   "https://m.media-amazon.com/images/M/MV5BNGE0YTVjNzUtNzJjOS00NGNlLTgxMzctZTY4YTE1Y2Y1ZTU4XkEyXkFqcGc@._V1_.jpg",
		},
		{
			Title:       "Inception",
			Description: "A skilled thief is offered a chance to have his past crimes forgiven if he can successfully perform inception — planting an idea into someone's mind.",
			Duration:    148,
			PosterURL:   "https://image.tmdb.org/t/p/original/xlaY2zyzMfkhk0HSC5VUwzoZPU1.jpg",
		},
	}

	// showtimes ต่อหนัง: 2 รอบ คืนนี้และพรุ่งนี้
	showtimeSlots := []struct {
		hall  string
		start time.Duration
	}{
		{hall: "Hall A", start: 24 * time.Hour},
		{hall: "Hall B", start: 48 * time.Hour},
	}

	for _, movie := range movies {
		if err := movieRepo.Create(context.Background(), movie); err != nil {
			log.Fatalf("Failed to create movie %s: %v", movie.Title, err)
		}
		log.Printf("[Movie] Created: %s (ID: %s)", movie.Title, movie.ID.Hex())

		for _, slot := range showtimeSlots {
			duration := time.Duration(movie.Duration) * time.Minute
			startTime := time.Now().Add(slot.start)
			endTime := startTime.Add(duration)

			showtime := &model.Showtime{
				MovieID:   movie.ID,
				StartTime: startTime,
				EndTime:   endTime,
				Hall:      slot.hall,
				Seats:     generateSeats(40),
			}

			if err := showtimeRepo.Create(context.Background(), showtime); err != nil {
				log.Fatalf("Failed to create showtime: %v", err)
			}
			log.Printf("[Showtime] Created: %s %s at %s (ID: %s)",
				movie.Title, slot.hall, startTime.Format("2006-01-02 15:04"), showtime.ID.Hex())
		}
	}

	fmt.Println("")
	log.Println("Seed complete!")
}

// generateSeats สร้าง seats จำนวน count ที่นั่ง
// 40 seats = A1-A20, B1-B20
func generateSeats(count int) []model.Seat {
	seats := make([]model.Seat, 0, count)
	rows := "ABCDEFGHIJ"
	cols := 20

	for i := 0; i < count; i++ {
		row := string(rows[i/cols])
		col := (i % cols) + 1
		seats = append(seats, model.Seat{
			SeatNumber: fmt.Sprintf("%s%d", row, col),
			Status:     model.SeatAvailable,
		})
	}

	return seats
}
