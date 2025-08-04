package postgresql

import (
	"gorm.io/gorm"
)

// PostgreSQLRepo PostgreSQL資料庫存取庫
type PostgreSQLRepo struct {
	PostgreSQLDB *gorm.DB
}

// NewPostgreSQLRepo 創建PostgreSQL存取庫實例
func NewPostgreSQLRepo(dbPostgreSQL *gorm.DB) *PostgreSQLRepo {
	return &PostgreSQLRepo{PostgreSQLDB: dbPostgreSQL}
}

// DB 獲取資料庫連接
func (pg *PostgreSQLRepo) DB() *gorm.DB {
	return pg.PostgreSQLDB
}
