package domain

import "time"

type PortfolioItem struct {
	ID                    uint      `json:"id" gorm:"primaryKey"`
	Nome                  string    `json:"nome"`
	Qtd                   float64   `json:"qtd"`
	PrecoMedio            float64   `json:"preco_medio"`
	Proventos             float64   `json:"proventos"`
	PrecoMercado          float64   `json:"preco_mercado"`
	ResultadoComProventos float64   `json:"resultado_com_proventos"`
	SaldoBruto            float64   `json:"saldo_bruto"`
	UploadedAt            time.Time `json:"uploaded_at"`
}

type RecommendedItem struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Fundo        string    `json:"fundo"`
	Segmento     string    `json:"segmento"`
	PrecoAtual   float64   `json:"preco_atual"`
	PrecoMedio   float64   `json:"preco_medio"`
	PrecoTeto    float64   `json:"preco_teto"`
	Alocacao     float64   `json:"alocacao"`
	Recomendacao string    `json:"recomendacao"`
	UploadedAt   time.Time `json:"uploaded_at"`
}

type ContributionResult struct {
	Fundo         string  `json:"fundo"`
	Segmento      string  `json:"segmento"`
	PrecoAtual    float64 `json:"preco_atual"`
	AlocacaoAtual float64 `json:"alocacao_atual_pct"`
	AlocacaoAlvo  float64 `json:"alocacao_alvo_pct"`
	ValorAportar  float64 `json:"valor_a_aportar"`
	QtdComprar    int     `json:"qtd_a_comprar"`
	TotalGasto    float64 `json:"total_gasto"`
}

type ContributionResponse struct {
	ValorDisponivel float64              `json:"valor_disponivel"`
	ValorUtilizado  float64              `json:"valor_utilizado"`
	ValorSobra      float64              `json:"valor_sobra"`
	Fundos          []ContributionResult `json:"fundos"`
}
