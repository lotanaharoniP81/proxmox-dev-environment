package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"log"
	"regexp"
)

func main() {
	ca := []byte(caPem)
	cert := []byte(certPem)
	key := []byte(certKey)

	convertedNode, err := strconv.Atoi(node)
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*30)
	defer cancel()
	p := NewProxmoxAPI(ca, cert, key)
	err = p.QemuCreate(ctx, PmApiUrl, host, convertedNode, "create_qemu.json")
	if err != nil {
		fmt.Printf("error! %v\n", err)
		return
	}
}

func createQemu(c *proxmox.Client, vmID int, fConfigFile string) error {
	configTemp, err := GetConfig(fConfigFile)
	if err != nil {
		return err
	}
	config, err := proxmox.NewConfigQemuFromJson(configTemp)
	if err != nil {
		return err
	}
	vmr := proxmox.NewVmRef(vmID)
	vmr.SetNode(host)
	return config.CreateVm(vmr, c)
}

func failError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var rxUserRequiresToken = regexp.MustCompile("[a-z0-9]+@[a-z0-9]+![a-z0-9]+")

func userRequiresAPIToken(userID string) bool {
	return rxUserRequiresToken.MatchString(userID)
}

func getVMList(c *proxmox.Client) (string, error) {
	vms, err := c.GetVmList()
	if err != nil {
		return "", err
	}
	vmList, err := json.Marshal(vms)
	if err != nil {
		return "", err
	}
	return string(vmList), nil
}

func getStorage(c *proxmox.Client, storageID string) (string, error) {
	config, err := proxmox.NewConfigStorageFromApi(storageID, c)
	if err != nil {
		return "", err
	}
	cj, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}
	return string(cj), nil
}

func deleteStorage(c *proxmox.Client, storageID string) error {
	return c.DeleteStorage(storageID)
}

func stopVM(c *proxmox.Client, vmID int) error {
	vmr := proxmox.NewVmRef(vmID)
	_, err := c.StopVm(vmr)
	return err
}

func startVM(c *proxmox.Client, vmID int) error {
	vmr := proxmox.NewVmRef(vmID)
	_, err := c.StartVm(vmr)
	return err
}

func destroyVM(c *proxmox.Client, vmID int) error {
	vmr := proxmox.NewVmRef(vmID)
	_, err := c.StopVm(vmr)
	if err != nil {
		return err
	}
	_, err = c.DeleteVm(vmr)
	return err
}

func ifVMIdExists(c *proxmox.Client, vmID int) (bool, error) {
	ifVMIdExists, err := c.VMIdExists(vmID)
	if err != nil {
		return false, err
	}
	return ifVMIdExists, nil
}

func resetVM(c *proxmox.Client, vmID int) error {
	vmr := proxmox.NewVmRef(vmID)
	_, err := c.ResetVm(vmr)
	return err
}

func getStorageList(c *proxmox.Client) (string, error) {
	storage, err := c.GetStorageList()
	if err != nil {
		return "", err
	}
	storageList, err := json.Marshal(storage)
	if err != nil {
		return "", err
	}
	return string(storageList), nil
}

func getNodeList(c *proxmox.Client) (map[string]interface{}, error) {
	return c.GetNodeList()
}

func getVMInfo(c *proxmox.Client, vmID int) (map[string]interface{}, error) {
	vmr := proxmox.NewVmRef(vmID)
	info, err := c.GetVmInfo(vmr)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func getVMState(c *proxmox.Client, vmID int) (map[string]interface{}, error) {
	vmr := proxmox.NewVmRef(vmID)
	info, err := c.GetVmState(vmr)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func createStorage(c *proxmox.Client, storageID string, fConfigFile string) error {
	configTemp, err := GetConfig(fConfigFile)
	if err != nil {
		return err
	}
	config, err := proxmox.NewConfigStorageFromJson(configTemp)
	if err != nil {
		return err
	}
	return config.CreateWithValidate(storageID, c)
}

//func updateStorage(c *proxmox.Client, storageID string) error{
//	config, err := proxmox.NewConfigStorageFromJson(GetConfig(*fConfigFile))
//	if err != nil {
//		return err
//	}
//	return config.UpdateWithValidate(storageID, c)
//}
