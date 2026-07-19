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
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	db := client.Database(cfg.MongoDBName)

	movieRepo := repository.NewMovieRepository(db)
	showtimeRepo := repository.NewShowtimeRepository(db)

	// สร้างหนัง
	// movie := &model.Movie{
	// 	Title:       "Inception",
	// 	Description: "A mind-bending thriller",
	// 	Duration:    148,
	// }

	movies := []*model.Movie{
		{
			Title:       "Inception",
			Description: "A mind-bending thriller",
			Duration:    148,
			PosterURL:   "",
		},
		{
			Title:       "Interstellar",
			Description: "Space Exploration",
			Duration:    169,
			PosterURL:   "",
		},
	}

	for _, movie := range movies {
		if err := movieRepo.Create(context.Background(), movie); err != nil {
			log.Fatal(err)
		}

		// สร้างรอบฉาย
		showtime := &model.Showtime{
			MovieID:   movie.ID,
			StartTime: time.Now().Add(24 * time.Hour),
			EndTime:   time.Now().Add(27 * time.Hour),
			Hall:      "Hall A",
		}
		if err := showtimeRepo.Create(context.Background(), showtime); err != nil {
			log.Fatal(err)
		}

		// generate seats 40 ที่นั่ง
		for i := 0; i < 40; i++ {
			row := string("ABCDEFGHIJ"[i/20])
			col := (i % 20) + 1
			showtime.Seats = append(showtime.Seats, model.Seat{
				SeatNumber: fmt.Sprintf("%s%d", row, col),
				Status:     model.SeatAvailable,
			})
		}

		log.Printf("Created movie %s", movie.Title)
		log.Printf("Created showtime (ID: %s)", showtime.ID.Hex())
	}

	log.Println("🎬 Seed complete! Copy these IDs for testing.")
}