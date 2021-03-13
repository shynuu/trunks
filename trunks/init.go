package trunks

import (
	"errors"
	"fmt"
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
func CheckInterfaces(st string, gw string) error {
	path := "/sys/class/net/%s/operstate"
	ifST := fmt.Sprintf(path, st)
	ifGW := fmt.Sprintf(path, gw)
	existST, _ := Exists(ifST)
	if !existST {
		fmt.Println("Interface ST does not exists")
		return errors.New("Interface ST does not exists")
	}
	existGW, _ := Exists(ifGW)
	if !existGW {
		fmt.Println("Interface GW does not exists")
		return errors.New("Interface GW does not exists")
	}
	return nil
}

// InitTrunks initialize the trunks module
func InitTrunks(file string) error {
	err := ParseConf(file)
	if err != nil {
		return err
	}
	err = CheckInterfaces(Trunks.NIC.ST, Trunks.NIC.GW)
	return err
}
