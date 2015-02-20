package readeef

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/urandom/webfw"
)

var (
	dbVersion   = 1
	initSql     = make(map[string][]string)
	sqlStmt     = make(map[string]string)
	upgradeFunc = make(map[string]func(db DB, old, new int) error)
)

type Validator interface {
	Validate() error
}

type DB struct {
	*sqlx.DB
	logger        webfw.Logger
	driver        string
	connectString string
}

type ValidationError struct {
	error
}

func NewDB(driver, conn string, logger webfw.Logger) DB {
	return DB{driver: driver, connectString: conn, logger: logger}
}

func (db *DB) Connect() error {
	dbx, err := sqlx.Connect(db.driver, db.connectString)
	if err != nil {
		return err
	}

	db.DB = dbx

	return db.init()
}

func (db DB) init() error {
	if init, ok := initSql[db.driver]; ok {
		for _, sql := range init {
			_, err := db.Exec(sql)
			if err != nil {
				return fmt.Errorf("Error executing '%s': %v", sql, err)
			}
		}
	} else {
		return fmt.Errorf("No init sql for driver '%s'", db.driver)
	}

	var version int
	if err := db.Get(&version, "SELECT db_version FROM readeef"); err != nil {
		if err == sql.ErrNoRows {
			version = dbVersion
		} else {
			return fmt.Errorf("Error getting the current db_version: %v\n", err)
		}
	}

	if version > dbVersion {
		panic(fmt.Sprintf("The db version '%d' is newer than the expected '%d'", version, dbVersion))
	}

	if version < dbVersion {
		db.logger.Infof("Database version mismatch: current is %d, expected %d\n", version, dbVersion)
		if upgrade, ok := upgradeFunc[db.driver]; ok {
			db.logger.Infof("Running upgrade function for %s driver\n", db.driver)
			if err := upgrade(db, version, dbVersion); err != nil {
				return fmt.Errorf("Error running upgrade function for %s driver: %v\n", db.driver, err)
			}
		}
	}

	_, err := db.Exec(`DELETE FROM readeef`)
	/* TODO: per-database statements */
	if err == nil {
		_, err = db.Exec(`INSERT INTO readeef(db_version) VALUES($1)`, dbVersion)
	}
	if err != nil {
		return fmt.Errorf("Error initializing readeef utility table: %v", err)
	}

	return nil
}

func (db DB) NamedSQL(name string) string {
	var stmt string

	if stmt = sqlStmt[db.driver+":"+name]; stmt == "" {
		stmt = sqlStmt["generic:"+name]
	}

	if stmt == "" {
		panic("No statement for name " + name)
	}

	return stmt
}

func (db DB) CreateWithId(stmt *sqlx.Stmt, args ...interface{}) (int64, error) {
	var id int64

	if db.driver == "postgres" {
		err := stmt.QueryRow(args...).Scan(&id)
		if err != nil {
			return 0, err
		}
	} else {
		res, err := stmt.Exec(args...)
		if err != nil {
			return 0, err
		}

		id, err = res.LastInsertId()
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}
