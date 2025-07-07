package repo

import (
	"gorm.io/gorm"
)

type MysqlRepo struct {
	MysqlDB *gorm.DB
}

func NewMysqlRepo(dbMysql *gorm.DB) *MysqlRepo {
	return &MysqlRepo{MysqlDB: dbMysql}
}

func (ma *MysqlRepo) DB() *gorm.DB {
	return ma.MysqlDB
}
