// Package trunks define the trunk runtime
package trunks

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

// RunACM simulate the used by the DVB-S2/RCS2 system
func (t *TrunksConfig) RunACM(qos bool) {
	if t.CurrentACM == nil || t.ACMCounter >= t.CurrentACM.Duration {
		log.Println("Changing link capacity")
		var l = len(t.ACMList)
		rand.Seed(time.Now().UnixNano())
		var index int = rand.Intn(l)
		t.CurrentACM = t.ACMList[index]

		var forward string
		var retun string

		if qos {
			forward = fmt.Sprintf("%dmbit", int64(math.Round(t.Bandwidth.Forward*t.CurrentACM.Weight))-1)
			retun = fmt.Sprintf("%dmbit", int64(math.Round(t.Bandwidth.Return*t.CurrentACM.Weight))-1)
			runTC("class", "change", "dev", t.NIC.GW, "parent", "1:0", "classid", "1:20", "htb", "rate", retun, "prio", "1")
			runTC("class", "change", "dev", t.NIC.ST, "parent", "1:0", "classid", "1:20", "htb", "rate", forward, "prio", "1")

		} else {
			forward = fmt.Sprintf("%dmbit", int64(math.Round(t.Bandwidth.Forward*t.CurrentACM.Weight)))
			retun = fmt.Sprintf("%dmbit", int64(math.Round(t.Bandwidth.Return*t.CurrentACM.Weight)))
			runTC("class", "change", "dev", t.NIC.GW, "parent", "1:0", "classid", "1:1", "htb", "rate", retun)
			runTC("class", "change", "dev", t.NIC.ST, "parent", "1:0", "classid", "1:1", "htb", "rate", forward)
		}
		log.Println("Setting the forward link bandwidth at", forward)
		log.Println("Setting the return link bandwidth at", retun)

		t.ACMCounter = 0
	}
	t.ACMCounter++
}
