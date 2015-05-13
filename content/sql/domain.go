package sql

import (
	"database/sql"

	"github.com/urandom/readeef/content/base"
	"github.com/urandom/readeef/content/sql/db"
	"github.com/urandom/webfw"
)

type Domain struct {
	base.Domain

	logger  webfw.Logger
	db      *db.DB
	checked bool
}

func (d *Domain) SupportsHTTPS() (supports bool) {
	if d.HasErr() {
		return
	}

	supports = d.Domain.SupportsHTTPS()
	if supports || d.checked {
		return
	}

	host := d.URL().Host
	d.logger.Infof("Check the db about https support for %s\n", host)

	var sup sql.NullBool
	var needsChecking bool
	if err := d.db.Get(&sup, d.db.SQL("get_domain_https_support"), host); err != nil {
		if err == sql.ErrNoRows {
			needsChecking = true
		} else {
			d.Err(err)
			return
		}
	} else {
		if sup.Valid {
			return sup.Bool
		} else {
			needsChecking = true
		}
	}

	if needsChecking {
		supports = d.CheckHTTPSSupport()

		u := d.URL()
		// This will produce a protocol-relative url, indicating that it supports HTTPS for future checks
		u.Scheme = ""

		d.URL(u.String())

		d.logger.Infof("Updating domain https support for %s\n", host)

		tx, err := d.db.Beginx()
		if err != nil {
			d.Err(err)
			return
		}
		defer tx.Rollback()

		stmt, err := tx.Preparex(d.db.SQL("update_domain_https_support"))
		if err != nil {
			d.Err(err)
			return
		}
		defer stmt.Close()

		res, err := stmt.Exec(host, supports)
		if err != nil {
			d.Err(err)
			return
		}

		if num, err := res.RowsAffected(); err == nil && num > 0 {
			if err := tx.Commit(); err != nil {
				d.Err(err)
			}

			return
		}

		stmt, err = tx.Preparex(d.db.SQL("create_domain_https_support"))
		if err != nil {
			d.Err(err)
			return
		}
		defer stmt.Close()

		_, err = stmt.Exec(supports, host)
		if err != nil {
			d.Err(err)
			return
		}

		if err := tx.Commit(); err != nil {
			d.Err(err)
		}

		d.checked = true
	}

	return
}
