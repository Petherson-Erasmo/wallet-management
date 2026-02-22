package domain

import "errors"

// ErrPrecondition é retornado quando uma pré-condição de negócio não foi satisfeita
// (ex: carteira recomendada não importada). O handler deve responder com 422.
var ErrPrecondition = errors.New("precondition not met")

// ErrInternal é retornado quando ocorre uma falha de infraestrutura (ex: banco de dados).
// O handler deve responder com 500.
var ErrInternal = errors.New("internal error")
