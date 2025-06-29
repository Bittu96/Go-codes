package main

import "time"

func main() {
	// service 1
	service1 := New(2*time.Second, 3)

	// service 2
	// service2 := New(5*time.Second, 10)

	for i := 0; i < 10; i++ {
		// send request
		req := "sample msg"
		service1.LB(req)

		//
		// service2.Call(req)
	}
}

var (
	LBState                 = "closed"
	LBFailureStreak         = 0
	LBRecoverySuccessStreak = 0
	LBTimeout               = 2 * time.Second
	restrictMode            bool
	restrictModeStart       time.Time
)

func LB(req string) {
	if LBFailureStreak > 3 {
		LBState = "open"
		restrictMode = true
		restrictModeStart = time.Now()
		LBFailureStreak = 0
	}

	if LBState == "closed" {
		//make req
		if err := serviceCall(); err != nil {
			LBFailureStreak++
		} else {
			LBFailureStreak = 0
		}
	} else if LBState == "open" {
		// block
		if isRestricted() {
			return
		} else {
			LBFailureStreak = 0
			LBState = "half-open"
		}

	} else if LBState == "half-open" {
		if err := serviceCall(); err != nil {
			LBState = "open"
			LBFailureStreak = 1
		} else {
			LBRecoverySuccessStreak++
			if LBRecoverySuccessStreak >= 2 {
				LBState = "closed"
				LBFailureStreak = 0
			}
		}
	}
}

func serviceCall() error {
	var err error

	// logic

	return err
}

func isRestricted() bool {
	//
	return false
}

type Service struct {
	LBState                 string
	MaxFailureTolarance     int
	LBFailureStreak         int
	LBRecoverySuccessStreak int
	LBTimeout               time.Duration
	restrictMode            bool
	restrictModeStart       time.Time
}

func New(timeOut time.Duration, maxFailureTolarance int) Service {
	return Service{
		LBState:                 "closed",
		MaxFailureTolarance:     maxFailureTolarance,
		LBFailureStreak:         0,
		LBRecoverySuccessStreak: 0,
		LBTimeout:               timeOut,
	}
}

func (s Service) LB(req string) {
	if LBFailureStreak > s.MaxFailureTolarance {
		LBState = "open"
		restrictMode = true
		restrictModeStart = time.Now()
		LBFailureStreak = 0
	}

	if LBState == "closed" {
		//make req
		if err := s.Call(); err != nil {
			LBFailureStreak++
		} else {
			LBFailureStreak = 0
		}
	} else if LBState == "open" {
		// block
		if isRestricted() {
			return
		} else {
			LBFailureStreak = 0
			LBState = "half-open"
		}

	} else if LBState == "half-open" {
		if err := s.Call(); err != nil {
			LBState = "open"
			LBFailureStreak = 1
		} else {
			LBRecoverySuccessStreak++
			if LBRecoverySuccessStreak >= s.LBRecoverySuccessStreak {
				LBState = "closed"
				LBFailureStreak = 0
			}
		}
	}
}

func (s Service) Call(req string) error {
	var err error

	// logic

	return err
}

// type LB interface {

// }

// func (s Service) successRecoveryMethod() {

// }
