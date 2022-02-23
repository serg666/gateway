package mpi

import "log"

type Mpi interface {
	Check()
	Verify()
	Enrollment()
}

type InternalMpi struct {
	a int
}

func (im *InternalMpi) Check() {
	log.Print("some message")
}

func (im *InternalMpi) Verify() {
	log.Print("some message")
}

func (im *InternalMpi) Enrollment() {
	log.Print("some message")
}

func NewInternalMpi(a int) Mpi {
	return &InternalMpi{
		a: a,
	}
}
