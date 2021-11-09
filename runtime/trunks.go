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
	cmd := exec.Command("iptables", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		errLog := fmt.Sprintf("Error running %s: %s", cmd.Args[0], err)
		log.Println(errLog)
		return err
	}
	return nil
}

func runTC(args ...string) error {
	cmd := exec.Command("tc", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		errLog := fmt.Sprintf("Error running %s: %s", cmd.Args[0], err)
		log.Println(errLog)
		return err
	}
	return nil
}

func runSYSCTL(args ...string) error {
	cmd := exec.Command("sysctl", args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if nil != err {
		errLog := fmt.Sprintf("Error running %s: %s", cmd.Args[0], err)
		log.Println(errLog)
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

	runSYSCTL("net.ipv4.ip_forward=1")

	if !Trunks.QoS {
		log.Println("Running without QoS")

		forward := fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Forward)))
		retun := fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Return)))
		delay := fmt.Sprintf("%dms", int64(math.Round(Trunks.Delay.Value/2)))
		offset := fmt.Sprintf("%dms", int64(math.Round(Trunks.Delay.Offset/2)))
		jitter := Trunks.Delay.Offset > 1

		// qlen formula: 1.5 * bandwidth[bits/s] * latency[s] / mtu[bits]
		qlenForward := fmt.Sprintf("%d", int64(math.Round(1.5 * (Trunks.Bandwidth.Forward * 1000000) * (Trunks.Delay.Value / (2 * 1000)) / (8 * 1500))))
		qlenReturn := fmt.Sprintf("%d", int64(math.Round(1.5 * (Trunks.Bandwidth.Return * 1000000) * (Trunks.Delay.Value / (2 * 1000)) / (8 * 1500))))

		log.Println("Configure IPTABLES")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", Trunks.NIC.ST, "-j", "MARK", "--set-mark", "10")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", Trunks.NIC.GW, "-j", "MARK", "--set-mark", "20")

		log.Println("Configure TC")
		runTC("qdisc", "add", "dev", Trunks.NIC.GW, "root", "handle", "1:0", "htb", "default", "30")
		runTC("class", "add", "dev", Trunks.NIC.GW, "parent", "1:0", "classid", "1:1", "htb", "rate", retun, "burst", "30k", "cburst", "30k")
		if jitter {
			runTC("qdisc", "add", "dev", Trunks.NIC.GW, "parent", "1:1", "handle", "2:0", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenReturn)
		} else {
			runTC("qdisc", "add", "dev", Trunks.NIC.GW, "parent", "1:1", "handle", "2:0", "netem", "delay", delay, "limit", qlenReturn)
		}
		runTC("filter", "add", "dev", Trunks.NIC.GW, "protocol", "ip", "parent", "1:0", "prio", "1", "handle", "10", "fw", "flowid", "1:1")

		runTC("qdisc", "add", "dev", Trunks.NIC.ST, "root", "handle", "1:0", "htb", "default", "30")
		runTC("class", "add", "dev", Trunks.NIC.ST, "parent", "1:0", "classid", "1:1", "htb", "rate", forward, "burst", "30k", "cburst", "30k")
		if jitter {
			runTC("qdisc", "add", "dev", Trunks.NIC.ST, "parent", "1:1", "handle", "2:0", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenForward)
		} else {
			runTC("qdisc", "add", "dev", Trunks.NIC.ST, "parent", "1:1", "handle", "2:0", "netem", "delay", delay, "limit", qlenForward)
		}
		runTC("filter", "add", "dev", Trunks.NIC.ST, "protocol", "ip", "parent", "1:0", "prio", "1", "handle", "20", "fw", "flowid", "1:1")

	} else {

		log.Println("Running with QoS")

		forward := fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Forward)))
		forwardVoIP := fmt.Sprintf("%dmbit", 2)
		forwardRest := fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Forward))-1)
		retun := fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Return)))
		returnVoIP := fmt.Sprintf("%dmbit", 2)
		returnRest := fmt.Sprintf("%dmbit", int64(math.Round(Trunks.Bandwidth.Return))-1)
		delay := fmt.Sprintf("%dms", int64(math.Round(Trunks.Delay.Value/2)))
		offset := fmt.Sprintf("%dms", int64(math.Round(Trunks.Delay.Offset/2)))
		jitter := Trunks.Delay.Offset > 1

		// qlen formula: 1.5 * bandwidth[bits/s] * latency[s] / mtu[bits]
		qlenForwardVoIP := fmt.Sprintf("%d", int64(math.Round(1.5 * (2 * 1000000) * (Trunks.Delay.Value / (2 * 1000)) / (8 * 1500))))
		qlenForwardRest := fmt.Sprintf("%d", int64(math.Round(1.5 * ((Trunks.Bandwidth.Forward - 1) * 1000000) * (Trunks.Delay.Value / (2 * 1000)) / (8 * 1500))))
		qlenReturnVoIP := fmt.Sprintf("%d", int64(math.Round(1.5 * (Trunks.Bandwidth.Return * 1000000) * (Trunks.Delay.Value / (2 * 1000)) / (8 * 1500))))
		qlenReturnRest := fmt.Sprintf("%d", int64(math.Round(1.5 * (Trunks.Bandwidth.Return * 1000000) * (Trunks.Delay.Value / (2 * 1000)) / (8 * 1500))))

		log.Println("Configure IPTABLES")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", Trunks.NIC.ST, "-j", "MARK", "--set-mark", "10")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", Trunks.NIC.ST, "-m", "dscp", "--dscp", "0x2c", "-j", "MARK", "--set-mark", "11")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", Trunks.NIC.ST, "-m", "dscp", "--dscp", "0x2e", "-j", "MARK", "--set-mark", "11")

		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", Trunks.NIC.GW, "-j", "MARK", "--set-mark", "20")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", Trunks.NIC.GW, "-m", "dscp", "--dscp", "0x2c", "-j", "MARK", "--set-mark", "21")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", Trunks.NIC.GW, "-m", "dscp", "--dscp", "0x2e", "-j", "MARK", "--set-mark", "21")

		log.Println("Configure TC")

		// Qdisc configuration
		runTC("qdisc", "add", "dev", Trunks.NIC.GW, "root", "handle", "1:0", "htb", "default", "20")
		runTC("class", "add", "dev", Trunks.NIC.GW, "parent", "1:0", "classid", "1:1", "htb", "rate", retun, "burst", "30k", "cburst", "30k")
		runTC("class", "add", "dev", Trunks.NIC.GW, "parent", "1:1", "classid", "1:10", "htb", "rate", returnVoIP, "prio", "0", "burst", "3k", "cburst", "3k")
		if jitter {
			runTC("qdisc", "add", "dev", Trunks.NIC.GW, "parent", "1:10", "handle", "110:", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenReturnVoIP)
		} else {
			runTC("qdisc", "add", "dev", Trunks.NIC.GW, "parent", "1:10", "handle", "110:", "netem", "delay", delay, "limit", qlenReturnVoIP)
		}
		runTC("class", "add", "dev", Trunks.NIC.GW, "parent", "1:1", "classid", "1:20", "htb", "rate", returnRest, "prio", "1", "burst", "30k", "cburst", "30k")
		if jitter {
			runTC("qdisc", "add", "dev", Trunks.NIC.GW, "parent", "1:20", "handle", "120:", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenReturnRest)
		} else {
			runTC("qdisc", "add", "dev", Trunks.NIC.GW, "parent", "1:20", "handle", "120:", "netem", "delay", delay, "limit", qlenReturnRest)
		}
		// Filters
		runTC("filter", "add", "dev", Trunks.NIC.GW, "protocol", "ip", "parent", "1:0", "prio", "0", "handle", "11", "fw", "flowid", "1:10")
		runTC("filter", "add", "dev", Trunks.NIC.GW, "protocol", "ip", "parent", "1:0", "prio", "1", "handle", "10", "fw", "flowid", "1:20")

		// Qdisc configuration
		runTC("qdisc", "add", "dev", Trunks.NIC.ST, "root", "handle", "1:0", "htb", "default", "20")
		runTC("class", "add", "dev", Trunks.NIC.ST, "parent", "1:0", "classid", "1:1", "htb", "rate", forward, "burst", "30k", "cburst", "30k")
		runTC("class", "add", "dev", Trunks.NIC.ST, "parent", "1:0", "classid", "1:10", "htb", "rate", forwardVoIP, "prio", "0", "burst", "3k", "cburst", "3k")
		if jitter {
			runTC("qdisc", "add", "dev", Trunks.NIC.ST, "parent", "1:10", "handle", "110:", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenForwardVoIP)
		} else {
			runTC("qdisc", "add", "dev", Trunks.NIC.ST, "parent", "1:10", "handle", "110:", "netem", "delay", delay, "limit", qlenForwardVoIP)
		}
		runTC("class", "add", "dev", Trunks.NIC.ST, "parent", "1:0", "classid", "1:20", "htb", "rate", forwardRest, "prio", "1", "burst", "30k", "cburst", "30k")
		if jitter {
			runTC("qdisc", "add", "dev", Trunks.NIC.ST, "parent", "1:20", "handle", "120:", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenForwardRest)
		} else {
			runTC("qdisc", "add", "dev", Trunks.NIC.ST, "parent", "1:20", "handle", "120:", "netem", "delay", "limit", qlenForwardRest)
		}


		// Filters
		runTC("filter", "add", "dev", Trunks.NIC.ST, "protocol", "ip", "parent", "1:0", "prio", "0", "handle", "21", "fw", "flowid", "1:10")
		runTC("filter", "add", "dev", Trunks.NIC.ST, "protocol", "ip", "parent", "1:0", "prio", "1", "handle", "20", "fw", "flowid", "1:20")
	}

	if acm {
		log.Println("Starting Trunks with ACM")
		scheduler := gocron.NewScheduler(time.UTC)
		scheduler.Every(1).Seconds().Do(RunACM, Trunks.QoS)
		scheduler.StartBlocking()
	} else {
		log.Println("Trunks started without ACM")
		time.Sleep(time.Duration(1<<63 - 1))
	}

}
