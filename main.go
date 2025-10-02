package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"steam-db/types"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type RawgGameDetails struct {
	ID    int `json:"id"`
	Added int `json:"added"`
}

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

func fetchGameDetails(rawgID int) (RawgGameDetails, error) {
	url := fmt.Sprintf("https://api.rawg.io/api/games/%d?key=4b5662e9a72941998ad77125ffd533f2", rawgID)

	resp, err := http.Get(url)
	if err != nil {
		return RawgGameDetails{}, err
	}
	defer resp.Body.Close()

	var details RawgGameDetails
	if err := json.NewDecoder(resp.Body).Decode(&details); err != nil {
		return RawgGameDetails{}, err
	}

	return details, nil
}

func main() {
	dsn := "host=localhost user=postgres password=root dbname=rawg-db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err.Error())
	}
	db.AutoMigrate(&types.Game{})

	// Buscar todos os jogos salvos
	var games []types.Game
	if err := db.Find(&games).Error; err != nil {
		log.Fatal("failed to fetch games:", err)
	}

	for _, g := range games {
		// assumindo que o campo `RawgID` existe na struct types.Game
		details, err := fetchGameDetails(g.ID)
		if err != nil {
			fmt.Println("erro na request:", err)
			continue
		}

		// atualizar o campo added
		if err := db.Model(&types.Game{}).
			Where("id = ?", g.ID).
			Updates(map[string]interface{}{"added": details.Added}).
			Error; err != nil {
			fmt.Println("erro no update:", err)
		} else {
			fmt.Printf("Atualizado jogo %d -> added=%d\n", g.ID, details.Added)
		}

		time.Sleep(200 * time.Millisecond) // respeitar rate limit da RAWG
	}
}
