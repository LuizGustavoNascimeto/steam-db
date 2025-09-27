package dateFormater

import (
	"fmt"
	"strings"
)

func ParseReleaseDate(dateStr string) (string, error) {
	var err error = nil
	var day, month, year string

	// Meses nos 2 formatos possíveis (com ponto e com vírgula)
	months := map[string]string{
		"jan.": "01", "feb.": "02", "mar.": "03", "apr.": "04",
		"may.": "05", "jun.": "06", "jul.": "07", "aug.": "08",
		"sep.": "09", "oct.": "10", "nov.": "11", "dec.": "12",

		"Jan,": "01", "Feb,": "02", "Mar,": "03", "Apr,": "04",
		"May,": "05", "Jun,": "06", "Jul,": "07", "Aug,": "08",
		"Sep,": "09", "Oct,": "10", "Nov,": "11", "Dec,": "12",

		"Jan": "01", "Feb": "02", "Mar": "03", "Apr": "04",
		"May": "05", "Jun": "06", "Jul": "07", "Aug": "08",
		"Sep": "09", "Oct": "10", "Nov": "11", "Dec": "12",
	}

	// Caso 1: formato com barras → "1/nov./2000"
	if strings.Contains(dateStr, "/") {
		parts := strings.Split(dateStr, "/")
		if len(parts) != 3 {
			return "", fmt.Errorf("formato inválido: %s", dateStr)
		}
		day = parts[0]
		month = months[strings.ToLower(parts[1])]
		year = parts[2]

	} else if strings.Contains(dateStr, ",") {
		// Caso 2 e 3: tem vírgula → "Jul 9, 2013" ou "9 May, 2004"
		parts := strings.Fields(dateStr) // split por espaço
		if len(parts) != 3 {
			return "", fmt.Errorf("formato inválido: %s", dateStr)
		}

		if _, ok := months[parts[0]]; ok {
			// "Jul 9, 2013"
			month = months[parts[0]]
			day = strings.TrimSuffix(parts[1], ",")
			year = parts[2]
		} else {
			// "9 May, 2004"
			day = parts[0]
			month = months[parts[1]]
			year = strings.TrimSuffix(parts[2], ",")
		}
	} else {
		return "", fmt.Errorf("não reconheço o formato: %s", dateStr)
	}

	// Normalizar dia com zero à esquerda
	if len(day) == 1 {
		day = "0" + day
	}

	normalized := fmt.Sprintf("%s-%s-%s", year, month, day)
	return normalized, err
}
