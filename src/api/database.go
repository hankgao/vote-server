package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"time" // this is required

	_ "github.com/go-sql-driver/mysql"
	skyutil "github.com/hankgao/superwallet-server/server/mobile"
	"github.com/skycoin/skycoin/src/util/droplet"
	"github.com/skycoin/skycoin/src/wallet"
)

const (
	refreshTime = time.Minute * 1

	// Database constants
	dbUser      = "vote"
	dbPassword  = "Vote#1234"      // Note: don't use password that contains special characters like Test#1234
	dbHost      = "localhost:3306" // "localhost:3306"
	dbSchema    = "shellpayvote"
	dbTableName = "project_coins"

	// SQL statements
	sqlCreateTable = `
	CREATE TABLE "project_coins" (
		"name" VARCHAR(45) NOT NULL,
		PRIMARY KEY ("name"))
		COMMENT = 'This table contains all information regarding project coins' `

	sqlUpdateBalance = `UPDATE ` + dbTableName +
		` SET balance = ?, balance_check_time = ? WHERE name = ? `

	sqlUpdateStatus = `UPDATE ` + dbTableName +
		` SET status = ? WHERE name = ? `

	sqlInsertNewCoin = `INSERT INTO ` + dbTableName + `
	(name, symbol, name_cn, platform_coin_name, platform_coin_name_cn, platform_coin_symbol, logo, vote_cap, 
		balance, balance_check_time, short_description, long_description, issuer, time_opening, time_closed, 
		close_reason, voting_address, seed, private_key, status) 
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	sqlSelect = `SELECT * FROM ` + dbTableName + ` WHERE status = ? `

	sqlSelectByName = `SELECT * FROM ` + dbTableName + ` WHERE name = ? `
)

var (
	dbInstance          *sql.DB
	dbInsertStmt        *sql.Stmt
	dbUpdateBalanceStmt *sql.Stmt
	dbUpdateStatusStmt  *sql.Stmt
	dbSelectStmt        *sql.Stmt
	dbSelectByNameStmt  *sql.Stmt
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

	dbSelectStmt, err = dbInstance.Prepare(sqlSelect)
	if err != nil {
		panic(fmt.Sprintf("failed to create select statement: %s", err))
	}

	dbSelectByNameStmt, err = dbInstance.Prepare(sqlSelectByName)
	if err != nil {
		panic(fmt.Sprintf("failed to create select by name statement: %s", err))
	}

	skyutil.SetServer("http://superwallet.shellpay2.com:6789")

}

// UpdateBalance will use service provided by superwallet.shellpay2.com
func UpdateBalance(b float64, coinName string) error {
	now := time.Now()
	_, err := dbUpdateBalanceStmt.Exec(b, now.Format("2006-01-02 15:04:05"), coinName)
	if err != nil {
		return err
	}

	return nil
}

// RetrieveProjectCoins retrieves all project coins with a given status
func RetrieveProjectCoins(status string) []ProjectCoin {
	coins := []ProjectCoin{}

	rows, err := dbSelectStmt.Query(status)
	if err != nil {
		return coins
	}

	for rows.Next() {
		coin := ProjectCoin{}
		err = rows.Scan(
			&coin.Name, &coin.Symbol, &coin.NameCN,
			&coin.PlatformCoinName, &coin.PlatformCoinNameCN, &coin.PlatformCoinSymbol,
			&coin.Logo, &coin.VoteCap,
			&coin.Balance, &coin.BalanceCheckTime, // these two coulumns might not be necessary
			&coin.ShortDescription, &coin.LongDescription,
			&coin.Issuer,
			&coin.TimeOpening, &coin.TimeClosed, &coin.CloseReason,
			&coin.VotingAddress, &coin.Seed, &coin.PrivateKey,
			&coin.Status)

		if err != nil {
			// Log error here, please
			return []ProjectCoin{}
		}

		// update Balance here
		b, err := skyutil.GetBalance(coin.Name, coin.VotingAddress)
		if err != nil {

		} else {

			bp := wallet.BalancePair{}
			err := json.Unmarshal([]byte(b), &bp)
			if err == nil {
				balance, err := droplet.ToString(bp.Confirmed.Coins)
				if err != nil {
					// Log warning
				}

				coin.Balance, err = strconv.ParseFloat(balance, 64)
				if err != nil {
					// Log warning
				}

				UpdateBalance(coin.Balance, coin.Name)

			} else {
				// Log warining

			}
		}

		coins = append(coins, coin)
	}

	return coins
}

// AddProjectCoin inserts a record into database for a new coin
func AddProjectCoin(coin ProjectCoin) error {

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

// UpdateStatus change project coin status from Open to either Closed or Aborted
func UpdateStatus(s string, coinName string) error {

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

func GetProjectCoin(coinName string) ProjectCoin {
	coin := ProjectCoin{}
	row := dbSelectByNameStmt.QueryRow(coinName)
	if row != nil {
		err := row.Scan(
			&coin.Name, &coin.Symbol, &coin.NameCN,
			&coin.PlatformCoinName, &coin.PlatformCoinNameCN, &coin.PlatformCoinSymbol,
			&coin.Logo, &coin.VoteCap,
			&coin.Balance, &coin.BalanceCheckTime, // these two coulumns might not be necessary
			&coin.ShortDescription, &coin.LongDescription,
			&coin.Issuer,
			&coin.TimeOpening, &coin.TimeClosed, &coin.CloseReason,
			&coin.VotingAddress, &coin.Seed, &coin.PrivateKey,
			&coin.Status)

		if err != nil {
			// Log warning
		}
	} else {
		// Log warning
	}

	return coin
}
