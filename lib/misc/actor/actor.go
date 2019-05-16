package actor

//Actor contains structural informations to build and work with an actor.
type Actor struct {
	running     bool
	identifier  int
	actionc     chan func()
	quitc       chan bool
	endCallback chan<- End
}

//End is send by endCallback to notify as to why an Actor ended.
//If called by quitc then it's without error; otherwise it'll be with error ;)
type End struct {
	ID        int
	WithError bool
}

func (a *Actor) loop() {
	for {
		select {
		case f := <-a.actionc:
			f()
		case <-a.quitc:
			break
		}
	}
	a.running = false
	a.endCallback <- End{a.identifier, false}
}

//New Create a new Actor
func New(id int, end chan<- End) *Actor {
	a := new(Actor)
	a.identifier = id
	a.running = false
	a.endCallback = end
	a.actionc = make(chan func())
	a.quitc = make(chan bool)
	return a
}

//Cast send and forget.
// If you want a reply, dont forget to provide your function a chan
// If you do so, DONT FORGET TO call defer close(<your chan>)
func (a *Actor) Cast(fn func()) {
	a.actionc <- fn
}

//Call send and wait for end of execution.
func (a *Actor) Call(fn func()) {
	exited := make(chan bool)
	defer close(exited)
	fn2 := func() {

		fn()
		exited <- true
	}

	a.actionc <- fn2

	<-exited
}

//ID of the Actor
func (a *Actor) ID() int {
	return a.identifier
}

//IsRunning Tell whether actor is running or not.
func (a *Actor) IsRunning() bool {
	return a.running
}

//Start run actor
func (a *Actor) Start() {
	a.running = true
	go a.loop()
}

//Stop actor
func (a *Actor) Stop() {
	a.quitc <- true
}
