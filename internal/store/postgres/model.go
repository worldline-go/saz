package postgres

import (
	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/types"
)

type Note struct {
	ID        string                      `db:"id"         goqu:"skipupdate"`
	Name      string                      `db:"name"`
	Content   types.JSON[service.Content] `db:"content"`
	Path      string                      `db:"path"`
	UpdatedBy types.Null[string]          `db:"updated_by"`
	CreatedAt types.Null[types.Time]      `db:"created_at" goqu:"skipupdate"`
	UpdatedAt types.Null[types.Time]      `db:"updated_at"`
}

type NoteIDName struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}
