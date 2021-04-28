package trunks

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
)

// RunACM simulate the used by the DVB-S2/RCS2 system
func RunACM(qos bool) {
	if Trunks.CurrentACM == nil || Trunks.ACMCounter >= Trunks.CurrentACM.Duration {
		log.Println("Changing link capacity")
		var l = len(Trunks.ACMList)
		rand.Seed(time.Now().UnixNano())
		var index int = rand.Intn(l)
		Trunks.CurrentACM = Trunks.ACMList[index]

		var forward string
		var retun string

		if qos {
			forward = fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Forward*Trunks.CurrentACM.Weight))-1)
			retun = fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Return*Trunks.CurrentACM.Weight))-1)
		} else {
			forward = fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Forward*Trunks.CurrentACM.Weight)))
			retun = fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Return*Trunks.CurrentACM.Weight)))
		}
		log.Println("Setting the forward link bandwidth at", forward)
		log.Println("Setting the return link bandwidth at", retun)
		runTC("class", "change", "dev", Trunks.NIC.GW, "parent", "1:0", "classid", "1:1", "htb", "rate", retun)
		runTC("class", "change", "dev", Trunks.NIC.ST, "parent", "1:0", "classid", "1:1", "htb", "rate", forward)
		Trunks.ACMCounter = 0
	}
	Trunks.ACMCounter++
}
