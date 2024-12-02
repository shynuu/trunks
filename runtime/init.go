package trunks

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
)

// Exists check if a folder or a file exists
func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CheckInterfaces checks if the interfaces exist
func (t *TrunksConfig) CheckInterfaces() error {
	path := "/sys/class/net/%s/operstate"
	ifST := fmt.Sprintf(path, t.NIC.A)
	ifGW := fmt.Sprintf(path, t.NIC.B)
	var err1, err2 error
	existST, _ := Exists(ifST)
	if !existST {
		err1 = errors.New("[L2] Interface for ST not found")
		log.Println(err1.Error())
	}
	existGW, _ := Exists(ifGW)
	if !existGW {
		err2 = errors.New("[L2] Interface for GW not found")
		log.Println(err2.Error())
	}

	if err1 != nil || err2 != nil {
		return errors.New("")
	}

	return nil
}

func (t *TrunksConfig) FindInterfaces() error {
	ip_st := t.NIC.A
	ip_gw := t.NIC.B
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("Error reading interfaces: %+v\n", err.Error())
		return err
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Printf("localAddresses: %+v\n", err.Error())
			return err
		}
		for _, a := range addrs {
			switch v := a.(type) {
			case *net.IPNet:
				if v.IP.To4().String() == t.NIC.B {
					t.NIC.B = i.Name
				}
				if v.IP.To4().String() == t.NIC.A {
					t.NIC.A = i.Name
				}
			}

		}
	}

	var err1, err2 error
	if ip_st == t.NIC.A {
		err1 = errors.New("[L3] Interface for ST not found")
		log.Println(err1.Error())
	}

	if ip_gw == t.NIC.B {
		err2 = errors.New("[L3] Interface for GW not found")
		log.Println(err2.Error())
	}

	if err1 != nil || err2 != nil {
		return errors.New("")
	}

	return nil
}

// InitTrunks initialize the trunks module
func InitTrunks(file string, qos bool, logs string, acm bool, disable_kernel_version_check bool) (*TrunksConfig, error) {
	t, err := ParseConf(file)
	if err != nil {
		return nil, err
	}
	if err := t.CheckInterfaces(); err != nil {
		log.Println("Interfaces configuration by IP")
		if err := t.FindInterfaces(); err != nil {
			return nil, err
		}
	}

	t.QoS = qos
	t.Logs = logs
	t.ACMEnabled = acm
	t.KernelVersionCheck = !disable_kernel_version_check
	t.SetupBridge()
	return t, nil
}

// InitISL initialize the trunks module for ISL scenario
func InitISL(if_a string, if_b string, delay float64, offset float64, logs string, disable_kernel_version_check bool) (*TrunksConfig, error) {
	t := &TrunksConfig{}
	t.NIC.A = if_a
	t.NIC.B = if_b
	t.Delay.Value = delay
	t.Delay.Offset = offset
	if err := t.CheckInterfaces(); err != nil {
		log.Println("Interfaces configuration by IP")
		if err := t.FindInterfaces(); err != nil {
			return nil, err
		}
	}

	t.QoS = false
	t.Logs = logs
	t.ACMEnabled = false
	t.KernelVersionCheck = !disable_kernel_version_check
	t.SetupBridge()
	return t, nil
}
