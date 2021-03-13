package trunks

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"time"

	"github.com/go-co-op/gocron"
)

var Trunks *TrunksConfig

func runIPtables(args ...string) error {
	cmd := exec.Command("/sbin/iptables", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		log.Println("Error running /sbin/iptables:", err)
		return err
	}
	return nil
}

func runTC(args ...string) error {
	cmd := exec.Command("/sbin/tc", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		log.Println("Error running /sbin/tc:", err)
		return err
	}
	return nil
}

func FlushTables() error {
	log.Println("Flushing tables")
	err := runIPtables("-F", "-t", "mangle")
	runTC("qdisc", "del", "dev", Trunks.NIC.GW, "root")
	runTC("filter", "del", "dev", Trunks.NIC.GW)
	runTC("qdisc", "del", "dev", Trunks.NIC.ST, "root")
	runTC("filter", "del", "dev", Trunks.NIC.ST)
	return err
}

// Run the Trunk link
func Run(acm bool) {

	forward := fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Forward)))
	retun := fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Return)))
	delay := fmt.Sprintf("%dms", int64(math.Round(Trunks.Delay.Value/2)))
	offset := fmt.Sprintf("%dms", int64(math.Round(Trunks.Delay.Offset/2)))

	log.Println("Configure IPTABLES")
	runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", Trunks.NIC.ST, "-j", "MARK", "--set-mark", "10")
	runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", Trunks.NIC.GW, "-j", "MARK", "--set-mark", "20")

	log.Println("Configure TC")
	runTC("qdisc", "add", "dev", Trunks.NIC.GW, "root", "handle", "1:0", "htb", "default", "30")
	runTC("class", "add", "dev", Trunks.NIC.GW, "parent", "1:0", "classid", "1:1", "htb", "rate", retun)
	runTC("qdisc", "add", "dev", Trunks.NIC.GW, "parent", "1:1", "handle", "2:0", "netem", "delay", delay, offset, "distribution", "normal")
	runTC("filter", "add", "dev", Trunks.NIC.GW, "protocol", "ip", "parent", "1:0", "prio", "1", "handle", "10", "fw", "flowid", "1:1")

	runTC("qdisc", "add", "dev", Trunks.NIC.ST, "root", "handle", "1:0", "htb", "default", "30")
	runTC("class", "add", "dev", Trunks.NIC.ST, "parent", "1:0", "classid", "1:1", "htb", "rate", forward)
	runTC("qdisc", "add", "dev", Trunks.NIC.ST, "parent", "1:1", "handle", "2:0", "netem", "delay", delay, offset, "distribution", "normal")
	runTC("filter", "add", "dev", Trunks.NIC.ST, "protocol", "ip", "parent", "1:0", "prio", "1", "handle", "20", "fw", "flowid", "1:1")

	if acm {
		log.Println("ACM activated")
		scheduler := gocron.NewScheduler(time.UTC)
		scheduler.Every(1).Seconds().Do(RunACM)
		scheduler.StartBlocking()
	}

}
