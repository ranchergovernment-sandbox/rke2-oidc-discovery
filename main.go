package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	apiService string
	tokenFile  string
	tlsEnabled bool
}

var authToken string
var config Config
var client http.Client

func init() {
	// Initialization of Config

	// K8s API Service
	config.apiService = "kubernetes.default.svc.cluster.local"
	if apiServiceEnv := os.Getenv("API_SERVICE"); apiServiceEnv != "" {
		config.apiService = apiServiceEnv
	}

	apiPort := 443
	if apiPortEnv := os.Getenv("API_PORT"); apiPortEnv != "" {
		i, err := strconv.Atoi(apiPortEnv)
		if err != nil {
			log.Fatal(fmt.Sprintf("ERROR: Invalid 'API_PORT' variable (Value=%s)", apiPortEnv))
		} else {
			apiPort = i
		}
	}
	config.apiService = fmt.Sprintf("%s:%d", config.apiService, apiPort)

	// Token File
	config.tokenFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	if tokenFileEnv := os.Getenv("TOKEN_FILE"); tokenFileEnv != "" {
		config.tokenFile = tokenFileEnv
	}

	// TLS Enabled
	config.tlsEnabled = false
	if tlsEnabledEnv := os.Getenv("TLS_ENABLED"); tlsEnabledEnv != "" {
		if strings.EqualFold(tlsEnabledEnv, "true") {
			config.tlsEnabled = true
		}
	}

	// K8s API Service
	apiCaCert := "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"
	if apiCaCertEnv := os.Getenv("API_CA_CERT"); apiCaCertEnv != "" {
		apiCaCert = apiCaCertEnv
	}

	// Configure HTTP client with ca.crt
	caCert, err := ioutil.ReadFile(apiCaCert)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	client = http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: caCertPool,
			},
		},
	}

	log.Printf(fmt.Sprintf(`
	----------------------------
	OIDC Discovery Configuration
	----------------------------
	apiService: %s
	apiCaCert: %s
	tokenFile: %s
	tlsEnabled: %t
	`, config.apiService, apiCaCert, config.tokenFile, config.tlsEnabled))
}

func getAuthToken() {
	fileContent, err := ioutil.ReadFile(config.tokenFile)
	if err != nil {
		log.Fatal(err)
	}

	authToken = strings.TrimSuffix(string(fileContent), "\n")
}

func getOidcConfiguration() string {
	// Make request
	url := "https://kubernetes.default.svc.cluster.local/.well-known/openid-configuration"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Get result and pass result body back
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return string(body)
}

func getJwks() string {
	// Make request
	url := fmt.Sprintf("https://%s/openid/v1/jwks", config.apiService)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Get result and pass result back
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return string(body)
}

func oidcConfiguration(w http.ResponseWriter, r *http.Request) {
	body := getOidcConfiguration()
	fmt.Fprintf(w, body)
}

func jwks(w http.ResponseWriter, r *http.Request) {
	body := getJwks()
	fmt.Fprintf(w, body)
}

func handleRequests() {
	http.HandleFunc("/.well-known/openid-configuration", oidcConfiguration)
	http.HandleFunc("/openid/v1/jwks", jwks)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte{})
	})

	log.Printf("Starting listener..")
	if config.tlsEnabled {
		log.Fatal(http.ListenAndServeTLS(":8443", "/certs/tls.crt", "/certs/tls.key", nil))
	} else {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}

func main() {
	getAuthToken()
	handleRequests()
}
