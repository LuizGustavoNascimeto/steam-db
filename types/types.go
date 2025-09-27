package types

type SteamResponse struct {
	Success bool           `json:"success"`
	Data    GameDetailsRes `json:"data"`
}

type Game struct {
	Appid int    `json:"appid" gorm:"primaryKey"`
	Name  string `json:"name"`
}
type GameDetails struct {
	Appid             int        `json:"steam_appid" gorm:"primaryKey"`
	Name              string     `json:"name" gorm:"column:name"`
	Type              string     `json:"type" gorm:"column:type"`
	Is_free           bool       `json:"is_free" gorm:"column:is_free"`
	Short_description string     `json:"short_description" gorm:"column:short_description"`
	Developers        string     `json:"developers" gorm:"column:developers"`
	Currency          string     `json:"currency" gorm:"column:currency"`
	Price             int        `json:"price" gorm:"column:price"`
	Mac               bool       `json:"mac" gorm:"column:mac"`
	Windows           bool       `json:"windows" gorm:"column:windows"`
	Linux             bool       `json:"linux" gorm:"column:linux"`
	Metacritic        int        `json:"metacritic" gorm:"column:metacritic"`
	Recommendations   int        `json:"recommendations" gorm:"column:recommendations"`
	Release_at        string     `json:"release_date" gorm:"column:release_at;type:date"`
	Genres            []Genre    `gorm:"many2many:game_has_genres;joinForeignKey:AppID;joinReferences:GenreID"`
	Categories        []Category `gorm:"many2many:game_has_categories;joinForeignKey:AppID;joinReferences:CategoryID"`
}
type GameDetailsRes struct {
	Appid             int           `json:"steam_appid" gorm:"primaryKey"`
	Name              string        `json:"name"`
	Type              string        `json:"type"`
	Is_free           bool          `json:"is_free"`
	Short_description string        `json:"short_description"`
	Developers        []string      `json:"developers"`
	Price_overview    PriceOverview `json:"price_overview"`
	Platforms         Plataform     `json:"platforms"`
	Metacritic        Metacritic    `json:"metacritic"`
	Categories        []CategoryRes `json:"categories" gorm:"-"`
	Genres            []GenreRes    `json:"genres" gorm:"-"`
	Recommended       Recommended   `json:"recommendations"`
	Release_date      ReleaseDate   `json:"release_date"`
}

type Category struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	Description string `json:"description"`
}
type Genre struct {
	ID          int    `json:"id" gorm:"primaryKey"`
	Description string `json:"description"`
}
type GameHasGenre struct {
	AppID   int `gorm:"column:app_id;primaryKey"`
	GenreID int `gorm:"column:genre_id;primaryKey"`

	Game  GameDetails `gorm:"foreignKey:AppID;references:Appid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Genre Genre       `gorm:"foreignKey:GenreID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type GameHasCategory struct {
	AppID      int `gorm:"column:app_id;primaryKey"`
	CategoryID int `gorm:"column:category_id;primaryKey"`

	Game     GameDetails `gorm:"foreignKey:AppID;references:Appid;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Category Category    `gorm:"foreignKey:CategoryID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

type PriceOverview struct {
	Currency string `json:"currency"`
	Initial  int    `json:"initial"`
}

type Plataform struct {
	Windows bool `json:"windows"`
	Mac     bool `json:"mac"`
	Linux   bool `json:"linux"`
}

type Metacritic struct {
	Score int    `json:"score"`
	URL   string `json:"url"`
}
type CategoryRes struct {
	ID          int    `json:"id"`
	Description string `json:"description"`
}
type GenreRes struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}
type Recommended struct {
	Total int `json:"total"`
}
type ReleaseDate struct {
	ComingSoon bool   `json:"coming_soon"`
	Date       string `json:"date"`
}
