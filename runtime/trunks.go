package trunks

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/go-co-op/gocron"
)

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

func (t *TrunksConfig) FlushTables() error {
	log.Println("Flushing tables")
	err := runIPtables("-F", "-t", "mangle")
	runTC("filter", "del", "dev", t.NIC.GW)
	runTC("qdisc", "del", "dev", t.NIC.GW, "root")
	runTC("filter", "del", "dev", t.NIC.ST)
	runTC("qdisc", "del", "dev", t.NIC.ST, "root")
	return err
}

func (t *TrunksConfig) isKernelVersionBugged() bool {
	if t.KernelVersionCheck {
		return false
	}
	// See https://github.com/shynuu/trunks/issues/6
	cmd := exec.Command("uname", "--kernel-version")
	out, err := cmd.Output()
	if err != nil {
		errLog := fmt.Sprintf("Error running %s: %s", cmd.Args[0], err)
		log.Println(errLog)
		return false
	}
	// See https://gist.github.com/louisroyer/90636c07dc4b205b813a56de718d9d09
	if strings.Contains(string(out), "Debian 6.1.38-") {
		log.Println("Warning: offset delay will be disabled because you are using Debian with Linux 6.1.38 which is known to crash with this settting. See https://github.com/shynuu/trunks/issues/6 for details.")
		return true
	}
	return false
}

// Run the Trunk link
func (t *TrunksConfig) Run() {

	runSYSCTL("net.ipv4.ip_forward=1")

	if !t.QoS {
		log.Println("Running without QoS")

		forward := fmt.Sprintf("%dmbit", int64(math.Round(t.Bandwidth.Forward)))
		retun := fmt.Sprintf("%dmbit", int64(math.Round(t.Bandwidth.Return)))
		delay := fmt.Sprintf("%dms", int64(math.Round(t.Delay.Value/2)))
		offset := fmt.Sprintf("%dms", int64(math.Round(t.Delay.Offset/2)))
		jitter := t.Delay.Offset > 1 && t.isKernelVersionBugged()

		// qlen formula: 1.5 * bandwidth[bits/s] * latency[s] / mtu[bits]
		qlenForward := fmt.Sprintf("%d", int64(math.Round(1.5*(t.Bandwidth.Forward*1000000)*(t.Delay.Value/(2*1000))/(8*1500))))
		qlenReturn := fmt.Sprintf("%d", int64(math.Round(1.5*(t.Bandwidth.Return*1000000)*(t.Delay.Value/(2*1000))/(8*1500))))

		log.Println("Configure IPTABLES")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", t.NIC.ST, "-j", "MARK", "--set-mark", "10")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", t.NIC.GW, "-j", "MARK", "--set-mark", "20")

		log.Println("Configure TC")
		runTC("qdisc", "add", "dev", t.NIC.GW, "root", "handle", "1:0", "htb", "default", "30")
		runTC("class", "add", "dev", t.NIC.GW, "parent", "1:0", "classid", "1:1", "htb", "rate", retun, "burst", "30k", "cburst", "30k")
		if jitter {
			runTC("qdisc", "add", "dev", t.NIC.GW, "parent", "1:1", "handle", "2:0", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenReturn)
		} else {
			runTC("qdisc", "add", "dev", t.NIC.GW, "parent", "1:1", "handle", "2:0", "netem", "delay", delay, "limit", qlenReturn)
		}
		runTC("filter", "add", "dev", t.NIC.GW, "protocol", "ip", "parent", "1:0", "prio", "1", "handle", "10", "fw", "flowid", "1:1")

		runTC("qdisc", "add", "dev", t.NIC.ST, "root", "handle", "1:0", "htb", "default", "30")
		runTC("class", "add", "dev", t.NIC.ST, "parent", "1:0", "classid", "1:1", "htb", "rate", forward, "burst", "30k", "cburst", "30k")
		if jitter {
			runTC("qdisc", "add", "dev", t.NIC.ST, "parent", "1:1", "handle", "2:0", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenForward)
		} else {
			runTC("qdisc", "add", "dev", t.NIC.ST, "parent", "1:1", "handle", "2:0", "netem", "delay", delay, "limit", qlenForward)
		}
		runTC("filter", "add", "dev", t.NIC.ST, "protocol", "ip", "parent", "1:0", "prio", "1", "handle", "20", "fw", "flowid", "1:1")

	} else {

		log.Println("Running with QoS")

		forward := fmt.Sprintf("%dmbit", int64(math.Round(t.Bandwidth.Forward)))
		forwardVoIP := fmt.Sprintf("%dmbit", 2)
		forwardRest := fmt.Sprintf("%dmbit", int64(math.Round(t.Bandwidth.Forward))-1)
		retun := fmt.Sprintf("%dmbit", int64(math.Round(t.Bandwidth.Return)))
		returnVoIP := fmt.Sprintf("%dmbit", 2)
		returnRest := fmt.Sprintf("%dmbit", int64(math.Round(t.Bandwidth.Return))-1)
		delay := fmt.Sprintf("%dms", int64(math.Round(t.Delay.Value/2)))
		offset := fmt.Sprintf("%dms", int64(math.Round(t.Delay.Offset/2)))
		jitter := t.Delay.Offset > 1 && t.isKernelVersionBugged()

		// qlen formula: 1.5 * bandwidth[bits/s] * latency[s] / mtu[bits]
		qlenForwardVoIP := fmt.Sprintf("%d", int64(math.Round(1.5*(2*1000000)*(t.Delay.Value/(2*1000))/(8*1500))))
		qlenForwardRest := fmt.Sprintf("%d", int64(math.Round(1.5*((t.Bandwidth.Forward-1)*1000000)*(t.Delay.Value/(2*1000))/(8*1500))))
		qlenReturnVoIP := fmt.Sprintf("%d", int64(math.Round(1.5*(t.Bandwidth.Return*1000000)*(t.Delay.Value/(2*1000))/(8*1500))))
		qlenReturnRest := fmt.Sprintf("%d", int64(math.Round(1.5*(t.Bandwidth.Return*1000000)*(t.Delay.Value/(2*1000))/(8*1500))))

		log.Println("Configure IPTABLES")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", t.NIC.ST, "-j", "MARK", "--set-mark", "10")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", t.NIC.ST, "-m", "dscp", "--dscp", "0x2c", "-j", "MARK", "--set-mark", "11")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", t.NIC.ST, "-m", "dscp", "--dscp", "0x2e", "-j", "MARK", "--set-mark", "11")

		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", t.NIC.GW, "-j", "MARK", "--set-mark", "20")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", t.NIC.GW, "-m", "dscp", "--dscp", "0x2c", "-j", "MARK", "--set-mark", "21")
		runIPtables("-t", "mangle", "-A", "PREROUTING", "-i", t.NIC.GW, "-m", "dscp", "--dscp", "0x2e", "-j", "MARK", "--set-mark", "21")

		log.Println("Configure TC")

		// Qdisc configuration
		runTC("qdisc", "add", "dev", t.NIC.GW, "root", "handle", "1:0", "htb", "default", "20")
		runTC("class", "add", "dev", t.NIC.GW, "parent", "1:0", "classid", "1:1", "htb", "rate", retun, "burst", "30k", "cburst", "30k")
		runTC("class", "add", "dev", t.NIC.GW, "parent", "1:1", "classid", "1:10", "htb", "rate", returnVoIP, "prio", "0", "burst", "3k", "cburst", "3k")
		if jitter {
			runTC("qdisc", "add", "dev", t.NIC.GW, "parent", "1:10", "handle", "110:", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenReturnVoIP)
		} else {
			runTC("qdisc", "add", "dev", t.NIC.GW, "parent", "1:10", "handle", "110:", "netem", "delay", delay, "limit", qlenReturnVoIP)
		}
		runTC("class", "add", "dev", t.NIC.GW, "parent", "1:1", "classid", "1:20", "htb", "rate", returnRest, "prio", "1", "burst", "30k", "cburst", "30k")
		if jitter {
			runTC("qdisc", "add", "dev", t.NIC.GW, "parent", "1:20", "handle", "120:", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenReturnRest)
		} else {
			runTC("qdisc", "add", "dev", t.NIC.GW, "parent", "1:20", "handle", "120:", "netem", "delay", delay, "limit", qlenReturnRest)
		}
		// Filters
		runTC("filter", "add", "dev", t.NIC.GW, "protocol", "ip", "parent", "1:0", "prio", "0", "handle", "11", "fw", "flowid", "1:10")
		runTC("filter", "add", "dev", t.NIC.GW, "protocol", "ip", "parent", "1:0", "prio", "1", "handle", "10", "fw", "flowid", "1:20")

		// Qdisc configuration
		runTC("qdisc", "add", "dev", t.NIC.ST, "root", "handle", "1:0", "htb", "default", "20")
		runTC("class", "add", "dev", t.NIC.ST, "parent", "1:0", "classid", "1:1", "htb", "rate", forward, "burst", "30k", "cburst", "30k")
		runTC("class", "add", "dev", t.NIC.ST, "parent", "1:0", "classid", "1:10", "htb", "rate", forwardVoIP, "prio", "0", "burst", "3k", "cburst", "3k")
		if jitter {
			runTC("qdisc", "add", "dev", t.NIC.ST, "parent", "1:10", "handle", "110:", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenForwardVoIP)
		} else {
			runTC("qdisc", "add", "dev", t.NIC.ST, "parent", "1:10", "handle", "110:", "netem", "delay", delay, "limit", qlenForwardVoIP)
		}
		runTC("class", "add", "dev", t.NIC.ST, "parent", "1:0", "classid", "1:20", "htb", "rate", forwardRest, "prio", "1", "burst", "30k", "cburst", "30k")
		if jitter {
			runTC("qdisc", "add", "dev", t.NIC.ST, "parent", "1:20", "handle", "120:", "netem", "delay", delay, offset, "distribution", "normal", "limit", qlenForwardRest)
		} else {
			runTC("qdisc", "add", "dev", t.NIC.ST, "parent", "1:20", "handle", "120:", "netem", "delay", "limit", qlenForwardRest)
		}

		// Filters
		runTC("filter", "add", "dev", t.NIC.ST, "protocol", "ip", "parent", "1:0", "prio", "0", "handle", "21", "fw", "flowid", "1:10")
		runTC("filter", "add", "dev", t.NIC.ST, "protocol", "ip", "parent", "1:0", "prio", "1", "handle", "20", "fw", "flowid", "1:20")
	}

	if t.ACMEnabled {
		log.Println("Trunks started with ACM")
		scheduler := gocron.NewScheduler(time.UTC)
		scheduler.Every(1).Seconds().Do(t.RunACM, t.QoS)
		scheduler.StartBlocking()
	} else {
		log.Println("Trunks started without ACM")
		select {}
	}

}
