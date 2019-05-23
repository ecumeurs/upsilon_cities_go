//Package actor Actors protect a ressource
//There are a few rules:
//  * An Actor may call on another through a CAST and won't wait for completion (means: no channel !)
//  * An Actor may call on another through a CALL (expecting a reply through channel)  ONLY IF dev is SURE 200% that this actor won't ever be used to poll on other actors ...
//    * Otherwise, dev may face deadlock.
//  * Dont try to alter ressource protected by the actor outside either CALL OR CAST
//  * CALL is a cast with an implicit channel waiting for completion of the function ... dont be abused by it.
package actor

//Actor contains structural informations to build and work with an actor.
type Actor struct {
	Running     bool
	Identifier  int
	Actionc     chan func()
	Quitc       chan bool
	EndCallback chan<- End
	Loop        func()
}

//End is send by endCallback to notify as to why an Actor ended.
//If called by quitc then it's without error; otherwise it'll be with error ;)
type End struct {
	ID        int
	WithError bool
}

func (a *Actor) loop() {
	a.Loop()
}

//New Create a new Actor
func New(id int, end chan<- End) *Actor {
	a := new(Actor)
	a.Identifier = id
	a.Running = false
	a.EndCallback = end
	a.Actionc = make(chan func())
	a.Quitc = make(chan bool)
	a.Loop = func() {
		for {
			select {
			case f := <-a.Actionc:
				f()
			case <-a.Quitc:
				break
			}
		}
		a.Running = false
		a.EndCallback <- End{a.Identifier, false}
	}
	return a
}

//Cast send and forget.
// If you want a reply, dont forget to provide your function a chan
// If you do so, DONT FORGET TO call defer close(<your chan>)
func (a *Actor) Cast(fn func()) {
	a.Actionc <- fn
}

//Call send and wait for end of execution.
func (a *Actor) Call(fn func()) {
	exited := make(chan bool)
	defer close(exited)
	fn2 := func() {

		fn()
		exited <- true
	}

	a.Actionc <- fn2

	<-exited
}

//ID of the Actor
func (a *Actor) ID() int {
	return a.Identifier
}

//IsRunning Tell whether actor is running or not.
func (a *Actor) IsRunning() bool {
	return a.Running
}

//Start run actor
func (a *Actor) Start() {
	a.Running = true
	go a.loop()
}

//Stop actor
func (a *Actor) Stop() {
	a.Quitc <- true
}
