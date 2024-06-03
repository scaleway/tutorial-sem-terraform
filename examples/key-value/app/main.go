package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	_ "github.com/lib/pq"
	"github.com/scaleway/scaleway-sdk-go/api/rdb/v1"
	secret "github.com/scaleway/scaleway-sdk-go/api/secret/v1beta1"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	dbname           = "rdb"
	databaseInstance = "dataBaseInstance"
)

// Auth credentials
type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {

	organizationID := os.Getenv("SCW_DEFAULT_ORGANIZATION_ID")
	defaultProjectID := os.Getenv("SCW_DEFAULT_PROJECT_ID")
	accessKey := os.Getenv("SCW_ACCESS_KEY")
	secretKey := os.Getenv("SCW_SECRET_KEY")
	defaultRegion := os.Getenv("SCW_DEFAULT_REGION")
	fmt.Printf("SCW_DEFAULT_REGION: %s", defaultRegion)
	// Create a Scaleway client
	client, err := scw.NewClient(
		// Get your organization ID at https://console.scaleway.com/organization/settings
		scw.WithDefaultOrganizationID(organizationID),
		scw.WithDefaultRegion(scw.Region(defaultRegion)),
		scw.WithDefaultProjectID(defaultProjectID),
		// Get your credentials at https://console.scaleway.com/iam/api-keys
		scw.WithAuth(accessKey, secretKey),
	)
	if err != nil {
		panic(err)
	}

	// Create SDK objects for Scaleway Secret Manager product
	secretApi := secret.NewAPI(client)

	// Name of the secret
	secretName := "database_secret"
	// Call the GetSecretVersionByName method on the Secret SDK
	response, err := secretApi.AccessSecretVersionByPath(&secret.AccessSecretVersionByPathRequest{
		SecretName: secretName,
		SecretPath: "/",
		Revision:   "latest",
	})
	if err != nil {
		panic(err)
	}

	// Unmarshal the credentials
	var auth Auth
	err = json.Unmarshal(response.Data, &auth)
	if err != nil {
		panic(err)
	}

	rdbAPI := rdb.NewAPI(client)
	instanceName := databaseInstance
	rdbInstance, err := rdbAPI.ListInstances(&rdb.ListInstancesRequest{Name: &instanceName})
	if err != nil {
		panic(err)
	}

	i := rdbInstance.Instances[0]
	// Retrieve the database endpoint
	var endpointData *rdb.Endpoint
	for _, endpoint := range i.Endpoints {
		if endpoint.LoadBalancer != nil {
			endpointData = endpoint
			break
		}
	}
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		endpointData.IP, endpointData.Port, auth.Username, auth.Password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
}
