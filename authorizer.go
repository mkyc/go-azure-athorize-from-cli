package main

import (
	"context"
	"log"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/resources/mgmt/subscriptions"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/authorization/mgmt/authorization"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/graphrbac/graphrbac"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	uuid "github.com/satori/go.uuid"
	"github.com/sethvargo/go-password/password"
)

const (
	cloudName        = "AzurePublicCloud"
	defaultPublisher = "Microsoft Services"
	roleName         = "Contributor"
	appName          = "20201031-auto-test-1"
)

type Credentials struct {
	appId        string
	password     string
	tenant       string
	subscription string
}

func main() {

	subscriptionsClient := subscriptions.NewClient()
	resourceManagerAuthorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		log.Fatal(err)
	}
	subscriptionsClient.Authorizer = resourceManagerAuthorizer
	subscriptionsIterator, err := subscriptionsClient.ListComplete(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	var subscriptionId string
	subscriptionsCount := 0
	for subscriptionsIterator.NotDone() {
		subscriptionsCount++
		sub := subscriptionsIterator.Value()
		log.Printf("Subscription name: %s and ID: %s\n", *sub.DisplayName, *sub.ID)
		subscriptionId = *sub.SubscriptionID
		err = subscriptionsIterator.NextWithContext(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}

	tenantsClient := subscriptions.NewTenantsClient()
	tenantsClient.Authorizer = resourceManagerAuthorizer
	tenantsIterator, err := tenantsClient.ListComplete(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	var tenantId string
	tenantsCount := 0
	for tenantsIterator.NotDone() {
		tenantsCount++
		ten := tenantsIterator.Value()
		log.Printf("Some tenant id: %s\n", *ten.ID)
		tenantId = *ten.TenantID
		err = tenantsIterator.NextWithContext(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}

	if subscriptionsCount != 1 || tenantsCount != 1 {
		log.Fatal("There is more than 1 subscription or tenant")
	}

	env, err := azure.EnvironmentFromName(cloudName)
	if err != nil {
		log.Fatal(err)
	}

	graphAuthorizer, err := auth.NewAuthorizerFromCLIWithResource(env.GraphEndpoint)

	spClient := graphrbac.NewServicePrincipalsClient(tenantId)
	spClient.Authorizer = graphAuthorizer

	spIterator, err := spClient.ListComplete(context.TODO(), "")
	if err != nil {
		log.Fatal(err)
	}
	for spIterator.NotDone() {
		sp := spIterator.Value()
		if *sp.PublisherName != defaultPublisher {
			log.Printf("Service Principal Name: %s, ID: %s\n", *sp.DisplayName, *sp.ObjectID)
		}
		err = spIterator.NextWithContext(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}

	appClient := graphrbac.NewApplicationsClient(tenantId)
	appClient.Authorizer = graphAuthorizer

	appIterator, err := appClient.ListComplete(context.TODO(), "")
	if err != nil {
		log.Fatal(err)
	}
	for appIterator.NotDone() {
		app := appIterator.Value()
		log.Printf("Application: %s\n", *app.DisplayName)
		err = appIterator.NextWithContext(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}

	roleDefinitionClient := authorization.NewRoleDefinitionsClient(subscriptionId)
	roleDefinitionClient.Authorizer = resourceManagerAuthorizer

	roleDefinitionIterator, err := roleDefinitionClient.ListComplete(context.TODO(), "/subscriptions/"+subscriptionId, "")
	if err != nil {
		log.Fatal(err)
	}
	var roleId string
	for roleDefinitionIterator.NotDone() {
		rd := roleDefinitionIterator.Value()
		if *rd.RoleName == roleName {
			roleId = *rd.ID
			log.Printf("RoleDefinition: %s\n", *rd.RoleName)
		}
		err = roleDefinitionIterator.NextWithContext(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}

	keyId := uuid.NewV4()
	pass, err := password.Generate(32, 10, 0, false, false)
	if err != nil {
		log.Fatal(err)
	}
	t := &date.Time{
		Time: time.Now(),
	}
	t2 := &date.Time{
		Time: t.AddDate(2, 0, 0),
	}
	app, err := appClient.Create(context.TODO(), graphrbac.ApplicationCreateParameters{
		DisplayName:             to.StringPtr(appName),
		IdentifierUris:          &[]string{"https://" + appName},
		AvailableToOtherTenants: to.BoolPtr(false),
		Homepage:                to.StringPtr("https://" + appName),
		PasswordCredentials: &[]graphrbac.PasswordCredential{{
			StartDate:           t,
			EndDate:             t2,
			KeyID:               to.StringPtr(keyId.String()),
			Value:               to.StringPtr(pass),
			CustomKeyIdentifier: to.ByteSlicePtr([]byte(appName)),
		}},
	})
	if err != nil {
		log.Fatal(err)
	}
	appJson, err := app.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("\n===========\nAPPLICATION\n%v\n===========\n", string(appJson))

	sp, err := spClient.Create(context.TODO(), graphrbac.ServicePrincipalCreateParameters{
		AppID:          app.AppID,
		AccountEnabled: to.BoolPtr(true),
	})
	if err != nil {
		log.Fatal(err)
	}
	spJson, err := sp.MarshalJSON()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("\n===========\nSERVICE PRINCIPAL\n%v\n===========\n", string(spJson))

	roleAssignmentClient := authorization.NewRoleAssignmentsClient(subscriptionId)
	roleAssignmentClient.Authorizer = resourceManagerAuthorizer

	roleAssignmentName := uuid.NewV4()
	for i := 0; i < 30; i++ {
		ra, err := roleAssignmentClient.Create(context.TODO(), "/subscriptions/"+subscriptionId, roleAssignmentName.String(), authorization.RoleAssignmentCreateParameters{
			Properties: &authorization.RoleAssignmentProperties{
				RoleDefinitionID: to.StringPtr(roleId),
				PrincipalID:      sp.ObjectID,
			},
		})
		if err != nil {
			log.Println(err)
			time.Sleep(1 * time.Second)
			continue
		} else {
			raJson, err := ra.MarshalJSON()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("\n===========\nROLE ASSIGNMENT\n%v\n===========\n", string(raJson))
			break
		}
	}

	c := &Credentials{
		appId:        *sp.AppID,
		password:     pass,
		tenant:       tenantId,
		subscription: subscriptionId,
	}
	log.Printf("\n===========\nCREDENCIALS\n%+v\n===========\n", c)
}
