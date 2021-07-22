package db

import (
	"database/sql"
	"github.com/MangoMilk/go-lib/dwarflog"
	"strings"
)

type TxInstance struct {
	Tx *sql.Tx
}

func BeginTx() (*TxInstance, error) {

	tx, err := mysql.Instance.Begin()
	if err != nil {
		return nil, err
	}

	return &TxInstance{tx}, nil
}

func (i *TxInstance) Commit() error {
	if i.Tx != nil {
		err := i.Tx.Commit()
		return err
	} else {
		return ErrNoTx
	}
}

func (i *TxInstance) Rollback() error {
	if i.Tx != nil {
		err := i.Tx.Rollback()
		return err
	} else {
		return ErrNoTx
	}
}

func (i *TxInstance) Add(table string, insertData map[string]interface{}) (int64, error) {

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

	res, insertErr := i.Tx.Exec(sqlStr, args...)

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

func (i *TxInstance) Update(table string, updateData map[string]interface{}, condition map[string]interface{}) (int64, error) {

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

	res, updateErr := i.Tx.Exec(sqlStr, args...)

	if updateErr != nil {
		dwarflog.Error(updateErr, sqlStr, args)
		return 0, updateErr
	}

	affectedRows, _ := res.RowsAffected()

	return affectedRows, nil

}

func (i *TxInstance) Delete(table string, condition map[string]interface{}) (int64, error) {

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

	res, updateErr := i.Tx.Exec(sqlStr, args...)

	if updateErr != nil {
		dwarflog.Error(updateErr, sqlStr, args)
		return 0, updateErr
	}

	affectedRows, _ := res.RowsAffected()

	return affectedRows, nil
}
