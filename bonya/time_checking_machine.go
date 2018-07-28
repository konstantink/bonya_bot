package main

import (
	"fmt"
	"log"
	"sort"
	"time"
)

// State interface for the structures that represent separate state of the FSM
type State interface {
	Prepare()
	// CheckTime(time time.Duration) bool
	Process(args ...interface{}) bool
}

// TimeChecker structure that represents the state of the finite machine.
// It implements interface State
type TimeChecker struct {
	compareTime time.Duration
}

func (tc TimeChecker) String() string {
	return fmt.Sprintf("%d", tc.compareTime)
}

// Prepare function to execute before main processing.
// TimeChecker does not use this function, but have to implement it to
// satisfy State interface
func (tc TimeChecker) Prepare() {}

// Process main function for the State. Performs main logic for the state
func (tc TimeChecker) Process(args ...interface{}) bool {
	var time = args[0].(time.Duration)
	if time > 0 && time <= tc.compareTime {
		log.Printf("%d <= %d, changing state to check\n", time, tc.compareTime)
		return true
	}
	//log.Printf("%d >= %d, changing state to check\n", time, tc.compareTime)
	return false
}

// Transition interface for the rules that defines relation between states
type Transition interface {
	Origin() State
	Exit() State
}

// T structure that implmenets Transition interface
type T struct {
	O, E State
}

// Ts list of Transitions
type Ts []T

// Origin implements method of Transition interface, returns the input state
// of the Transition
func (t T) Origin() State {
	return t.O
}

// Exit implements method of Transition interface, returns the output state
// of the Transition
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

// TransitionCheck function type that defines the rule to transform from the input state
// to the output state
type TransitionCheck func(stater Stater, goal State) bool

// Ruleset map of the rules for Transitions
type Ruleset map[T]TransitionCheck

// AddRule adds/updates Transition with function to check it
func (r Ruleset) AddRule(t T, check TransitionCheck) {
	r[t] = check
}

// AddTransition creates new rule for the transition. Rule is simple - just checks
// that there is any rule to make a transition from this state
func (r *Ruleset) AddTransition(origin State, exit State) {
	var t = T{origin, exit}
	r.AddRule(t, func(subject Stater, goal State) bool {
		return subject.CurrentState() == t.Origin()
	})
}

type Stater interface {
	CurrentState() State
	SetState(State)
	ResetState(args ...interface{})
}

type Machine struct {
	rules   *Ruleset
	subject Stater
}

type CheckingMachine struct {
	state State

	machine *Machine
}

// CurrentState - returns the current state of machine
func (fsm *CheckingMachine) CurrentState() State {
	return fsm.state
}

// SetState - changes the state of machine to some state
func (fsm *CheckingMachine) SetState(state State) {
	fsm.state = state
}

type LevelTimeCheckingMachine struct {
	CheckingMachine
}

// ResetState - resets the state of the machine according to the time value
// of new level. Formula to define state to reset machine to:
// levelTime - state.time >= state.time
func (fsm *LevelTimeCheckingMachine) ResetState(args ...interface{}) {
	var (
		keys      = make(Ts, 0)
		levelTime = args[0].(time.Duration)
	)
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

// Process
func (fsm *LevelTimeCheckingMachine) Process(levelTime time.Duration) bool {
	var state = fsm.CurrentState()
	if state.Process(levelTime) {
		for t := range *fsm.machine.rules {
			if state == t.Origin() {
				fsm.SetState(t.Exit())
				// log.Printf("New state: %d\n", t.Exit().(TimeChecker).compareTime)
				log.Printf("New state: %s\n", t.Exit())
				break
			}
		}
		return true
	}
	return false
}

// NewLevelTimeCheckingMachine creates new instance of LevelTimeCheckingMachine
func NewLevelTimeCheckingMachine(state State, rules *Ruleset) LevelTimeCheckingMachine {
	fsm := LevelTimeCheckingMachine{CheckingMachine{state, nil}}
	machine := &Machine{rules: rules, subject: &fsm}
	fsm.machine = machine
	return fsm
}
