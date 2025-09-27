package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"steam-db/dateFormater"
	steamhtml "steam-db/scrapper"
	"steam-db/types"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func PaginateFrom[T any](db *gorm.DB, ctx context.Context, offset int, pageSize int) ([]T, error) {
	var results []T
	if err := db.
		Offset(offset). // começa após o último elemento
		Limit(pageSize).
		Find(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func fetchGameDetails(appID int, cc string) (types.GameDetailsRes, error) {
	url := fmt.Sprintf("https://store.steampowered.com/api/appdetails?appids=%d&cc=%s&l=en", appID, cc)

	res, err := http.Get(url)
	if err != nil {
		return types.GameDetailsRes{}, fmt.Errorf("error making http request: %w", err)
	}
	defer res.Body.Close()

	resData, err := io.ReadAll(res.Body)
	if err != nil {
		return types.GameDetailsRes{}, fmt.Errorf("error reading response body: %w", err)
	}

	resp := make(map[string]types.SteamResponse)
	if err := json.Unmarshal(resData, &resp); err != nil {
		return types.GameDetailsRes{}, fmt.Errorf("error unmarshalling: %w", err)
	}

	stringID := fmt.Sprintf("%d", appID)
	return resp[stringID].Data, nil
}

func fillGames(db *gorm.DB) {
	var url string = "https://api.steampowered.com/ISteamApps/GetAppList/v2/"
	res, err := http.Get(url)
	if err != nil {
		fmt.Printf("error making http request: %s\n", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	resData, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("error reading response body: %s\n", err)
		os.Exit(1)
	}
	var appList struct {
		Applist struct {
			Apps []types.Game `json:"apps"`
		} `json:"applist"`
	}
	if err := json.Unmarshal(resData, &appList); err != nil {
		log.Fatal(err)
	}
	db.CreateInBatches(appList.Applist.Apps, 1000)
}
func testGameDetailsFetch(db *gorm.DB, ctx context.Context) (types.GameDetailsRes, error) {
	url := fmt.Sprintf("https://store.steampowered.com/app/%d/", 3720)
	resp, err := http.Get(url)
	if err != nil {
		return types.GameDetailsRes{}, err
	}
	defer resp.Body.Close()

	// resp.Body já é um io.Reader
	html, err := io.ReadAll(resp.Body) // converte para []byte
	if err != nil {
		return types.GameDetailsRes{}, err
	}
	// game, err := fetchGameDetails(3720, "uk")
	gameRef, err := steamhtml.ScrapeFromReader(bytes.NewReader(html))
	game := *gameRef
	if err != nil {
		return types.GameDetailsRes{}, err
	}
	if err != nil {
		return types.GameDetailsRes{}, err
	}
	releaseDate, err := dateFormater.ParseReleaseDate(game.Release_date.Date)
	if err != nil {
		return types.GameDetailsRes{}, err
	}
	fmt.Printf("Fetched game details: %+v\n", game.Appid)
	gorm.G[types.GameDetails](db).Create(ctx, &types.GameDetails{
		Appid:             game.Appid,
		Name:              game.Name,
		Type:              game.Type,
		Is_free:           game.Is_free,
		Short_description: game.Short_description,
		Developers:        fmt.Sprintf("%v", game.Developers),
		Currency:          game.Price_overview.Currency,
		Price:             game.Price_overview.Initial,
		Mac:               game.Platforms.Mac,
		Windows:           game.Platforms.Windows,
		Linux:             game.Platforms.Linux,
		Metacritic:        game.Metacritic.Score,
		Recommendations:   game.Recommended.Total,
		Release_at:        releaseDate,
	})
	return game, nil
}

func main() {
	dsn := "host=localhost user=postgres password=root dbname=steam-db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database", err.Error())
	}

	db.AutoMigrate(&types.Game{}, &types.GameDetails{}, &types.Genre{}, &types.Category{}, &types.GameHasGenre{}, &types.GameHasCategory{})
	//fillGames(db)

	ctx := context.Background()
	//testGameDetailsFetch(db, ctx)

	genres, err := gorm.G[types.Genre](db).Find(ctx)
	if err != nil {
		log.Fatal(err)
	}
	genreMap := make(map[int]types.Genre)
	for _, g := range genres {
		genreMap[g.ID] = g
	}
	categories, err := gorm.G[types.Category](db).Find(ctx)
	if err != nil {
		log.Fatal(err)
	}
	categoryMap := make(map[int]types.Category)
	for _, c := range categories {
		categoryMap[c.ID] = c
	}
	var count int64
	db.Model(&types.Game{}).Count(&count)
	bufferSize := 200
	total_pages := int(count)/bufferSize + 1
	var start int = 0
	time.Sleep(5 * time.Minute)
	// Paginação simples
	for i := 1; i <= total_pages; i++ {
		games, err := PaginateFrom[types.Game](db, ctx, start, bufferSize)
		if err != nil {
			log.Fatal(err)
		}
		var gameBuffer []types.GameDetails
		var game_has_categoryBuffer []types.GameHasCategory
		var game_has_genreBuffer []types.GameHasGenre
		for _, game := range games {
			// POR ALGUM CARALHO O UK É MELHOR
			gameDetailsRes, err := fetchGameDetails(game.Appid, "br")
			time.Sleep(100 * time.Millisecond)

			if err != nil || gameDetailsRes.Appid == 0 {
				fmt.Printf("BR falhou: %d - ", game.Appid)

				// Tenta no servidor US
				gameDetailsRes, err = fetchGameDetails(game.Appid, "uk")
				time.Sleep(100 * time.Millisecond)

				if err != nil || gameDetailsRes.Appid == 0 {
					fmt.Printf("UK falhou: %s\n", game.Name)

					continue
				}
			}
			if gameDetailsRes.Appid != game.Appid {
				fmt.Printf("ID MISMATCH %d:%d\n", game.Appid, gameDetailsRes.Appid)
				continue
			}
			if gameDetailsRes.Release_date.ComingSoon {
				fmt.Printf("skipping %d, coming soon\n", game.Appid)
				db.Delete(&types.Game{}, "appid = ?", game.Appid)
				continue
			}
			if gameDetailsRes.Type != "game" {
				db.Delete(&types.Game{}, "appid = ?", game.Appid)
				fmt.Printf("skipping %d %s\n", game.Appid, gameDetailsRes.Type)
				continue
			}
			fmt.Println("Processing", game.Appid, gameDetailsRes.Name)

			releaseDate, err := dateFormater.ParseReleaseDate(gameDetailsRes.Release_date.Date)
			if err != nil {
				fmt.Printf("error parsing date: %s\n", err)
			}
			gameBuffer = append(gameBuffer, types.GameDetails{
				Appid:             gameDetailsRes.Appid,
				Name:              gameDetailsRes.Name,
				Type:              gameDetailsRes.Type,
				Is_free:           gameDetailsRes.Is_free,
				Short_description: gameDetailsRes.Short_description,
				Developers:        fmt.Sprintf("%v", gameDetailsRes.Developers),
				Currency:          gameDetailsRes.Price_overview.Currency,
				Price:             gameDetailsRes.Price_overview.Initial,
				Mac:               gameDetailsRes.Platforms.Mac,
				Windows:           gameDetailsRes.Platforms.Windows,
				Linux:             gameDetailsRes.Platforms.Linux,
				Metacritic:        gameDetailsRes.Metacritic.Score,
				Recommendations:   gameDetailsRes.Recommended.Total,
				Release_at:        releaseDate,
			})

			for _, category := range gameDetailsRes.Categories {
				if _, ok := categoryMap[category.ID]; !ok {
					db.Create(&types.Category{
						ID:          category.ID,
						Description: category.Description,
					})
					categoryMap[category.ID] = types.Category(category)
				}
				game_has_categoryBuffer = append(game_has_categoryBuffer, types.GameHasCategory{
					AppID:      gameDetailsRes.Appid,
					CategoryID: category.ID,
				})
			}
			for _, genre := range gameDetailsRes.Genres {
				genreID, err := strconv.Atoi(genre.ID)
				if err != nil {
					fmt.Printf("error converting genre ID: %s\n", err)
					continue
				}
				if _, ok := genreMap[genreID]; !ok {
					if err := db.Create(&types.Genre{
						ID:          genreID,
						Description: genre.Description,
					}).Error; err != nil {
						fmt.Printf("error inserting genre ID %d: %s\n", genreID, err)
						// decida: continuar ou abortar. Aqui eu continuo.
						continue
					}
					genreMap[genreID] = types.Genre{ID: genreID, Description: genre.Description}
				}
				game_has_genreBuffer = append(game_has_genreBuffer, types.GameHasGenre{
					AppID:   gameDetailsRes.Appid,
					GenreID: genreID,
				})
			}
		}
		db.CreateInBatches(gameBuffer, bufferSize)
		fmt.Println("Inserted batch", i, "of", total_pages)
		fmt.Println("FINAL BUFFER", len(gameBuffer))
		time.Sleep(5 * time.Minute)
		db.CreateInBatches(game_has_categoryBuffer, bufferSize)
		db.CreateInBatches(game_has_genreBuffer, bufferSize)
		start += bufferSize
	}
}
