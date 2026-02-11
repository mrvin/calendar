package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"log/slog"
	"net/http"
	"os"
)

//nolint:tagliatelle
type ConfHTTPS struct {
	CertFile       string `yaml:"cert_file"`
	KeyFile        string `yaml:"key_file"`
	ServerCertFile string `yaml:"server_cert_file"`
}

type Conf struct {
	Host    string    `yaml:"host"`
	Port    int       `yaml:"port"`
	IsHTTPS bool      `yaml:"is_https"`
	HTTPS   ConfHTTPS `yaml:"https"`
}
type Client struct {
	http.Client
}

func New(conf *Conf) *Client {
	var transport *http.Transport

	if conf.IsHTTPS {
		cert, err := tls.LoadX509KeyPair(conf.HTTPS.CertFile, conf.HTTPS.KeyFile)
		if err != nil {
			slog.Warn("Failed to read client tls certificate and key file: " + err.Error())
		} else {
			serverCert, err := os.ReadFile(conf.HTTPS.ServerCertFile)
			if err != nil {
				slog.Warn("Failed to read server tls certificate file: " + err.Error())
			} else {
				pool := x509.NewCertPool()
				pool.AppendCertsFromPEM(serverCert)

				tlsConf := &tls.Config{
					Certificates: []tls.Certificate{cert},
					RootCAs:      pool,
				}

				tlsConf.BuildNameToCertificate()
				transport = &http.Transport{
					TLSClientConfig: tlsConf,
				}
			}
		}
	}
	return &Client{
		http.Client{
			Transport: transport,
		},
	}
}
