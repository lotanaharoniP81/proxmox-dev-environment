package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/Telmate/proxmox-api-go/proxmox"
	"os"
)

const ProxmoxRequestTimeout = 300 // todo: in seconds?

type ProxmoxAPI struct {
	caCert     []byte
	clientCert []byte
	clientKey  []byte

	client *proxmox.Client

	//SemaphoreLocker // todo: check
}

// QemuCreate something.... (it's the same as Libvirt '')
func (api *ProxmoxAPI) QemuCreate(ctx context.Context, log Logger, host string, vmID int, fConfigFile string) error {
	//deadline, cancel := context.WithDeadline(ctx, time.Now().Add(api.ProxmoxRequestTimeout()))

	// todo: check on context in actions?
	config, err := GetConfig(fConfigFile)
	if err != nil {
		return fmt.Errorf("failed to get config from file") // todo: fix error
	}
	configQemu, err := proxmox.NewConfigQemuFromJson(config)
	if err != nil {
		return err
	}
	vmr := proxmox.NewVmRef(vmID)
	vmr.SetNode(host)
	return configQemu.CreateVm(vmr, api.client)
}

// Logger represents log object interface
type Logger interface {
	Info(a ...interface{})
	Debug(a ...interface{})
	Error(a ...interface{})
	Errorf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

// GetConfig get config from file
func GetConfig(configFile string) ([]byte, error) { // todo: change to 'getConfig'
	if configFile == "" {
		return nil, fmt.Errorf("file is empty") // todo: fix message (with error number)
	}
	configSource, err := os.ReadFile(configFile)
	//configSource = []byte("{\"name\":\"webserver20\",\"memory\":2048,\"cores\":1,\"sockets\":1,\"kvm\":false,\"iso\":\"local:iso/ubuntu-22.04.1-live-server-amd64.iso\"}")
	if err != nil {
		return nil, err
	}
	return configSource, nil
}

// tlsConfig something... (it's like Libvirt connect)
func tlsConfig(ca, cert, key []byte) (*tls.Config, error) {
	if len(ca) != 0 && len(cert) != 0 && len(key) != 0 {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(ca)

		cert, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert and key: %w", err)
		}

		config := &tls.Config{
			RootCAs:            caCertPool,
			Certificates:       []tls.Certificate{cert},
			InsecureSkipVerify: false,
		}
		return config, nil
	}
	return nil, fmt.Errorf("certificates length are not valid")
}
