package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"flag"
	"fmt"
	"github.com/innovia/portal/server"
	"github.com/innovia/portal/server/signals"
	"k8s.io/klog/v2"
	"net/http"
	"os"
)

func main() {
	err := run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error running server: %v", err)
		os.Exit(0)
	}
}

// startServer
func startServer(ctx context.Context, client *server.KubernetesClient, stopCh <-chan struct{}, healthAddress, listenAddress, caCertFile, tlsCertFile, tlsPrivateKeyFile string) {
	routerApiHandler, err := server.ApiHandler(client)
	if err != nil {
		klog.Errorf("error getting API handler %v", err)
	}

	healthApiHandler, err := server.HealthCheckHandler(client)
	if err != nil {
		klog.Errorf("error getting healthcheck API handler %v", err)
	}

	// Create a CA certificate pool and add cert.pem to it
	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		klog.Errorf("error reading Root CA cert file %s: %v", caCertFile, err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	// Create a Server instance to listen on main port with the TLS config
	srv := &http.Server{
		Addr: listenAddress,
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  certPool,
			MinVersion: tls.VersionTLS12,
		},
		Handler: routerApiHandler,
	}
	klog.Infof("Server started and listening on https://%s", listenAddress)

	go func() {
		if err = srv.ListenAndServeTLS(tlsCertFile, tlsPrivateKeyFile); err != nil {
			klog.Error(err)
		}
	}()

	// Create a Server instance to listen on healthcheck port without TLS config
	// this is to allow kubelet which does nto have the client certificates to call it for
	// pod health checks
	healthSrv := &http.Server{
		Addr:    healthAddress,
		Handler: healthApiHandler,
	}
	klog.Infof("HealthChecks started and listening on http://%s", healthAddress)

	go func() {
		if err = healthSrv.ListenAndServe(); err != nil {
			klog.Error(err)
		}
	}()

	// wait here until signal is received
	<-stopCh
	klog.Info("Shutting down server")
	if err = srv.Shutdown(ctx); err != nil {
		klog.Errorf("could not shutdown server gracefully: %v", err) // failure/timeout shutting down the server gracefully
	}
	if err = healthSrv.Shutdown(ctx); err != nil {
		klog.Errorf("could not shutdown healthcheck server gracefully: %v", err) // failure/timeout shutting down the server gracefully
	}
}

func run(args []string) error {
	// set up signals, so we handle the first shutdown signal gracefully
	stopCh := signals.SetupSignalHandler()
	ctx := context.Background()
	var caCertFile, kubeconfig, masterURL, tlsPrivateKeyFile, tlsCertFile, listenAddress, healthAddress string

	flag.StringVar(&tlsPrivateKeyFile, "tls_private_key_file", "", "path to server private key")
	flag.StringVar(&caCertFile, "ca_cert_file", "", "path to ca root certificate")
	flag.StringVar(&tlsCertFile, "tls_cert_file", "", "path to server certificate")
	flag.StringVar(&listenAddress, "listen_address", ":8443", "server port")
	flag.StringVar(&healthAddress, "health_address", ":8080", "health check port")
	flag.StringVar(&kubeconfig, "kubeconfig", kubeconfig, "Full path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")

	flag.Parse()

	if tlsCertFile == "" || tlsPrivateKeyFile == "" || caCertFile == "" {
		return errors.New("--tls_cert_file or --tls_private_key_file or --ca_cert_file flags are missing")
	}

	client, err := server.NewClient(kubeconfig)
	if err != nil {
		return err
	}

	// Start reconcile loop
	go client.StartReconcileLoop(ctx, stopCh)

	startServer(ctx, client, stopCh, healthAddress, listenAddress, caCertFile, tlsCertFile, tlsPrivateKeyFile)

	return nil
}
