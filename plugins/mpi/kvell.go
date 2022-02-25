package mpi

import "log"

type Mpi interface {
	Check()
	Verify()
}

type KvellMpi struct {
	a int
}

func (km *KvellMpi) Check() {
	log.Print("some message")
}

func (km *KvellMpi) Verify() {
	log.Print("some message")
}


func NewKvellMpi(a int) Mpi {
	return &KvellMpi{
		a: a,
	}
}
