package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
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
			log.Fatalf("ERROR: Invalid 'API_PORT' variable (Value=%s)", apiPortEnv)
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
	caCert, err := os.ReadFile(apiCaCert)
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

	log.Printf(`
	----------------------------
	OIDC Discovery Configuration
	----------------------------
	apiService: %s
	apiCaCert: %s
	tokenFile: %s
	tlsEnabled: %t
	`, config.apiService, apiCaCert, config.tokenFile, config.tlsEnabled)
}

func getAuthToken() (string, error) {
	fileContent, err := os.ReadFile(config.tokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %w", err)
	}
	return strings.TrimSuffix(string(fileContent), "\n"), nil
}

func getOidcConfiguration() (string, error) {
	// Read token fresh for this request
	authToken, err := getAuthToken()
	if err != nil {
		return "", fmt.Errorf("failed to get auth token: %w", err)
	}

	// Make request
	url := "https://kubernetes.default.svc.cluster.local/.well-known/openid-configuration"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Get result and pass result body back
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func getJwks() (string, error) {
	// Read token fresh for this request
	authToken, err := getAuthToken()
	if err != nil {
		return "", fmt.Errorf("failed to get auth token: %w", err)
	}

	// Make request
	url := fmt.Sprintf("https://%s/openid/v1/jwks", config.apiService)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Get result and pass result back
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(body), nil
}

func oidcConfiguration(w http.ResponseWriter, r *http.Request) {
	body, err := getOidcConfiguration()
	if err != nil {
		log.Printf("Error getting OIDC configuration: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, body)
}

func jwks(w http.ResponseWriter, r *http.Request) {
	body, err := getJwks()
	if err != nil {
		log.Printf("Error getting JWKS: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, body)
}

func handleRequests() {
	http.HandleFunc("/.well-known/openid-configuration", oidcConfiguration)
	http.HandleFunc("/openid/v1/jwks", jwks)
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	log.Printf("Starting listener..")
	if config.tlsEnabled {
		log.Fatal(http.ListenAndServeTLS(":8443", "/certs/tls.crt", "/certs/tls.key", nil))
	} else {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}
}

func main() {
	handleRequests()
}
