package upgrades

import "database/sql"

func init() {
	upgrades[39] = upgrade{"Add backfill queue", func(tx *sql.Tx, ctx context) error {
		_, err := tx.Exec(`
			CREATE TABLE backfill_queue (
				queue_id            INTEGER PRIMARY KEY,
				user_mxid           TEXT,
				type                INTEGER NOT NULL,
				priority            INTEGER NOT NULL,
				portal_jid          VARCHAR(255),
				portal_receiver     VARCHAR(255),
				time_start          TIMESTAMP,
				time_end            TIMESTAMP,
				max_batch_events    INTEGER NOT NULL,
				max_total_events    INTEGER,
				batch_delay         INTEGER,

				FOREIGN KEY (user_mxid) REFERENCES "user"(mxid) ON DELETE CASCADE ON UPDATE CASCADE,
				FOREIGN KEY (portal_jid, portal_receiver) REFERENCES portal(jid, receiver) ON DELETE CASCADE
			)
		`)
		if err != nil {
			return err
		}

		// The queue_id needs to auto-increment every insertion. For SQLite,
		// INTEGER PRIMARY KEY is an alias for the ROWID, so it will
		// auto-increment. See https://sqlite.org/lang_createtable.html#rowid
		// For Postgres, we have to manually add the sequence.
		if ctx.dialect == Postgres {
			_, err = tx.Exec(`
				CREATE SEQUENCE backfill_queue_queue_id_seq
				START WITH 1
				OWNED BY backfill_queue.queue_id
			`)
			if err != nil {
				return err
			}
			_, err = tx.Exec(`
				ALTER TABLE backfill_queue
				ALTER COLUMN queue_id
				SET DEFAULT nextval('backfill_queue_queue_id_seq'::regclass)
			`)
			if err != nil {
				return err
			}
		}

		return err
	}}
}
