package service

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Petherson-Erasmo/wallet-management/internal/domain"
)

const validRecommendationCSV = `Fundo,Segmento,Preco Atual,Preco Medio,Preco Teto,Alocacao,Recomendacao
HGLG11,Logistica,110.00,100.00,120.00,10.00,COMPRAR
`

const invalidRecommendationCSV = `Coluna errada,Outra coluna
valor1,valor2
`

func TestRecommendationService_GetAll_Sucesso(t *testing.T) {
	items := []domain.RecommendedItem{{Fundo: "HGLG11", Alocacao: 10}}
	svc := NewRecommendationService(&mockRecommendationRepo{items: items})

	result, err := svc.GetAll()
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("esperava 1 item; got %d", len(result))
	}
	if result[0].Fundo != "HGLG11" {
		t.Errorf("esperava Fundo=HGLG11; got %s", result[0].Fundo)
	}
}

func TestRecommendationService_GetAll_RepoFalha_RetornaErrInternal(t *testing.T) {
	svc := NewRecommendationService(&mockRecommendationRepo{err: fmt.Errorf("db error")})

	_, err := svc.GetAll()
	if err == nil {
		t.Fatal("esperava erro")
	}
	if !errors.Is(err, domain.ErrInternal) {
		t.Errorf("esperava domain.ErrInternal; got: %v", err)
	}
}

func TestRecommendationService_ImportCSV_CSVValido(t *testing.T) {
	repo := &mockRecommendationRepo{}
	svc := NewRecommendationService(repo)

	items, err := svc.ImportCSV(strings.NewReader(validRecommendationCSV))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("esperava 1 item; got %d", len(items))
	}
	if items[0].Fundo != "HGLG11" {
		t.Errorf("esperava Fundo=HGLG11; got %s", items[0].Fundo)
	}
	if items[0].Recomendacao != "COMPRAR" {
		t.Errorf("esperava Recomendacao=COMPRAR; got %s", items[0].Recomendacao)
	}
}

func TestRecommendationService_ImportCSV_CSVInvalido_NaoRetornaErrInternal(t *testing.T) {
	svc := NewRecommendationService(&mockRecommendationRepo{})

	_, err := svc.ImportCSV(strings.NewReader(invalidRecommendationCSV))
	if err == nil {
		t.Fatal("esperava erro para CSV inválido")
	}
	// Erro de parsing é de responsabilidade do cliente (400), não infraestrutura (500)
	if errors.Is(err, domain.ErrInternal) {
		t.Error("erro de CSV nao deve ser ErrInternal")
	}
}

func TestRecommendationService_ImportCSV_SaveAllFalha_RetornaErrInternal(t *testing.T) {
	repo := &mockRecommendationRepo{err: fmt.Errorf("disk full")}
	svc := NewRecommendationService(repo)

	_, err := svc.ImportCSV(strings.NewReader(validRecommendationCSV))
	if err == nil {
		t.Fatal("esperava erro")
	}
	if !errors.Is(err, domain.ErrInternal) {
		t.Errorf("esperava domain.ErrInternal; got: %v", err)
	}
}
