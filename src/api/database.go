package api

import (
	"database/sql"
	"fmt"
	"time"

	// this is required
	_ "github.com/go-sql-driver/mysql"
)

const (
	refreshTime = time.Minute * 1

	// Database constants
	dbUser      = "root"
	dbPassword  = "01Hgko05"
	dbHost      = "192.168.50.104:3306"
	dbSchema    = "shellpayvote"
	dbTableName = "project_coins"

	// SQL statements
	sqlCreateTable = `
	CREATE TABLE "project_coins" (
		"name" VARCHAR(45) NOT NULL,
		PRIMARY KEY ("name"))
		COMMENT = 'This table contains all information regarding project coins' `

	sqlUpdateBalance = `UPDATE ` + dbTableName +
		` SET balance = ? WHERE name = ? `

	sqlUpdateStatus = `UPDATE ` + dbTableName +
		` SET status = ? WEHRE name = ? `

	sqlInsertNewCoin = `INSERT INTO ` + dbTableName + `
	(name, symbol, name_cn, platform_coin_name, platform_coin_name_cn, platform_coin_symbol, logo, vote_cap, 
		balance, balance_check_time, short_description, long_description, issuer, time_opening, time_closed, 
		close_reason, voting_address, seed, private_key, status) 
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
)

var (
	dbInstance          *sql.DB
	dbInsertStmt        *sql.Stmt
	dbUpdateBalanceStmt *sql.Stmt
	dbUpdateStatusStmt  *sql.Stmt
)

func init() {
	// create a DB connection
	conn := fmt.Sprintf("%s:%s@tcp(%s)/%s",
		dbUser, dbPassword, dbHost, dbSchema)

	dbInstance, err := sql.Open("mysql", conn)
	if err != nil {
		panic(fmt.Sprintf("failed to open database connection: %s", err))
	}

	dbInsertStmt, err = dbInstance.Prepare(sqlInsertNewCoin)
	if err != nil {
		panic(fmt.Sprintf("failed to create insert statement: %s", err))
	}

	dbUpdateBalanceStmt, err = dbInstance.Prepare(sqlUpdateBalance)
	if err != nil {
		panic(fmt.Sprintf("failed to create update balance statement: %s", err))
	}

	dbUpdateStatusStmt, err = dbInstance.Prepare(sqlUpdateStatus)
	if err != nil {
		panic(fmt.Sprintf("failed to create update status statement : %s", err))
	}
}

// updateBalance updates balance every 1 minute
// updateBalance will use service provided by superwallet.shellpay2.com
func updateBalance(b float64, coinName string) error {
	_, err := dbUpdateBalanceStmt.Exec(b, coinName)
	if err != nil {
		return err
	}

	return nil
}

func addProjectCoin(coin ProjectCoin) error {

	_, err := dbInsertStmt.Exec(
		coin.Name, coin.Symbol, coin.NameCN,
		coin.PlatformCoinName, coin.PlatformCoinNameCN, coin.PlatformCoinSymbol,
		coin.Logo, coin.VoteCap,
		coin.Balance, coin.BalanceCheckTime,
		coin.ShortDescription, coin.LongDescription,
		coin.Issuer,
		coin.TimeOpening, coin.TimeClosed, coin.CloseReason,
		coin.VotingAddress, coin.Seed, coin.PrivateKey,
		coin.Status)

	if err != nil {
		return err
	}

	return nil
}

// udpateStatus change project coin status from Open to either Closed or Aborted
func updateStatus(s string, coinName string) error {

	_, err := dbUpdateStatusStmt.Exec(s, coinName)

	if err != nil {
		return err
	}

	return nil
}

// notifyStatusChange notify Shellpay of a status change by email
func notifyStatusChange() {

}

// creaetDatabase creates a blank database
func createDatabase() {

}

// CloseDatabaseConn release all database resources
func CloseDatabaseConn() {
	if dbInsertStmt != nil {
		dbInsertStmt.Close()
	}

	if dbUpdateStatusStmt != nil {
		dbUpdateStatusStmt.Close()
	}

	if dbUpdateBalanceStmt != nil {
		dbUpdateBalanceStmt.Close()
	}

	if dbInstance != nil {
		dbInstance.Close()
	}
}
