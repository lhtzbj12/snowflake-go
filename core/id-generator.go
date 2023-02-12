package core

type IdGenerator interface {
	Init()
	GetId() (int64, error)
	GetIds(n int) ([]int64, error)
}
