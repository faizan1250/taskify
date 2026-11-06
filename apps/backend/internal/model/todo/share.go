package todo

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ShareTokenLength is the byte length of a share token before hex encoding.
// 32 bytes -> 64 hex chars, unguessable capability URL.
const ShareTokenLength = 32

// TodoShareToken is a bearer capability granting public read access to one
// todo. The token itself is the credential: anyone holding it can read.
// Revocation deletes the row (see UnshareTodo); revoked_at marks soft-revoked
// rows when retained for audit.
type TodoShareToken struct {
	Token     string     `json:"token" db:"token"`
	TodoID    uuid.UUID  `json:"todoId" db:"todo_id"`
	CreatedBy string     `json:"createdBy" db:"created_by"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	RevokedAt *time.Time `json:"revokedAt" db:"revoked_at"`
}

// GenerateShareToken mints a cryptographically random hex capability token.
func GenerateShareToken() (string, error) {
	b := make([]byte, ShareTokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ------------------------------------------------------------

type ShareTodoPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *ShareTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type UnshareTodoPayload struct {
	ID uuid.UUID `param:"id" validate:"required,uuid"`
}

func (p *UnshareTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

type GetSharedTodoPayload struct {
	Token string `param:"token" validate:"required,min=16,max=128"`
}

func (p *GetSharedTodoPayload) Validate() error {
	validate := validator.New()
	return validate.Struct(p)
}

// ------------------------------------------------------------

// ShareTodoResult is the explicit response DTO for POST /todos/:id/share.
// Only the token is returned; the todo itself is read via GET /shared/:token.
type ShareTodoResult struct {
	Token  string    `json:"token"`
	TodoID uuid.UUID `json:"todoId"`
}
