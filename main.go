package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	azidentity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraph "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go-core/authentication"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var (
	clientID       = getEnv("CLIENT_ID", "")
	clientSecret   = getEnv("CLIENT_SECRET", "")
	tenantID       = getEnv("TENANT_ID", "")
	targetIDs      = strings.Split(getEnv("TARGET_IDS", ""), ",")
	pushgatewayURL = getEnv("PUSHGATEWAY_URL", "")

	secretExpiryDate = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "azad_secret_expiry_timestamp",
		Help: "Expiry date for Azure secrets as a Unix timestamp",
	}, []string{"application_name", "application_id", "status"})
)

func main() {
	// Register Prometheus metrics
	prometheus.MustRegister(secretExpiryDate)

	// Create a new context
	ctx := context.Background()

	// Set up the credentials for authentication
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		log.Fatalf("Failed to create credentials: %v", err)
	}

	// Create an auth provider
	authProvider, err := authentication.NewAzureIdentityAuthenticationProviderWithScopes(cred, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		log.Fatalf("Failed to create authentication provider: %v", err)
	}

	// Create a Graph client
	adapter, err := msgraph.NewGraphRequestAdapter(authProvider)
	if err != nil {
		log.Fatalf("Failed to create Graph request adapter: %v", err)
	}

	client := msgraph.NewGraphServiceClient(adapter)

	for _, targetID := range targetIDs {
		// Get the application details using its ID to retrieve the display name
		appDetails, err := client.ApplicationsWithAppId(&targetID).Get(ctx, nil)
		if err != nil {
			log.Printf("Failed to get application details for client ID %s: %v", targetID, err)
			continue
		}

		if appDetails != nil {
			appName := *appDetails.GetDisplayName()
			appID := *appDetails.GetAppId()
			fmt.Printf("Application Name: %s\n", appName)
			fmt.Printf("Application ID: %s\n", appID)

			passwordCredentials := appDetails.GetPasswordCredentials()
			if passwordCredentials == nil {
				fmt.Println("No password credentials found.")
				continue
			}

			for _, cred := range passwordCredentials {
				expiryDate := cred.GetEndDateTime()
				if expiryDate != nil {
					fmt.Printf("Secret expires on: %s\n", expiryDate.Format(time.RFC3339))

					status := "ACTIVE"
					if time.Now().After(*expiryDate) {
						status = "EXPIRED"
					}
					fmt.Printf("Status: %s\n", status)

					secretExpiryDate.WithLabelValues(appName, appID, status).Set(float64(expiryDate.Unix()))
				} else {
					fmt.Println("No expiry date found for secret.")
				}
			}
		} else {
			fmt.Println("Failed to retrieve application details.")
			continue
		}
	}

	// Push metrics to the Pushgateway
	if err := push.New(pushgatewayURL, "azuread_app_checker").
		Collector(secretExpiryDate).
		Push(); err != nil {
		log.Fatalf("Could not push metrics to Pushgateway: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		if defaultValue == "" {
			log.Fatalf("Environment variable %s is required but not set.", key)
		}
		return defaultValue
	}
	return value
}
