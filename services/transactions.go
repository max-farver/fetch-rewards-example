package services

import (
	"context"
	"database/sql"
	"fetch-rewards/common"
	"log"
	"time"
)

type PointsService struct {
	Ctx context.Context
	DB  *sql.DB
}

// Add a new transaction to the database.
func (p *PointsService) Add(t common.Transaction) error {
	tx, err := p.DB.Begin()
	if err != nil {
		err = common.NoType.New(err.Error())
		return common.Wrap(err, "Internal Server Error")
	}

	stmt, err := tx.Prepare("INSERT INTO fetch_rewards(payer, remaining_points, timestamp, points) VALUES(?, ?, ?, ?)")
	if err != nil {
		err = common.NoType.New(err.Error())
		return common.Wrap(err, "Internal Server Error")
	}
	defer stmt.Close()

	_, err = stmt.Exec(t.Payer, t.Points, t.Timestamp, t.Points)
	if err != nil {
		err := common.NoType.New(err.Error())
		return common.Wrap(err, "Internal Server Error")
	}

	err = tx.Commit()
	if err != nil {
		err := common.NoType.New(err.Error())
		return common.Wrap(err, "Internal Server Error")
	}

	if t.Points < 0 {
		_, err = p.Spend(-1 * t.Points)
		if err != nil {
			return common.BadRequest.New("Balance cannot be negative")
		}
	}

	return nil
}

// Spend points.
func (p *PointsService) Spend(a int) ([]common.SpendingDetail, error) {
	remainingPointsInTransaction := a

	balances, err := p.Balance()
	newBalances := make(map[string]int)
	for k, v := range balances {
		newBalances[k] = v
	}

	if err != nil {
		return []common.SpendingDetail{}, err
	}

	rows, err := p.DB.Query(`SELECT id, payer, remaining_points FROM fetch_rewards WHERE remaining_points > 0 ORDER BY timestamp`)
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return []common.SpendingDetail{}, common.NoType.New("Internal Server Error")
	}

	var result []common.SpendingDetail
	updatedRemainingPoints := make(map[int]int)
	for rows.Next() {
		var id int
		var payer string
		var remainingPoints int
		err = rows.Scan(&id, &payer, &remainingPoints)
		if err != nil {
			err := common.NoType.New(err.Error())
			_ = common.Wrap(err, "Internal Server Error")
			return []common.SpendingDetail{}, common.NoType.New("Internal Server Error")
		}

		if remainingPointsInTransaction <= remainingPoints {
			newBalances[payer] = newBalances[payer] - remainingPointsInTransaction
			updatedRemainingPoints[id] = remainingPoints - remainingPointsInTransaction
			remainingPointsInTransaction = 0
		} else {
			remainingPointsInTransaction = remainingPointsInTransaction - remainingPoints
			newBalances[payer] = newBalances[payer] - remainingPoints
			updatedRemainingPoints[id] = 0
		}
	}
	rows.Close()

	if remainingPointsInTransaction > 0 {
		return []common.SpendingDetail{}, common.BadRequest.New("Insufficient points")
	}

	for k, v := range balances {
		if newBalances[k] < v {
			result = append(result, common.SpendingDetail{Payer: k, Points: newBalances[k] - v})
		}
	}

	err = p.insertSpendingTransactions(result)
	if err != nil {
		return []common.SpendingDetail{}, err
	}

	err = p.updatedRemainingPoints(updatedRemainingPoints)
	if err != nil {
		return []common.SpendingDetail{}, err
	}

	return result, nil
}

// Balance gets the current balance for every payer with a non-zero balance.
func (p *PointsService) Balance() (map[string]int, error) {
	rows, err := p.DB.Query("SELECT payer, SUM(remaining_points) FROM fetch_rewards WHERE remaining_points > 0 GROUP BY payer")
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

// insertSpendingTransactions creates transactions for the amount of points used from each payer.
func (p *PointsService) insertSpendingTransactions(sd []common.SpendingDetail) error {
	tx, err := p.DB.Begin()
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return common.NoType.New("Internal Server Error")
	}

	stmt, err := tx.Prepare("INSERT INTO fetch_rewards(payer, remaining_points, timestamp, points) VALUES(?, ?, ?, ?)")
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return common.NoType.New("Internal Server Error")
	}
	defer stmt.Close()

	for _, sd := range sd {
		_, err = stmt.Exec(sd.Payer, sd.Points, time.Now(), sd.Points)
		if err != nil {
			err := common.NoType.New(err.Error())
			_ = common.Wrap(err, "Internal Server Error")
			return common.NoType.New("Internal Server Error")
		}
	}

	err = tx.Commit()
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return common.NoType.New("Internal Server Error")
	}

	return nil
}

// updatedRemainingPoints udjusts the amount of points available for transactions.
func (p *PointsService) updatedRemainingPoints(updatedRemainingPoints map[int]int) error {
	tx, err := p.DB.Begin()
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return common.NoType.New("Internal Server Error")
	}

	stmt, err := tx.Prepare("UPDATE fetch_rewards SET remaining_points = ? WHERE id = ? ")
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return common.NoType.New("Internal Server Error")
	}
	defer stmt.Close()

	for id, remaining := range updatedRemainingPoints {
		_, err = stmt.Exec(remaining, id)
		if err != nil {
			err := common.NoType.New(err.Error())
			_ = common.Wrap(err, "Internal Server Error")
			return common.NoType.New("Internal Server Error")
		}
	}

	err = tx.Commit()
	if err != nil {
		err := common.NoType.New(err.Error())
		_ = common.Wrap(err, "Internal Server Error")
		return common.NoType.New("Internal Server Error")
	}

	return nil
}
