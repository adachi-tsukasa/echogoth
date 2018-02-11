package main

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

type config struct {
	DB DBConfig
}

func gormConnect() *gorm.DB {
	var dbconfig config
	_, err := toml.DecodeFile("../config/db.toml", &dbconfig)
	if err != nil {
		panic(err.Error())
	}
	db, err := gorm.Open(dbconfig.DB.DBMS, fmt.Sprintf("%s:%s@%s/%s",
		dbconfig.DB.USER, dbconfig.DB.PASS, dbconfig.DB.PROTOCOL, dbconfig.DB.DBNAME))
	if err != nil {
		panic(err.Error())
	}
	return db
}
