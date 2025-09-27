// steamhtml/steamhtml.go
package steamhtml

import (
	"io"
	"net/url"
	"regexp"
	"steam-db/types"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Public API: ScrapeFromReader lê HTML (io.Reader) da página do jogo e tenta preencher GameDetailsRes.
func ScrapeFromReader(r io.Reader) (*types.GameDetailsRes, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	out := &types.GameDetailsRes{
		Price_overview: types.PriceOverview{Currency: "", Initial: 0},
		Platforms:      types.Plataform{},
		Metacritic:     types.Metacritic{},
		Recommended:    types.Recommended{Total: 0},
		Release_date:   types.ReleaseDate{ComingSoon: false, Date: ""},
	}

	// Nome do jogo
	out.Name = firstNonEmptyText(doc,
		"#appHubAppName",
		".apphub_AppName",
		".app-title h2",
		".page_title h2",
		"title") // fallback para <title>

	// Short description
	out.Short_description = firstNonEmptyText(doc,
		".game_description_snippet",
		"#game_area_description .game_description_snippet",
		".block_content .game_description_snippet")

	// Developers
	out.Developers = uniqueStringsFromSelection(doc, []string{
		".dev_row a",
		".developer_link a",
		"#developers_list a",
		".details_block a[href*='developer']",
	})

	// Type (single-player / software / demo / etc.) - tenta pegar do meta ou do tipo visível
	out.Type = strings.TrimSpace(firstNonEmptyText(doc,
		".product_type",
		".game_area_details_specs .type",
		".glance_tags a[href*='type']"))

	// Platforms — procura ícones / classes comuns
	out.Platforms.Windows = existsSelection(doc, ".platform_img.win, .platform_img.windows, .icon_win, .platform_img[src*='win']")
	out.Platforms.Mac = existsSelection(doc, ".platform_img.mac, .icon_mac, .platform_img[src*='mac']")
	out.Platforms.Linux = existsSelection(doc, ".platform_img.linux, .icon_linux, .platform_img[src*='linux']")

	// Release date
	release := firstNonEmptyText(doc,
		"#release_date .date",
		".release_date .date",
		".date",
		".release_date")
	out.Release_date.Date = strings.TrimSpace(release)
	out.Release_date.ComingSoon = strings.Contains(strings.ToLower(release), "coming") ||
		strings.Contains(strings.ToLower(release), "em breve") // pt-br

	// Recommendations / total (ex.: "123,456 people found this review helpful")
	recText := firstNonEmptyText(doc,
		"#recommendations .game_review_summary",
		"#recommendations .user_reviews_count",
		".user_reviews_count",
		".recommendations .count",
		".game_details .recommendations")
	out.Recommended.Total = extractInt(recText)

	// Metacritic
	metaScore := firstNonEmptyText(doc,
		"#game_area_metascore .score",
		".metacritic_score .score",
		".metacritic a .score")
	out.Metacritic.Score = extractInt(metaScore)
	if href, ok := doc.Find("#game_area_metascore a, .metacritic a").First().Attr("href"); ok {
		out.Metacritic.URL = href
	}

	// Categories / genres
	out.Categories = parseCategories(doc)
	out.Genres = parseGenres(doc)

	// Price: tenta várias seleções
	priceText := firstNonEmptyText(doc,
		".discount_final_price",
		".game_purchase_price",
		".price",
		".user_reviews_pricing .price",
		".game_area_purchase_price")
	priceCurrency, priceCents := parsePrice(priceText)
	out.Price_overview.Currency = priceCurrency
	out.Price_overview.Initial = priceCents

	// Free?
	out.Is_free = strings.Contains(strings.ToLower(priceText), "free") || strings.Contains(strings.ToLower(out.Short_description), "free to play") || priceText == ""

	// AppID: tenta extrair do URL em meta tags ou scripts
	out.Appid = extractAppIDFromPage(doc)

	return out, nil
}

//
// Helpers
//

// firstNonEmptyText tenta várias seleções e retorna o primeiro texto não-vazio.
func firstNonEmptyText(doc *goquery.Document, selectors ...string) string {
	for _, s := range selectors {
		sel := strings.TrimSpace(s)
		if sel == "" {
			continue
		}
		txt := strings.TrimSpace(doc.Find(sel).First().Text())
		if txt != "" {
			return txt
		}
	}
	// fallback: title tag
	if t := strings.TrimSpace(doc.Find("title").Text()); t != "" {
		return t
	}
	return ""
}

func existsSelection(doc *goquery.Document, selector string) bool {
	return doc.Find(selector).Length() > 0
}

// uniqueStringsFromSelection pesquisa múltiplos seletores e junta links/textos (ex.: developers)
func uniqueStringsFromSelection(doc *goquery.Document, selectors []string) []string {
	set := map[string]struct{}{}
	out := []string{}
	for _, s := range selectors {
		doc.Find(s).Each(func(i int, sel *goquery.Selection) {
			text := strings.TrimSpace(sel.Text())
			if text == "" {
				// tenta atributo title ou alt
				if alt, ok := sel.Attr("title"); ok {
					text = strings.TrimSpace(alt)
				} else if alt, ok := sel.Attr("alt"); ok {
					text = strings.TrimSpace(alt)
				}
			}
			if text != "" {
				if _, found := set[text]; !found {
					set[text] = struct{}{}
					out = append(out, text)
				}
			}
		})
	}
	return out
}

// extractInt extrai o primeiro número inteiro de uma string (ex.: "123,456 users" -> 123456)
func extractInt(s string) int {
	if s == "" {
		return 0
	}
	// remove pontos, espaços e palavras; pega apenas dígitos e vírgulas
	re := regexp.MustCompile(`[\d\.,]+`)
	m := re.FindString(s)
	if m == "" {
		return 0
	}
	// normaliza: remove '.' como separador de milhares, ',' -> ''
	m = strings.ReplaceAll(m, ".", "")
	m = strings.ReplaceAll(m, ",", "")
	val, err := strconv.Atoi(m)
	if err != nil {
		return 0
	}
	return val
}

// parsePrice tenta extrair moeda e valor em centavos (ou menor unidade inteira)
// Ex.: "$19.99" -> ("USD", 1999) se não souber a moeda, retorna símbolo como Currency.
func parsePrice(raw string) (currency string, cents int) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", 0
	}
	// tenta extrair código de moeda em atributos como data-currency
	// Mas aqui, como não temos DOM do atributo facilmente, vamos inferir da string
	// Remove palavras como "Buy", "Free", "R$" etc.
	// Primeiro, procura símbolo ou código ISO
	// exemplos: "R$ 19,99", "€9.99", "$19.99", "19,99 R$", "USD 19.99"
	isoRe := regexp.MustCompile(`([A-Z]{3})\s*([0-9\.,]+)`)
	if m := isoRe.FindStringSubmatch(raw); len(m) >= 3 {
		currency = m[1]
		num := m[2]
		return currency, parseNumberToCents(num)
	}

	// símbolo seguido por número
	symNumRe := regexp.MustCompile(`([^\d\.,\s]+)\s*([0-9\.,]+)`)
	if m := symNumRe.FindStringSubmatch(raw); len(m) >= 3 {
		currency = m[1]
		num := m[2]
		return currency, parseNumberToCents(num)
	}

	// número seguido por símbolo (ex: "19,99 R$")
	numSymRe := regexp.MustCompile(`([0-9\.,]+)\s*([^\d\.,\s]+)`)
	if m := numSymRe.FindStringSubmatch(raw); len(m) >= 3 {
		num := m[1]
		currency = m[2]
		return currency, parseNumberToCents(num)
	}

	// apenas número
	numRe := regexp.MustCompile(`([0-9\.,]+)`)
	if m := numRe.FindStringSubmatch(raw); len(m) >= 2 {
		return "", parseNumberToCents(m[1])
	}

	return "", 0
}

// parseNumberToCents transforma "19.99" ou "19,99" ou "1.234,56" em centavos (1999 / 123456)
func parseNumberToCents(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	// Se tem vírgula e ponto: ex 1.234,56 -> remover pontos, trocar vírgula por ponto
	if strings.Contains(s, ",") && strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	}
	// Se apenas vírgula e sem ponto: trocar vírgula por ponto
	if strings.Contains(s, ",") && !strings.Contains(s, ".") {
		s = strings.ReplaceAll(s, ",", ".")
	}
	// Agora parse float
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		// fallback: extrair dígitos
		re := regexp.MustCompile(`[\d]+`)
		m := re.FindAllString(s, -1)
		if len(m) == 0 {
			return 0
		}
		joined := strings.Join(m, "")
		val, _ := strconv.Atoi(joined)
		return val
	}
	// retorna em centavos (inteiro)
	return int(f * 100)
}

// parseCategories tenta extrair categorias da página (glance_tags ou details_block)
func parseCategories(doc *goquery.Document) []types.CategoryRes {
	out := []types.CategoryRes{}
	// tags (simples)
	doc.Find(".glance_tags a.app_tag").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			out = append(out, types.CategoryRes{ID: 0, Description: text})
		}
	})
	// detalhes (links que contenham /category/ ou /ccat/)
	doc.Find(".details_block a[href]").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			if strings.Contains(href, "/category/") || strings.Contains(href, "/ccat/") {
				desc := strings.TrimSpace(s.Text())
				if desc != "" {
					out = append(out, types.CategoryRes{ID: 0, Description: desc})
				}
			}
		}
	})
	return uniqueCategories(out)
}

func uniqueCategories(in []types.CategoryRes) []types.CategoryRes {
	seen := map[string]struct{}{}
	out := []types.CategoryRes{}
	for _, c := range in {
		k := strings.ToLower(strings.TrimSpace(c.Description))
		if k == "" {
			continue
		}
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			out = append(out, c)
		}
	}
	return out
}

// parseGenres (semelhante)
func parseGenres(doc *goquery.Document) []types.GenreRes {
	out := []types.GenreRes{}
	// itemprop genre
	doc.Find("a[itemprop='genre']").Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		href, _ := s.Attr("href")
		if text != "" {
			out = append(out, types.GenreRes{ID: href, Description: text})
		}
	})
	// detalhes: links contendo /genre/
	doc.Find(".details_block a[href]").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			if strings.Contains(href, "/genre/") {
				desc := strings.TrimSpace(s.Text())
				out = append(out, types.GenreRes{ID: href, Description: desc})
			}
		}
	})
	return uniqueGenres(out)
}

func uniqueGenres(in []types.GenreRes) []types.GenreRes {
	seen := map[string]struct{}{}
	out := []types.GenreRes{}
	for _, g := range in {
		k := strings.ToLower(strings.TrimSpace(g.Description))
		if k == "" {
			continue
		}
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			out = append(out, g)
		}
	}
	return out
}

// extractAppIDFromPage tenta pegar o appid de scripts, metas ou do canonical URL
func extractAppIDFromPage(doc *goquery.Document) int {
	// 1) tentar meta tags (ex: og:url)
	if og, ok := doc.Find("meta[property='og:url']").Attr("content"); ok {
		if id := extractAppIDFromURL(og); id != 0 {
			return id
		}
	}
	// 2) canonical link
	if href, ok := doc.Find("link[rel='canonical']").Attr("href"); ok {
		if id := extractAppIDFromURL(href); id != 0 {
			return id
		}
	}
	// 3) procurar em scripts por "appid" ou "data-ds-appid"
	var appid int
	doc.Find("script").EachWithBreak(func(i int, s *goquery.Selection) bool {
		txt := s.Text()
		re := regexp.MustCompile(`app\s*id["'\s:=]+([0-9]{1,8})`)
		if m := re.FindStringSubmatch(txt); len(m) >= 2 {
			appid, _ = strconv.Atoi(m[1])
			return false
		}
		// data-ds-appid="12345"
		re2 := regexp.MustCompile(`data-ds-appid\s*=\s*["']?([0-9]{1,8})["']?`)
		if m := re2.FindStringSubmatch(txt); len(m) >= 2 {
			appid, _ = strconv.Atoi(m[1])
			return false
		}
		return true
	})
	if appid != 0 {
		return appid
	}
	// 4) try to find numeric in title like "Game Name on Steam (12345)" - less reliable
	title := doc.Find("title").Text()
	reTitle := regexp.MustCompile(`\b([0-9]{4,8})\b`)
	if m := reTitle.FindStringSubmatch(title); len(m) >= 2 {
		id, _ := strconv.Atoi(m[1])
		return id
	}
	return 0
}

func extractAppIDFromURL(u string) int {
	if u == "" {
		return 0
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return 0
	}
	// Steam store URL pattern: /app/<appid>/...
	parts := strings.Split(parsed.Path, "/")
	for i, p := range parts {
		if p == "app" && i+1 < len(parts) {
			idStr := parts[i+1]
			id, err := strconv.Atoi(idStr)
			if err == nil {
				return id
			}
		}
	}
	// fallback: last numeric segment
	for i := len(parts) - 1; i >= 0; i-- {
		if n, err := strconv.Atoi(parts[i]); err == nil {
			return n
		}
	}
	return 0
}
