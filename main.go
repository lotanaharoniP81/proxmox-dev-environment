package main

import (
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/Telmate/proxmox-api-go/cli/command/commands"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"log"
	"os"
	"regexp"
	"strconv"
)

func main() {
	insecure := flag.Bool("insecure", true, "TLS insecure mode")
	proxmox.Debug = flag.Bool("debug", false, "debug mode")
	//fConfigFile := flag.String("file", "", "file to get the config from")
	taskTimeout := flag.Int("timeout", 300, "api task timeout in seconds")
	proxyURL := flag.String("proxy", "", "proxy url to connect to")
	fvmid := flag.Int("vmid", -1, "custom vmid (instead of auto)")
	ca := []byte(caPem)
	cert := []byte(certPem)
	key := []byte(certKey)
	flag.Parse()
	tlsconf, err := tlsConfig(ca, cert, key)
	if err != nil {
		failError(err)
		os.Exit(0)
	}
	tlsconf.InsecureSkipVerify = false
	//tlsconf := &tls.Config{InsecureSkipVerify: true}
	if !*insecure {
		tlsconf = nil
	}
	//c, err := proxmox.NewClient(PmApiUrl, nil, os.Getenv("PM_HTTP_HEADERS"), tlsconf, *proxyURL, *taskTimeout)
	c, err := proxmox.NewClient(PmApiUrl, nil, "", tlsconf, *proxyURL, *taskTimeout)
	failError(err)
	if userRequiresAPIToken(PmUser) {
		c.SetAPIToken(PmUser, PmPass)
		// As test, get the version of the server
		_, err := c.GetVersion()
		if err != nil {
			log.Fatalf("login error: %s", err)
		}
	} else {
		err = c.Login(PmUser, PmPass, os.Getenv("PM_OTP"))
		failError(err)
	}

	vmid := *fvmid
	if vmid < 0 {
		//if len(flag.Args()) > 1 {
		if true {
			vmid, err = strconv.Atoi("123")
			if err != nil {
				fmt.Println("error")
				vmid = 0
			}
		} else if len(flag.Args()) == 0 || (flag.Args()[0] == "idstatus") {
			vmid = 0
		}
	}

	//var jbody interface{}
	//var vmr *proxmox.VmRef

	//fmt.Println("create storage")
	//if err := createStorage(c, "local-temp", "create_storage.json"); err != nil {
	//	fmt.Printf("create storage failed: %v\n", err)
	//}

	fmt.Println("get vm list")
	if list, err := getVMList(c); err != nil {
		fmt.Printf("get vm list failed: %v\n", err)
	} else {
		fmt.Printf("get vm list succeeded! the list: %v\n", list)
	}

	fmt.Println("create vm")
	if err := createQemu(c, 101, "create_qemu.json"); err != nil {
		fmt.Printf("create vm failed: %v\n", err)
	} else {
		fmt.Println("create vm succeeded!")
	}
}

func createQemu(c *proxmox.Client, vmID int, fConfigFile string) error {
	config, err := proxmox.NewConfigQemuFromJson(GetConfig(fConfigFile))
	if err != nil {
		return err
	}
	vmr := proxmox.NewVmRef(vmID)
	vmr.SetNode(hostConst)
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
	config, err := proxmox.NewConfigStorageFromJson(GetConfig(fConfigFile))
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
