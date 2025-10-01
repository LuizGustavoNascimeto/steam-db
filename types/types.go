package types

import "database/sql"

type RawgResponse struct {
	Count int       `json:"count"`
	Next  string    `json:"next"`
	Prev  string    `json:"previous"`
	Res   []GameRes `json:"results"`
}
type RawgResponse2 struct {
	Count int     `json:"count"`
	Next  string  `json:"next"`
	Prev  string  `json:"previous"`
	Res   []Store `json:"results"`
}
type Game struct {
	ID           int            `json:"id" gorm:"primaryKey;column:id"`
	Name         string         `json:"name" gorm:"column:name"`
	Released     sql.NullString `json:"released" gorm:"column:released;type:date"`
	Rating       float64        `json:"rating" gorm:"column:rating"`
	RatingsCount int            `json:"ratings_count" gorm:"column:ratings_count"`
	Metacritic   int            `json:"metacritic" gorm:"column:metacritic;type:int"`
	Owned        int            `json:"owned" gorm:"column:owned;type:int"`
	Beaten       int            `json:"beaten" gorm:"column:beaten;type:int"`
	Toplay       int            `json:"toplay" gorm:"column:toplay;type:int"`
	Dropped      int            `json:"dropped" gorm:"column:dropped;type:int"`
	Playing      int            `json:"playing" gorm:"column:playing;type:int"`
	Yet          int            `json:"yet" gorm:"column:yet;type:int"`
	Platforms    []Platform     `json:"platforms" gorm:"many2many:game_platforms;"`
	Stores       []Store        `json:"stores" gorm:"many2many:game_stores;"`
	Genres       []Genre        `json:"genres" gorm:"many2many:game_genres;"`
	Tags         []Tag          `json:"tags" gorm:"many2many:game_tags;"`
}
type Tag struct {
	ID   int    `json:"id" gorm:"primaryKey;column:id"`
	Name string `json:"name" gorm:"uniqueIndex;column:name"`
}
type Genre struct {
	ID   int    `json:"id" gorm:"primaryKey;column:id"`
	Name string `json:"name" gorm:"uniqueIndex;column:name"`
}
type Platform struct {
	ID   int    `json:"id" gorm:"primaryKey;column:id"`
	Name string `json:"name" gorm:"uniqueIndex;column:name"`
}
type Store struct {
	ID   int    `json:"id" gorm:"primaryKey;column:id"`
	Name string `json:"name" gorm:"uniqueIndex;column:name"`
}
type Developer struct {
	ID   int    `json:"id" gorm:"primaryKey;column:id"`
	Name string `json:"name" gorm:"uniqueIndex;column:name"`
}
type GameRes struct {
	ID            int            `json:"id"`
	Name          string         `json:"name"`
	Released      string         `json:"released"`
	Rating        float64        `json:"rating"`
	RatingsCount  int            `json:"ratings_count"`
	Metacritic    int            `json:"metacritic"`
	AddedByStatus AddStatus      `json:"added_by_status"`
	Platforms     []PlatformsRes `json:"platforms"`
	Stores        []StoreRes     `json:"stores"`
	Genres        []Genre        `json:"genres"`
	Tags          []Tag          `json:"tags"`
}
type AddStatus struct {
	Owned   int `json:"owned"`
	Beaten  int `json:"beaten"`
	Toplay  int `json:"toplay"`
	Dropped int `json:"dropped"`
	Playing int `json:"playing"`
	Yet     int `json:"yet"`
}
type PlatformsRes struct {
	Platform Platform `json:"platform"`
}
type StoreRes struct {
	Store Store `json:"store"`
}
