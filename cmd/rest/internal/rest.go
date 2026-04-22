package internal

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
)

const (
	CERT_FILE     string = "CERT_FILE"
	KEY_FILE      string = "KEY_FILE"
	HTTPS_ENABLED string = "HTTPS_ENABLED"
	HTTP_PORT     string = "HTTP_PORT"
)

var (
	certFile     string = "./certs/ssl.crt"
	keyFile      string = "./certs/ssl.key"
	httpPort     string = "80"
	httpsEnabled bool   = false
)

func GetCertificates() (tls.Certificate, error) {
	bytesCert, err := os.ReadFile(certFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	bytesKey, err := os.ReadFile(keyFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	return tls.X509KeyPair(bytesCert, bytesKey)
}

func Main(pwd string, args []string, envs map[string]string, osSignal chan os.Signal) error {
	var wg sync.WaitGroup

	//print version
	fmt.Printf("Version: \"%s\"\n", Version)
	fmt.Printf("Git Commit: \"%s\"\n", GitCommit)
	fmt.Printf("Git Branch: \"%s\"\n", GitBranch)

	//generate and create handle func, when connecting, it will use this port
	//indicate via console that the webserver is starting
	if v, ok := envs[CERT_FILE]; ok && v != "" {
		certFile = v
	}
	if v, ok := envs[KEY_FILE]; ok && v != "" {
		keyFile = v
	}
	if v, ok := envs[HTTP_PORT]; ok && v != "" {
		httpPort = v
	}
	httpsEnabled, _ = strconv.ParseBool(envs[HTTPS_ENABLED])
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, "Hello, World!")
	})
	server := &http.Server{
		Addr: ":" + httpPort,
	}
	if httpsEnabled {
		tlsCertificate, err := GetCertificates()
		if err != nil {
			return err
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{tlsCertificate},
		}
	}
	fmt.Printf("starting web server on :%s\n", httpPort)
	stopped := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(stopped)

		var err error

		switch {
		default:
			err = server.ListenAndServe()
		case httpsEnabled:
			err = server.ListenAndServeTLS(certFile, keyFile)
		}
		if err != nil {
			fmt.Println(err)
		}
	}()
	select {
	case <-stopped:
	case <-osSignal:
		return server.Close()
	}
	wg.Wait()
	return nil
}
