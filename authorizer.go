package main

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/resources/mgmt/subscriptions"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/graphrbac/graphrbac"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

func main() {

	subscriptionsClient := subscriptions.NewClient()
	authorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		log.Fatal(err)
	}
	subscriptionsClient.Authorizer = authorizer
	subscriptionsPage, err := subscriptionsClient.List(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	//var sId string
	for _, sub := range subscriptionsPage.Values() {
		log.Printf("Subscription name: %s and ID: %s\n", *sub.DisplayName, *sub.SubscriptionID)
		//sId = *sub.SubscriptionID
	}
	tenantsClient := subscriptions.NewTenantsClient()
	tenantsClient.Authorizer = authorizer
	tenantsPage, err := tenantsClient.List(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	var tId string
	for _, ten := range tenantsPage.Values() {
		log.Printf("Some tenant id: %s and some other id: %s\n", *ten.ID, *ten.TenantID)
		tId = *ten.TenantID
	}
	if len(subscriptionsPage.Values()) != 1 || len(tenantsPage.Values()) != 1 {
		log.Fatal("There is more than 1 subscription or tenant")
	}

	env, err := azure.EnvironmentFromName("AzurePublicCloud")
	if err != nil {
		log.Fatal(err)
	}

	a, err := auth.NewAuthorizerFromCLIWithResource(env.GraphEndpoint)

	spClient := graphrbac.NewServicePrincipalsClient(tId)
	spClient.Authorizer = a

	spPage, err := spClient.ListComplete(context.TODO(), "")
	if err != nil {
		log.Fatal(err)
	}
	for spPage.NotDone() {
		sp := spPage.Value()
		if *sp.PublisherName != "Microsoft Services" {
			log.Printf("Name: %s\n", *sp.DisplayName)
		}
		err := spPage.NextWithContext(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}
}
