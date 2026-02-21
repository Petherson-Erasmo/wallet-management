# wallet-management

API REST em Go para gerenciamento de carteira de Fundos Imobiliários (FIIs).

## Funcionalidades

- Importa a carteira atual via CSV mensal
- Importa a carteira recomendada via CSV mensal
- Calcula quanto comprar de cada fundo para um aporte, respeitando:
  - A recomendação do fundo (COMPRAR / AGUARDAR)
  - A alocação alvo definida na carteira recomendada
  - O teto de **10% por ativo** em relação ao total da carteira
  - Distribuição proporcional se o orçamento for insuficiente

## Stack

- **Go 1.21+**
- **Gin** — HTTP router
- **GORM + SQLite** — persistência de dados

## Estrutura do projeto

```
cmd/api/            → entry point (main.go)
internal/
  csvparser/        → parse dos arquivos CSV
  domain/           → modelos de domínio
  handler/          → handlers HTTP e roteador
  repository/       → acesso ao banco de dados
  service/          → regras de negócio
data/               → banco SQLite (gerado automaticamente)
```

## Como executar

```bash
go run cmd/api/main.go
# Servidor sobe na porta 8080
```

Variáveis de ambiente opcionais:

| Variável       | Padrão           | Descrição               |
| -------------- | ---------------- | ----------------------- |
| `PORT`         | `8080`           | Porta do servidor       |
| `DATABASE_DSN` | `data/wallet.db` | Caminho do banco SQLite |

## Endpoints

### `POST /api/v1/portfolio`

Importa a carteira de investimentos. Substitui os dados anteriores.

**Form-data:** `file` — arquivo CSV com colunas:
`Nome do ativo, Qtd, Preco Medio, Proventos, Preco de mercado, Resultado C/ Proventos, Saldo bruto`

```bash
curl -X POST http://localhost:8080/api/v1/portfolio \
  -F "file=@carteira.csv"
```

---

### `GET /api/v1/portfolio`

Retorna a carteira atual importada.

---

### `POST /api/v1/recommendation`

Importa a carteira recomendada mensal. Substitui os dados anteriores.

**Form-data:** `file` — arquivo CSV com colunas:
`FUNDO, SEGMENTO, PRECO ATUAL, PRECO MEDIO, PRECO TETO, ALOCACAO, RECOMENDACAO`

```bash
curl -X POST http://localhost:8080/api/v1/recommendation \
  -F "file=@recomendacao.csv"
```

---

### `GET /api/v1/recommendation`

Retorna a carteira recomendada atual.

---

### `GET /api/v1/contribution?valor=1000`

Calcula o plano de aporte para o valor informado.

**Query param:** `valor` — valor disponível para investimento

**Exemplo de resposta:**

```json
{
  "valor_disponivel": 1000,
  "valor_utilizado": 960,
  "valor_sobra": 40,
  "fundos": [
    {
      "fundo": "HGLG11",
      "segmento": "Logistica",
      "preco_atual": 175,
      "alocacao_atual_pct": 7.8,
      "alocacao_alvo_pct": 8,
      "valor_a_aportar": 480,
      "qtd_a_comprar": 2,
      "total_gasto": 350
    }
  ]
}
```

## Algoritmo de aporte

1. Descarta fundos com `RECOMENDACAO = AGUARDAR`
2. Calcula `total_alvo = saldo_atual + valor`
3. Para cada fundo elegível:
   - `alvo = total_alvo × min(ALOCACAO%, 10%)`
   - `necessario = alvo - saldo_atual_do_fundo`
4. Se a soma necessária supera o valor, distribui proporcionalmente
5. Calcula `qtd = floor(investimento / preco_atual)` (cotas inteiras)
6. Retorna o plano com sobra de caixa

Tenho uma carteira de fundos imobiliários e todo mês tenho um valor X para investir. Esse valor deve ser distribuído igualmente entre todos os fundos da minha carteira, respeitando a porcentagem máxima de alocação que eu defini e considerando se pode ou não comprar ("comprar" ou "aguardar"). Caso alguns fundos estejam como aguardar é esperado que os outros fiquem com uma porcentagem acima do ideal, mas não deve nunca ultrapassar de 10% em um ativo.

Todo mês os valores dos ativos sofrem ajustes de acordo com o preço do mercado, então todo mês vou enviar um csv com a minha carteira atualizada contendo os dados dos fundos imobiliários (Nome do ativo, Qtd, Preço Médio, Proventos, Preço de mercado, Resultado c/ proventos, Saldo bruto). Todo mês existe um relatório de recomendação (carteira recomendada) que deve ser considerado para saber quais fundos comprar ou não e em qual porcentagem. Esses dados também serão enviados em um csv contendo como dados FUNDO, SEGMENTO, \*PREÇO ATUAL, PREÇO MÉDIO, PREÇO TETO, ALOCAÇÃO, RECOMENDAÇÃO.

Detalhamento técnico:

- Siga as boas práticas de progração
- Crie uma api rest usando golang e salvando os dados
- Crie um endpoint para receber a carteira de investimentos
- Crie um endpoint para receber a carteira recomendada
- Crie um endpoint para receber o valor a ser aportado em que deve ter como resposta a lista dos arquivos a aportar e quanto comprar de cada um, considerando o preço atual da carteira recomendada
