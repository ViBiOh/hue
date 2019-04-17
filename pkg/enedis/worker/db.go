package worker

import (
	"database/sql"

	"github.com/ViBiOh/httputils/pkg/db"
	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/iot/pkg/enedis"
)

const insertQuery = `
INSERT INTO
  enedis_value
(
  ts,
  value
) VALUES (
  to_timestamp($1),
  $2
);
`

func (a *App) saveValue(o *enedis.Value, tx *sql.Tx) (err error) {
	if o == nil {
		return errors.New("cannot save nil Value")
	}

	var usedTx *sql.Tx
	if usedTx, err = db.GetTx(a.db, tx); err != nil {
		return
	}

	if usedTx != tx {
		defer func() {
			err = db.EndTx(usedTx, err)
		}()
	}

	if _, err = usedTx.Exec(insertQuery, o.Timestamp, o.Valeur); err != nil {
		err = errors.WithStack(err)
		return
	}

	return
}
