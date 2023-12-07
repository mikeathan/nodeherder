package automations

import (
	"node-herder/models/devices"
)

type Evaluator interface {
	Evaluate(exposes map[string]*devices.Entity) bool
}

// default evaluator checks all cinditions with passed existing data/payload
type DefaultEvaluator struct {
}

//sequencial evluator dont haveall data, the will be coming later on

// is temp == 10

// is presence == true
// is light > 40

// is btn1 == pressed, store pressed
// is btn1 == released

// is btn1 == pressed, true, store pressed
// is btn1 == hold , false but dont clear previous state. we domt care for case 1
// is btn1 == released, true, clear state

// not possible
// is btn1 == pressed, store pressed
// is btn2 == pressed ,false but  reset  state
// is btn1 == released
