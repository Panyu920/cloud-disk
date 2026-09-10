package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type Store struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db:      db,
		Queries: New(db),
	}
}

var StoreInstance *Store

func init() {
	// 初始化数据库连接
	db, err := sql.Open("mysql", "panyu:panyu@tcp(localhost:3306)/cloud-disk?multiStatements=true&parseTime=true")
	if err != nil {
		log.Fatalf("can not connect to %v: %v", "panyu:panyu@tcp(localhost:3306)/cloud-disk?multiStatements=true&parseTime=true\n", err)
	}
	// 真正尝试连接数据库
	if err := db.Ping(); err != nil {
		log.Fatalf("无法连接数据库: %v", err)
	}

	StoreInstance = NewStore(db)
}
