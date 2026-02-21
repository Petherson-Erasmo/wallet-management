package csvparser

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
)

func parseFloat(s string) (float64, error) {
	s = strings.TrimSpace(s)
	// Remove prefixo monetário "R$" (com ou sem espaço)
	s = strings.ReplaceAll(s, "R$", "")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "%", "")
	s = strings.TrimSpace(s)
	if s == "" || s == "-" {
		return 0, nil
	}

	dotCount := strings.Count(s, ".")
	commaCount := strings.Count(s, ",")

	switch {
	case commaCount > 0 && dotCount > 0:
		// Formato brasileiro: 1.234,56 → ponto é milhar, vírgula é decimal
		s = strings.ReplaceAll(s, ".", "")
		s = strings.ReplaceAll(s, ",", ".")
	case commaCount > 0 && dotCount == 0:
		// Só vírgula: 1234,56 → vírgula é decimal
		s = strings.ReplaceAll(s, ",", ".")
	// dotCount > 0 && commaCount == 0: formato padrão (1234.56), não altera
	}

	return strconv.ParseFloat(s, 64)
}

// removeAccents normaliza caracteres acentuados para ASCII
func removeAccents(s string) string {
	replacer := strings.NewReplacer(
		"Á", "A", "À", "A", "Â", "A", "Ã", "A",
		"É", "E", "È", "E", "Ê", "E",
		"Í", "I", "Ì", "I", "Î", "I",
		"Ó", "O", "Ò", "O", "Ô", "O", "Õ", "O",
		"Ú", "U", "Ù", "U", "Û", "U",
		"Ç", "C",
	)
	return replacer.Replace(s)
}

func normalizeHeader(h string) string {
	s := strings.TrimSpace(h)
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ToUpper(s)
	s = removeAccents(s)
	return s
}

// extractTicker extrai o código do ativo (ex: "ALZR11FII ALIANZA..." → "ALZR11")
func extractTicker(name string) string {
	// Pega a primeira palavra
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return name
	}
	ticker := parts[0]
	// Remove sufixos FII/FIAGRO do ticker
	for _, suffix := range []string{"FIAGRO", "FII"} {
		if strings.HasSuffix(strings.ToUpper(ticker), suffix) {
			ticker = ticker[:len(ticker)-len(suffix)]
			break
		}
	}
	return strings.ToUpper(ticker)
}

func ParsePortfolioCSV(r io.Reader) ([]domain.PortfolioItem, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("falha ao ler cabecalho: %w", err)
	}

	idx := make(map[string]int)
	for i, h := range headers {
		idx[normalizeHeader(h)] = i
	}

	// Aceita "PRODUTO" como alias de "NOME DO ATIVO"
	if _, ok := idx["NOME DO ATIVO"]; !ok {
		if i, ok := idx["PRODUTO"]; ok {
			idx["NOME DO ATIVO"] = i
		}
	}

	required := []string{"NOME DO ATIVO", "QTD"}
	for _, col := range required {
		if _, ok := idx[col]; !ok {
			return nil, fmt.Errorf("coluna obrigatoria nao encontrada: %s", col)
		}
	}

	now := time.Now()
	var items []domain.PortfolioItem

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("falha ao ler linha: %w", err)
		}

		if strings.TrimSpace(record[idx["NOME DO ATIVO"]]) == "" {
			continue
		}

		item := domain.PortfolioItem{UploadedAt: now}
		// Extrai apenas o ticker (ex: "ALZR11FII ALIANZA..." → "ALZR11")
		item.Nome = extractTicker(record[idx["NOME DO ATIVO"]])

		getFloat := func(col string) float64 {
			if i, ok := idx[col]; ok && i < len(record) {
				v, _ := parseFloat(record[i])
				return v
			}
			return 0
		}

		item.Qtd = getFloat("QTD")
		item.PrecoMedio = getFloat("PRECO MEDIO")
		item.Proventos = getFloat("PROVENTOS")
		item.PrecoMercado = getFloat("PRECO DE MERCADO")
		item.ResultadoComProventos = getFloat("RESULTADO C/ PROVENTOS")
		item.SaldoBruto = getFloat("SALDO BRUTO")

		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("nenhum ativo encontrado no CSV")
	}

	return items, nil
}

func ParseRecommendationCSV(r io.Reader) ([]domain.RecommendedItem, error) {
	reader := csv.NewReader(r)
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("falha ao ler cabecalho: %w", err)
	}

	idx := make(map[string]int)
	for i, h := range headers {
		idx[normalizeHeader(h)] = i
	}

	required := []string{"FUNDO", "SEGMENTO", "PRECO ATUAL", "ALOCACAO", "RECOMENDACAO"}
	for _, col := range required {
		if _, ok := idx[col]; !ok {
			return nil, fmt.Errorf("coluna obrigatoria nao encontrada: %s", col)
		}
	}

	now := time.Now()
	var items []domain.RecommendedItem

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("falha ao ler linha: %w", err)
		}

		if strings.TrimSpace(record[idx["FUNDO"]]) == "" {
			continue
		}

		item := domain.RecommendedItem{UploadedAt: now}
		item.Fundo = strings.TrimSpace(record[idx["FUNDO"]])
		item.Segmento = strings.TrimSpace(record[idx["SEGMENTO"]])
		item.Recomendacao = strings.ToUpper(strings.TrimSpace(record[idx["RECOMENDACAO"]]))

		getFloat := func(col string) float64 {
			if i, ok := idx[col]; ok && i < len(record) {
				v, _ := parseFloat(record[i])
				return v
			}
			return 0
		}

		item.PrecoAtual = getFloat("PRECO ATUAL")
		item.PrecoMedio = getFloat("PRECO MEDIO")
		item.PrecoTeto = getFloat("PRECO TETO")
		item.Alocacao = getFloat("ALOCACAO")

		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, fmt.Errorf("nenhum fundo encontrado no CSV")
	}

	return items, nil
}
