package db

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/MangoMilk/go-lib/dwarflog"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"time"
)

type MysqlMode string

const (
	MysqlParent = MysqlMode("parent")
	MysqlChild  = MysqlMode("child")
)

type MysqlConfig struct {
	Host     string    `yaml:"Host"`
	Port     int       `yaml:"Port"`
	User     string    `yaml:"User"`
	Password string    `yaml:"Password"`
	Database string    `yaml:"Database"`
	Mode     MysqlMode `yaml:"Mode"`
}

type Mysql struct {
	Instance *sql.DB
}

var (
	mysql   *Mysql
	ErrNoTx = errors.New("no tx")
)

func Setup(configs []MysqlConfig) *Mysql {
	if mysql != nil {
		return mysql
	}

	mysql = NewMysql(configs[0])
	mysql.Open()

	return mysql
}

func NewMysql(config MysqlConfig) *Mysql {
	// gen config
	var dsn string = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4", config.User, config.Password, config.Host, config.Port, config.Database)

	// check config
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		// log.Println("dsn: " + dsn)
		panic("连接配置错误: " + err.Error())
	}

	// set conn config
	db.SetConnMaxLifetime(time.Minute * 3) // set client max life time less than mysql param "wait_timeout"
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return &Mysql{
		Instance: db,
	}
}

func (db Mysql) Open() {
	err := db.Instance.Ping()
	if err != nil {
		panic(err)
	}
}

func (db Mysql) Close() {
	err := db.Instance.Close()
	if err != nil {
		panic(err)
	}
}

func Add(table string, insertData map[string]interface{}) (int64, error) {

	var fieldStr string = "("
	var placeHolderStr string = "("
	var args []interface{}

	for k, v := range insertData {
		fieldStr += "`" + k + "`,"
		placeHolderStr += "?,"
		args = append(args, v)
	}

	fieldStr = strings.Trim(fieldStr, ",") + ")"
	placeHolderStr = strings.Trim(placeHolderStr, ",") + ")"

	var sqlStr string = "INSERT INTO " + table + " " + fieldStr + " VALUES " + placeHolderStr

	res, insertErr := mysql.Instance.Exec(sqlStr, args...)

	// panic("sql:" + sqlStr)
	// panic(args)

	if insertErr != nil {
		dwarflog.Error(insertErr, sqlStr, args)
		return 0, insertErr
	}

	//插入数据的主键id
	lastInsertId, _ := res.LastInsertId()

	return lastInsertId, nil
}

/*
 * Update
 */
func Update(table string, updateData map[string]interface{}, condition map[string]interface{}) (int64, error) {

	var fieldStr string = ""
	var conditionStr string = ""
	var args []interface{}

	// update data
	for k, v := range updateData {
		fieldStr += "`" + k + "`=?,"
		args = append(args, v)
	}

	fieldStr = strings.Trim(fieldStr, ",")

	// condition
	var conditionCount int = len(condition)
	var c int = 1
	for key, val := range condition {
		conditionStr += "`" + key + "`=?"
		args = append(args, val)

		if c != conditionCount {
			conditionStr += " AND "
		}

		c++
	}

	var sqlStr string = "UPDATE " + table + " SET " + fieldStr + " WHERE " + conditionStr

	res, updateErr := mysql.Instance.Exec(sqlStr, args...)

	if updateErr != nil {
		dwarflog.Error(updateErr, sqlStr, args)
		return 0, updateErr
	}

	affectedRows, _ := res.RowsAffected()

	return affectedRows, nil

}

/*
 * Delete
 */
func Delete(table string, condition map[string]interface{}) (int64, error) {

	var conditionStr string = ""
	var args []interface{}

	// condition
	var conditionCount int = len(condition)
	var c int = 1
	for key, val := range condition {
		conditionStr += "`" + key + "`=?"
		args = append(args, val)

		if c != conditionCount {
			conditionStr += " AND "
		}

		c++
	}

	var sqlStr string = "DELETE FROM " + table + " WHERE " + conditionStr

	res, updateErr := mysql.Instance.Exec(sqlStr, args...)

	if updateErr != nil {
		dwarflog.Error(updateErr, sqlStr, args)
		return 0, updateErr
	}

	affectedRows, _ := res.RowsAffected()

	return affectedRows, nil
}

func QueryRow(query string, args ...interface{}) *sql.Row {
	return mysql.Instance.QueryRow(query, args...)
}

func Query(query string, args ...interface{}) (*sql.Rows, error) {
	return mysql.Instance.Query(query, args...)
}

func Exec(query string, args ...interface{}) (sql.Result, error) {
	return mysql.Instance.Exec(query, args...)
}
