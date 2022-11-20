package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"log"
	"os"
	"strconv"
)

// todo: add logger
// Connect connect
func Connect(ctx context.Context, uri string, ca, cert, key []byte) (client *proxmox.Client, err error) {
	if len(ca) == 0 || len(cert) == 0 || len(key) == 0 {
		return nil, fmt.Errorf("certificates length are not valid")
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(ca)

	certificate, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return nil, fmt.Errorf("failed to load client cert and key: %w", err)
	}

	config := &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{certificate},
		InsecureSkipVerify: false,
	}

	//insecure := flag.Bool("insecure", true, "TLS insecure mode")
	proxmox.Debug = flag.Bool("debug", false, "debug mode")
	//fConfigFile := flag.String("file", "", "file to get the config from")
	taskTimeout := flag.Int("timeout", 300, "api task timeout in seconds") // todo: timeout
	proxyURL := flag.String("proxy", "", "proxy url to connect to")
	fvmid := flag.Int("vmid", -1, "custom vmid (instead of auto)")
	flag.Parse()

	conn, err := proxmox.NewClient(uri, nil, "", config, *proxyURL, *taskTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed creating new proxmox client: %w", err)
	}

	if userRequiresAPIToken(PmUser) {
		conn.SetAPIToken(PmUser, PmPass)
		// As test, get the version of the server
		_, err := conn.GetVersion()
		if err != nil {
			log.Fatalf("login error: %s", err)
		}
	} else {
		err = conn.Login(PmUser, PmPass, os.Getenv("PM_OTP"))
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

	return conn, nil
}
