package service

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
)

// ---------------------------------------------------------------------------
// Mocks que implementam as interfaces do ports.go sem nenhuma dependência
// ---------------------------------------------------------------------------

type mockPortfolioRepo struct {
	items []domain.PortfolioItem
	err   error
}

func (m *mockPortfolioRepo) FindAll() ([]domain.PortfolioItem, error) {
	return m.items, m.err
}
func (m *mockPortfolioRepo) SaveAll(items []domain.PortfolioItem) error {
	return m.err
}

type mockRecommendationRepo struct {
	items []domain.RecommendedItem
	err   error
}

func (m *mockRecommendationRepo) FindAll() ([]domain.RecommendedItem, error) {
	return m.items, m.err
}
func (m *mockRecommendationRepo) SaveAll(items []domain.RecommendedItem) error {
	return m.err
}

func newSvc(pItems []domain.PortfolioItem, pErr error, rItems []domain.RecommendedItem, rErr error) *ContributionService {
	return NewContributionService(
		&mockPortfolioRepo{items: pItems, err: pErr},
		&mockRecommendationRepo{items: rItems, err: rErr},
	)
}

// ---------------------------------------------------------------------------
// Testes de validação de entrada
// ---------------------------------------------------------------------------

func TestCalculate_ValorZero(t *testing.T) {
	svc := newSvc(nil, nil, nil, nil)
	_, err := svc.Calculate(0)
	if err == nil {
		t.Fatal("esperava erro para valor = 0")
	}
}

func TestCalculate_ValorNegativo(t *testing.T) {
	svc := newSvc(nil, nil, nil, nil)
	_, err := svc.Calculate(-100)
	if err == nil {
		t.Fatal("esperava erro para valor negativo")
	}
}

// ---------------------------------------------------------------------------
// Testes de erro de infraestrutura (deve emitir ErrInternal)
// ---------------------------------------------------------------------------

func TestCalculate_PortfolioRepoFalha_RetornaErrInternal(t *testing.T) {
	dbErr := fmt.Errorf("connection refused")
	svc := newSvc(nil, dbErr, nil, nil)

	_, err := svc.Calculate(1000)
	if err == nil {
		t.Fatal("esperava erro")
	}
	if !errors.Is(err, domain.ErrInternal) {
		t.Errorf("esperava domain.ErrInternal; got: %v", err)
	}
}

func TestCalculate_RecommendationRepoFalha_RetornaErrInternal(t *testing.T) {
	portfolio := []domain.PortfolioItem{
		{Nome: "HGLG11", SaldoBruto: 500},
	}
	dbErr := fmt.Errorf("timeout")
	svc := newSvc(portfolio, nil, nil, dbErr)

	_, err := svc.Calculate(1000)
	if err == nil {
		t.Fatal("esperava erro")
	}
	if !errors.Is(err, domain.ErrInternal) {
		t.Errorf("esperava domain.ErrInternal; got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Testes de pré-condição de negócio (deve emitir ErrPrecondition)
// ---------------------------------------------------------------------------

func TestCalculate_SemRecomendacoes_RetornaErrPrecondition(t *testing.T) {
	portfolio := []domain.PortfolioItem{
		{Nome: "HGLG11", SaldoBruto: 500},
	}
	svc := newSvc(portfolio, nil, []domain.RecommendedItem{}, nil)

	_, err := svc.Calculate(1000)
	if err == nil {
		t.Fatal("esperava erro")
	}
	if !errors.Is(err, domain.ErrPrecondition) {
		t.Errorf("esperava domain.ErrPrecondition; got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Testes do algoritmo de cálculo
// ---------------------------------------------------------------------------

// Nenhum fundo é elegível (recomendação AGUARDAR) → resposta vazia, sobra = valor
func TestCalculate_SemFundosElegiveis(t *testing.T) {
	portfolio := []domain.PortfolioItem{
		{Nome: "HGLG11", SaldoBruto: 5000},
	}
	recommended := []domain.RecommendedItem{
		{Fundo: "HGLG11", PrecoAtual: 100, Alocacao: 10, Recomendacao: "AGUARDAR"},
	}
	svc := newSvc(portfolio, nil, recommended, nil)

	result, err := svc.Calculate(1000)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result.Fundos) != 0 {
		t.Errorf("esperava 0 fundos; got %d", len(result.Fundos))
	}
	if result.ValorSobra != 1000 {
		t.Errorf("esperava ValorSobra = 1000; got %.2f", result.ValorSobra)
	}
	if result.ValorUtilizado != 0 {
		t.Errorf("esperava ValorUtilizado = 0; got %.2f", result.ValorUtilizado)
	}
}

// Fundo elegível com déficit, com espaço abaixo do teto de 10%
//
// Setup:
//   - totalAtual = 1000 (MXRF11=950, HGLG11=50)
//   - valor = 200 → totalAlvo = 1200
//   - teto por ativo = 1200 * 10% = 120
//   - HGLG11: saldoAtual=50, déficit = 120-50 = 70
//   - investir=200, capped em 70 (pelo teto)
//   - precoAtual=10 → qtd=7, gasto=70
func TestCalculate_FundoComDeficit(t *testing.T) {
	portfolio := []domain.PortfolioItem{
		{Nome: "MXRF11", SaldoBruto: 950},
		{Nome: "HGLG11", SaldoBruto: 50},
	}
	recommended := []domain.RecommendedItem{
		{Fundo: "HGLG11", PrecoAtual: 10, Alocacao: 10, Recomendacao: "COMPRAR"},
	}
	svc := newSvc(portfolio, nil, recommended, nil)

	result, err := svc.Calculate(200)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result.Fundos) != 1 {
		t.Fatalf("esperava 1 fundo; got %d", len(result.Fundos))
	}

	f := result.Fundos[0]
	if f.QtdComprar != 7 {
		t.Errorf("esperava QtdComprar=7; got %d", f.QtdComprar)
	}
	if f.TotalGasto != 70 {
		t.Errorf("esperava TotalGasto=70; got %.2f", f.TotalGasto)
	}
	if result.ValorUtilizado != 70 {
		t.Errorf("esperava ValorUtilizado=70; got %.2f", result.ValorUtilizado)
	}
	if result.ValorSobra != 130 {
		t.Errorf("esperava ValorSobra=130; got %.2f", result.ValorSobra)
	}
}

// Fundo já acima do teto de 10% → teto corta investimento a zero → nada comprado
//
// Setup:
//   - totalAtual = 1200 (MXRF11=1000, HGLG11=200)
//   - valor = 200 → totalAlvo = 1400
//   - teto HGLG11 = 1400*10% = 140; saldoAtual=200 → teto negativo → investir=0
func TestCalculate_FundoAcimaDoTeto(t *testing.T) {
	portfolio := []domain.PortfolioItem{
		{Nome: "MXRF11", SaldoBruto: 1000},
		{Nome: "HGLG11", SaldoBruto: 200},
	}
	recommended := []domain.RecommendedItem{
		{Fundo: "HGLG11", PrecoAtual: 10, Alocacao: 10, Recomendacao: "COMPRAR"},
	}
	svc := newSvc(portfolio, nil, recommended, nil)

	result, err := svc.Calculate(200)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result.Fundos) != 0 {
		t.Errorf("esperava 0 fundos (fundo acima do teto); got %d", len(result.Fundos))
	}
	if result.ValorUtilizado != 0 {
		t.Errorf("esperava ValorUtilizado=0; got %.2f", result.ValorUtilizado)
	}
}

// Fundo com Alocacao > 10% → alvo é capped em 10%
func TestCalculate_AlocacaoAcimaDoMaxCap(t *testing.T) {
	portfolio := []domain.PortfolioItem{
		{Nome: "HGLG11", SaldoBruto: 0},
	}
	recommended := []domain.RecommendedItem{
		// Alocacao = 25%, mas o sistema capa em 10%
		{Fundo: "HGLG11", PrecoAtual: 10, Alocacao: 25, Recomendacao: "COMPRAR"},
	}
	svc := newSvc(portfolio, nil, recommended, nil)

	result, err := svc.Calculate(1000)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// totalAlvo = 1000, teto = 1000*10% = 100 → qtd = 10
	if len(result.Fundos) != 1 {
		t.Fatalf("esperava 1 fundo; got %d", len(result.Fundos))
	}
	if result.Fundos[0].AlocacaoAlvo != 10 {
		t.Errorf("esperava AlocacaoAlvo=10 (capped); got %.2f", result.Fundos[0].AlocacaoAlvo)
	}
}

// Fundo não está na carteira atual → não é elegível
func TestCalculate_FundoNaoNaCarteira(t *testing.T) {
	portfolio := []domain.PortfolioItem{
		{Nome: "MXRF11", SaldoBruto: 1000},
	}
	recommended := []domain.RecommendedItem{
		{Fundo: "HGLG11", PrecoAtual: 100, Alocacao: 10, Recomendacao: "COMPRAR"},
	}
	svc := newSvc(portfolio, nil, recommended, nil)

	result, err := svc.Calculate(500)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result.Fundos) != 0 {
		t.Errorf("esperava 0 fundos (fundo nao na carteira); got %d", len(result.Fundos))
	}
}

// ValorDisponivel deve sempre refletir o valor passado como argumento
func TestCalculate_ValorDisponivelCorreto(t *testing.T) {
	portfolio := []domain.PortfolioItem{
		{Nome: "HGLG11", SaldoBruto: 0},
	}
	recommended := []domain.RecommendedItem{
		{Fundo: "HGLG11", PrecoAtual: 1000, Alocacao: 10, Recomendacao: "COMPRAR"},
	}
	svc := newSvc(portfolio, nil, recommended, nil)

	const valor = 4000.50
	result, err := svc.Calculate(valor)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.ValorDisponivel != valor {
		t.Errorf("esperava ValorDisponivel=%.2f; got %.2f", valor, result.ValorDisponivel)
	}
}
