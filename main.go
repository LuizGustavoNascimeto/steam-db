package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"steam-db/types"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func fetchGames(url string) ([]types.GameRes, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	var resData types.RawgResponse
	if err := json.NewDecoder(res.Body).Decode(&resData); err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return resData.Res, nil
}

func gamesResToGame(games []types.GameRes) []types.Game {
	var result []types.Game
	for _, g := range games {
		game := types.Game{
			ID:           g.ID,
			Name:         g.Name,
			Rating:       g.Rating,
			RatingsCount: g.RatingsCount,
			Metacritic:   g.Metacritic,
			Owned:        g.AddedByStatus.Owned,
			Beaten:       g.AddedByStatus.Beaten,
			Toplay:       g.AddedByStatus.Toplay,
			Dropped:      g.AddedByStatus.Dropped,
			Playing:      g.AddedByStatus.Playing,
			Yet:          g.AddedByStatus.Yet,
			Platforms:    []types.Platform{},
			Stores:       []types.Store{},
			Genres:       g.Genres,
			Tags:         g.Tags,
		}
		if g.Released != "" {
			game.Released = sql.NullString{String: g.Released, Valid: true}
		} else {
			game.Released = sql.NullString{Valid: false}
		}
		for _, p := range g.Platforms {
			platform := types.Platform{
				ID:   p.Platform.ID,
				Name: p.Platform.Name,
			}
			game.Platforms = append(game.Platforms, platform)
		}
		for _, s := range g.Stores {
			store := types.Store{
				ID:   s.Store.ID,
				Name: s.Store.Name,
			}
			game.Stores = append(game.Stores, store)
		}
		if game.RatingsCount > 100 {
			result = append(result, game)
		}
	}
	return result
}

func main() {
	dsn := "host=localhost user=postgres password=root dbname=rawg-db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err.Error())
	}

	db.AutoMigrate(&types.Game{}, &types.Platform{}, &types.Store{}, &types.Genre{}, &types.Tag{})
	// //fillGames(db)

	// ctx := context.Background()
	// //testGameDetailsFetch(db, ctx)
	// games, err := fetchGames("https://api.rawg.io/api/games?key=4b5662e9a72941998ad77125ffd533f2&page_size=1")
	// if err != nil {
	// 	fmt.Println("Error fetching games:", err)
	// 	return
	// }
	// gorm.G[types.Games](db).Create(ctx, &gamesResToGame(games)[0])
	// fmt.Println("Fetched games:", games)
	start := 1
	pages := (7157 / 40) + 1
	for i := start; i <= pages; i++ {
		url := fmt.Sprintf("https://api.rawg.io/api/games?key=4b5662e9a72941998ad77125ffd533f2&page_size=40&page=%d&metacritic=1,100", i)
		data, err := fetchGames(url)
		if err != nil {
			fmt.Println("Error fetching tags:", err)
			return
		}
		g := gamesResToGame(data)
		time.Sleep(200 * time.Millisecond)
		db.Model(&types.Game{}).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&g, 40)

		fmt.Printf("Fetched games page %d/%d\n", i, pages)
	}
	fmt.Println("Vc conseguiu!")
}
