package services

import (
	"context"
	"database/sql"
	"fetch-rewards/common"
	"log"
	"time"
)

type PointsService struct{
	Ctx context.Context
	DB  *sql.DB
}

func (p *PointsService) Add(t common.Transaction) error {
	tx, err := p.DB.Begin()
	if err != nil {
		err = common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
	}

	stmt, err := tx.Prepare("INSERT INTO fetch_rewards(payer, points, timestamp) VALUES(?, ?, ?)")
	if err != nil {
		err = common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.Payer, t.Points, t.Timestamp)
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
	}

	err = tx.Commit()
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
	}

	return nil
}

func (p *PointsService) Spend(a int) ([]common.SpendingDetail, error) {
	remainingPointsInTransaction := a

	balances, err := p.Balance()
	newBalances := make(map[string]int)
	for k,v := range balances {
		newBalances[k] = v
	}

	if err != nil {
		return []common.SpendingDetail{}, err
	}

	rows, err := p.DB.Query(`SELECT payer, points FROM fetch_rewards ORDER BY timestamp DESC`)
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return []common.SpendingDetail{}, common.NoType.New("Failed to retrieve balances")
	}

	var result []common.SpendingDetail
	for rows.Next() && remainingPointsInTransaction > 0 {
		var payer string
		var points int
		err = rows.Scan(&payer, &points)
		if err != nil {
			err := common.NoType.New(err.Error())
			_ = common.Wrap(err, "Internal Server Error")
			return []common.SpendingDetail{}, common.NoType.New("Failed to retrieve balances")
		}

		if remainingPointsInTransaction <= balances[payer] {
			newBalances[payer] -= remainingPointsInTransaction
			remainingPointsInTransaction = 0
		} else {
			remainingPointsInTransaction -= newBalances[payer]
			newBalances[payer] = 0
		}
	}
	rows.Close()

	for k, v := range balances {
		if newBalances[k] < v {
			result = append(result, common.SpendingDetail{Payer: k, Points: newBalances[k] - v})
		}
	}

	tx, err := p.DB.Begin()
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return []common.SpendingDetail{}, common.NoType.New("Failed to retrieve balances")
	}

	stmt, err := tx.Prepare("INSERT INTO fetch_rewards(payer, points, timestamp) VALUES(?, ?, ?)")
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return []common.SpendingDetail{}, common.NoType.New("Failed to retrieve balances")
	}
	defer stmt.Close()

	for _, sd := range result {
		_, err = stmt.Exec(sd.Payer, sd.Points, time.Now())
		if err != nil {
			err := common.NoType.New(err.Error())
			_ = common.Wrap(err, "Internal Server Error")
			return []common.SpendingDetail{}, common.NoType.New("Failed to retrieve balances")
		}
	}

	err = tx.Commit()
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return []common.SpendingDetail{}, common.NoType.New("Failed to retrieve balances")
	}

	return result, nil
}

func (p *PointsService) Balance() (map[string]int, error) {
	rows, err := p.DB.Query("SELECT payer, SUM(POINTS) FROM fetch_rewards GROUP BY payer")
	if err != nil {
		return map[string]int{}, common.NoType.New("Failed to retrieve balances")
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var payer string
		var points int
		err = rows.Scan(&payer, &points)
		if err != nil {
			log.Fatal(err)
		}
		result[payer] = points
	}

	return result, nil
}
