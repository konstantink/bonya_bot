package main

import (
	"log"
	"sort"
	"time"
)

type State interface {
	CheckTime(time time.Duration) bool
}

type TimeChecker struct {
	compareTime time.Duration
}

func (tc TimeChecker) CheckTime(time time.Duration) bool {
	if time > 0 && time <= tc.compareTime {
		log.Printf("%d <= %d, changing state to check\n", time, tc.compareTime)
		return true
	}
	//log.Printf("%d >= %d, changing state to check\n", time, tc.compareTime)
	return false
}

type Transition interface {
	Origin() State
	Exit() State
}

type T struct {
	O, E State
}

type Ts []T

func (t T) Origin() State {
	return t.O
}

func (t T) Exit() State {
	return t.E
}

func (t Ts) Len() int {
	return len(t)
}

func (t Ts) Less(i, j int) bool {
	return t[i].Origin().(TimeChecker).compareTime > t[j].Origin().(TimeChecker).compareTime
}

func (t Ts) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

type TransitionCheck func(stater Stater, goal State) bool

type Ruleset map[T]TransitionCheck

func (r Ruleset) AddRule(t T, check TransitionCheck) {
	r[t] = check
}

func (r *Ruleset) AddTransition(origin State, exit State) {
	var t T = T{origin, exit}
	r.AddRule(t, func(subject Stater, goal State) bool {
		return subject.CurrentState() == t.Origin()
	})
}

type Stater interface {
	CurrentState() State
	SetState(State)
	ResetState(time.Duration)
}

type Machine struct {
	rules   *Ruleset
	subject Stater
}

type LevelTimeCheckingMachine struct {
	state State

	machine *Machine
}

// CurrentState - returns the current state of machine
func (sfm *LevelTimeCheckingMachine) CurrentState() State {
	return sfm.state
}

// SetState - changes the state of machine to some state
func (sfm *LevelTimeCheckingMachine) SetState(state State) {
	sfm.state = state
}

// ResetState - resets the state of the machine according to the time value
// of new level. Formula to define state to reset machine to:
// levelTime - state.time >= state.time
func (fsm *LevelTimeCheckingMachine) ResetState(levelTime time.Duration) {
	var keys Ts = make(Ts, 0)
	for t := range *fsm.machine.rules {
		keys = append(keys, t)
	}
	sort.Sort(keys)
	for _, t := range keys {
		if (levelTime - t.Origin().(TimeChecker).compareTime) >= t.Origin().(TimeChecker).compareTime {
			fsm.SetState(t.Origin())
			log.Printf("New state: %.0f minute(s)\n", t.Origin().(TimeChecker).compareTime.Minutes())
			break
		}
	}
}

func (sfm *LevelTimeCheckingMachine) CheckTime(levelTime time.Duration) bool {
	if sfm.CurrentState().CheckTime(levelTime) {
		for t, _ := range *sfm.machine.rules {
			if sfm.CurrentState() == t.Origin() {
				sfm.SetState(t.Exit())
				log.Printf("New state: %d\n", t.Exit().(TimeChecker).compareTime)
				break
			}
		}
		return true
	}
	return false
}

func NewLevelTimeCheckingMachine(state State, rules *Ruleset) LevelTimeCheckingMachine {
	fsm := LevelTimeCheckingMachine{state, nil}
	machine := &Machine{rules: rules, subject: &fsm}
	fsm.machine = machine
	return fsm
}
