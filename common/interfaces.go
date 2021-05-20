package common

type PointsService interface {
	Add(t Transaction) error
	Spend(a int) ([]SpendingDetail, error)
	Balance() (map[string]int, error)
}
