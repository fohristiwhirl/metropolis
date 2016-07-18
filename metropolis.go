package main

// This is a framework for a metropolis-coupled search. In this example, the program searches
// for a number that has zero difference from 500. It does this by starting at zero and mutating
// the state, either accepting the new state or not. Multiple threads are run in parallel to pull
// the state out of local optima (not a problem in this example, meh).
//
// The point is, this approach can work for more complicated things.

import (
    "bufio"
    "fmt"
    "math/rand"
    "os"
    "time"
)

// Thread parameters...

const THREADS = 6

var Heat =  [...]float64{
                    0,
                    0.00001,
                    0.0001,
                    0.0005,
                    0.001,
                    0.01,
                    }

// Each thread will indicate it's finished its current iteration by sending a pointer to the current state.
// The hub will tell the thread to resume by sending it a pointer to the state it will work on next iteration.

var ReportChan [THREADS]chan *State
var ResumeChan [THREADS]chan *State

type State struct {
    Score int32
    World int32
}

// ----------------------------------------------------------------- INIT and MAIN functions called at startup

func init() {
    rand.Seed(time.Now().UTC().UnixNano())

    for n := 0 ; n < THREADS ; n++ {
        ReportChan[n] = make(chan *State)
        ResumeChan[n] = make(chan *State)
    }
}

func main() {
    for n := 0 ; n < THREADS ; n++ {
        go chain(n)
        ResumeChan[n] <- NewState()
    }

    hub()

    // Once we find a solution, wait for user input (prevent window from closing)...

    reader := bufio.NewReader(os.Stdin)
    reader.ReadString('\n')
}

// ----------------------------------------------------------------- METHODS for dealing with a State

func (s *State) SetScore() {   // As a convention, lets say lower is better, i.e. 0 is best possible score

    // In this example, the score is the distance to the number 500
    s.Score = 500 - s.World
    if s.Score < 0 {
        s.Score *= -1
    }
}

func (s *State) Mutate() {
    s.World += rand.Int31n(100000) - 50000
    return
}

func (s *State) Dump() {
    fmt.Printf("World: %d (score: %d)\n", s.World, s.Score)
}

func NewState() *State {
    var s *State = new(State)
    s.SetScore()                    // IMPORTANT!
    return s
}

// ----------------------------------------------------------------- HUB (controller)

func hub() {

    var state_pointers [THREADS]*State

    for {

        for n := 0; n < THREADS; n++ {
            state_pointers[n] = <- ReportChan[n]
            if state_pointers[n].Score == 0 {
                fmt.Printf("Success in thread %d: ", n)
                state_pointers[n].Dump()
                return
            }
        }

        for n := 0; n < THREADS - 1; n++ {
            if state_pointers[n].Score > state_pointers[n + 1].Score {
                state_pointers[n], state_pointers[n + 1] = state_pointers[n + 1], state_pointers[n]
            }
        }

        for n := 0; n < THREADS; n++ {
            ResumeChan[n] <- state_pointers[n]
        }
    }
}

func chain(index int) {

    var my_state *State

    for {
        my_state = <- ResumeChan[index]

        old_state := *my_state      // Keep a copy of the initial state in case we don't accept the mutation

        my_state.Mutate()
        my_state.SetScore()

        // Revert to old state?
        if my_state.Score > old_state.Score && rand.Float64() > Heat[index] {
            *my_state = old_state
        }

        ReportChan[index] <- my_state
    }
}
