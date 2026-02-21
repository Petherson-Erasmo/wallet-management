package service

import (
	"fmt"
	"math"
	"strings"

	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
	"github.com/Petherson-Erasmo/wallet-management/internal/repository"
)

const maxAllocacaoPorAtivo = 10.0

type ContributionService struct {
	portfolioRepo      *repository.PortfolioRepository
	recommendationRepo *repository.RecommendationRepository
}

func NewContributionService(
	portfolioRepo *repository.PortfolioRepository,
	recommendationRepo *repository.RecommendationRepository,
) *ContributionService {
	return &ContributionService{
		portfolioRepo:      portfolioRepo,
		recommendationRepo: recommendationRepo,
	}
}

// Calculate calcula quanto aportar em cada fundo dado um valor disponível.
//
// Regras:
//  1. Apenas fundos presentes na carteira atual E com RECOMENDACAO = COMPRAR são elegíveis
//  2. O valor é distribuído proporcionalmente às alocações alvo da carteira recomendada
//  3. Nenhum fundo pode ultrapassar 10% do total da carteira após o aporte
//  4. Se um fundo atingir o teto de 10%, o excedente é redistribuído aos demais
func (s *ContributionService) Calculate(valor float64) (*domain.ContributionResponse, error) {
	if valor <= 0 {
		return nil, fmt.Errorf("o valor do aporte deve ser maior que zero")
	}

	portfolio, err := s.portfolioRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar carteira: %w", err)
	}

	recommended, err := s.recommendationRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar carteira recomendada: %w", err)
	}

	if len(recommended) == 0 {
		return nil, fmt.Errorf("carteira recomendada nao encontrada; importe o CSV de recomendacao primeiro")
	}

	// Mapeia carteira atual pelo ticker
	portfolioMap := make(map[string]domain.PortfolioItem)
	var totalAtual float64
	for _, item := range portfolio {
		key := normalizeKey(item.Nome)
		portfolioMap[key] = item
		totalAtual += item.SaldoBruto
	}

	totalAlvo := totalAtual + valor

	type elegivel struct {
		rec        domain.RecommendedItem
		saldoAtual float64
		alocAlvo   float64 // % alvo capped em 10
	}

	// 1. Filtra fundos elegíveis: em carteira + COMPRAR
	var elegiveis []elegivel
	for _, rec := range recommended {
		if strings.ToUpper(strings.TrimSpace(rec.Recomendacao)) != "COMPRAR" {
			continue
		}
		if rec.PrecoAtual <= 0 {
			continue
		}
		p, exists := portfolioMap[normalizeKey(rec.Fundo)]
		if !exists {
			continue
		}
		alocAlvo := math.Min(rec.Alocacao, maxAllocacaoPorAtivo)
		elegiveis = append(elegiveis, elegivel{
			rec:        rec,
			saldoAtual: p.SaldoBruto,
			alocAlvo:   alocAlvo,
		})
	}

	if len(elegiveis) == 0 {
		return &domain.ContributionResponse{
			ValorDisponivel: valor,
			ValorUtilizado:  0,
			ValorSobra:      valor,
			Fundos:          []domain.ContributionResult{},
		}, nil
	}

	// 2. Calcula o quanto cada fundo deve receber do valor disponível.
	//    Estratégia: prioriza fundos abaixo do alvo (deficit); se todos já estão
	//    acima do alvo, distribui proporcionalmente ao alocAlvo.
	type alocacao struct {
		el        elegivel
		investir  float64
	}

	var alocacoes []alocacao

	// Calcula deficits
	totalDeficit := 0.0
	for _, e := range elegiveis {
		alvo := totalAlvo * (e.alocAlvo / 100.0)
		deficit := math.Max(0, alvo-e.saldoAtual)
		totalDeficit += deficit
	}

	if totalDeficit > 0 {
		// Distribui proporcionalmente ao deficit de cada fundo
		for _, e := range elegiveis {
			alvo := totalAlvo * (e.alocAlvo / 100.0)
			deficit := math.Max(0, alvo-e.saldoAtual)
			if deficit == 0 {
				continue
			}
			investir := valor * (deficit / totalDeficit)
			alocacoes = append(alocacoes, alocacao{el: e, investir: investir})
		}
	} else {
		// Todos já estão acima do alvo: distribui proporcionalmente à alocação alvo
		somaAlvo := 0.0
		for _, e := range elegiveis {
			somaAlvo += e.alocAlvo
		}
		for _, e := range elegiveis {
			investir := valor * (e.alocAlvo / somaAlvo)
			alocacoes = append(alocacoes, alocacao{el: e, investir: investir})
		}
	}

	// 3. Aplica o teto de 10% por ativo
	for i, a := range alocacoes {
		tetoValor := totalAlvo*(maxAllocacaoPorAtivo/100.0) - a.el.saldoAtual
		if tetoValor < a.investir {
			alocacoes[i].investir = math.Max(0, tetoValor)
		}
	}

	// 4. Calcula cotas inteiras iniciais
	type resultado struct {
		el     elegivel
		qtd    int
		gasto  float64
		investir float64
	}

	var resultMap []resultado
	totalGasto := 0.0

	for _, a := range alocacoes {
		if a.investir <= 0 {
			continue
		}
		qtd := int(math.Floor(a.investir / a.el.rec.PrecoAtual))
		gasto := float64(qtd) * a.el.rec.PrecoAtual
		totalGasto += gasto
		resultMap = append(resultMap, resultado{el: a.el, qtd: qtd, gasto: gasto, investir: a.investir})
	}

	// 5. Reaproveita a sobra comprando cotas extras nos fundos mais baratos
	//    que ainda têm espaço (abaixo do teto de 10%)
	sobra := valor - totalGasto
	melhorou := true
	for melhorou && sobra > 0 {
		melhorou = false
		// Ordena por preço crescente para maximizar uso da sobra
		for i := range resultMap {
			preco := resultMap[i].el.rec.PrecoAtual
			if preco > sobra {
				continue
			}
			// Verifica teto de 10%
			gastoAtual := resultMap[i].gasto
			saldo := resultMap[i].el.saldoAtual
			tetoValor := totalAlvo*(maxAllocacaoPorAtivo/100.0) - saldo - gastoAtual
			if tetoValor < preco {
				continue
			}
			// Compra mais 1 cota
			resultMap[i].qtd++
			resultMap[i].gasto += preco
			totalGasto += preco
			sobra -= preco
			melhorou = true
		}
	}

	// 6. Monta resultado final
	var resultados []domain.ContributionResult
	for _, r := range resultMap {
		if r.qtd == 0 {
			continue
		}
		alocacaoApos := 0.0
		if totalAlvo > 0 {
			alocacaoApos = ((r.el.saldoAtual + r.gasto) / totalAlvo) * 100.0
		}
		resultados = append(resultados, domain.ContributionResult{
			Fundo:         r.el.rec.Fundo,
			Segmento:      r.el.rec.Segmento,
			PrecoAtual:    r.el.rec.PrecoAtual,
			AlocacaoAtual: math.Round(alocacaoApos*100) / 100,
			AlocacaoAlvo:  r.el.alocAlvo,
			ValorAportar:  math.Round(r.investir*100) / 100,
			QtdComprar:    r.qtd,
			TotalGasto:    math.Round(r.gasto*100) / 100,
		})
	}

	return &domain.ContributionResponse{
		ValorDisponivel: valor,
		ValorUtilizado:  math.Round(totalGasto*100) / 100,
		ValorSobra:      math.Round((valor-totalGasto)*100) / 100,
		Fundos:          resultados,
	}, nil
}

func normalizeKey(name string) string {
	return strings.ToUpper(strings.TrimSpace(name))
}
