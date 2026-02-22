package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
)

const validPortfolioCSV = `Nome do ativo,Qtd,Preco Medio,Proventos,Preco de mercado,Resultado C/ Proventos,Saldo bruto
HGLG11FII HEDGE GLOBAL,10,100.00,5.00,110.00,55.00,1100.00
`

const invalidPortfolioCSV = `Coluna errada,Outra coluna
valor1,valor2
`

func TestPortfolioService_GetAll_Sucesso(t *testing.T) {
	items := []domain.PortfolioItem{{Nome: "HGLG11", SaldoBruto: 1100}}
	svc := NewPortfolioService(&mockPortfolioRepo{items: items})

	result, err := svc.GetAll()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("esperava 1 item; got %d", len(result))
	}
	if result[0].Nome != "HGLG11" {
		t.Errorf("esperava Nome=HGLG11; got %s", result[0].Nome)
	}
}

func TestPortfolioService_GetAll_RepoFalha_RetornaErrInternal(t *testing.T) {
	svc := NewPortfolioService(&mockPortfolioRepo{err: fmt.Errorf("db error")})

	_, err := svc.GetAll()
	if err == nil {
		t.Fatal("esperava erro")
	}
	if !errors.Is(err, domain.ErrInternal) {
		t.Errorf("esperava domain.ErrInternal; got: %v", err)
	}
}

func TestPortfolioService_ImportCSV_CSVValido(t *testing.T) {
	repo := &mockPortfolioRepo{}
	svc := NewPortfolioService(repo)

	items, err := svc.ImportCSV(strings.NewReader(validPortfolioCSV))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("esperava 1 item; got %d", len(items))
	}
	if items[0].Nome != "HGLG11" {
		t.Errorf("esperava ticker HGLG11; got %s", items[0].Nome)
	}
}

func TestPortfolioService_ImportCSV_CSVInvalido_NaoRetornaErrInternal(t *testing.T) {
	svc := NewPortfolioService(&mockPortfolioRepo{})

	_, err := svc.ImportCSV(strings.NewReader(invalidPortfolioCSV))
	if err == nil {
		t.Fatal("esperava erro para CSV inválido")
	}
	// Erro de parsing é de responsabilidade do cliente (400), não infraestrutura (500)
	if errors.Is(err, domain.ErrInternal) {
		t.Error("erro de CSV nao deve ser ErrInternal")
	}
}

func TestPortfolioService_ImportCSV_SaveAllFalha_RetornaErrInternal(t *testing.T) {
	repo := &mockPortfolioRepo{err: fmt.Errorf("disk full")}
	svc := NewPortfolioService(repo)

	_, err := svc.ImportCSV(strings.NewReader(validPortfolioCSV))
	if err == nil {
		t.Fatal("esperava erro")
	}
	if !errors.Is(err, domain.ErrInternal) {
		t.Errorf("esperava domain.ErrInternal; got: %v", err)
	}
}
