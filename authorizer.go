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
	rmAuthorizer, err := auth.NewAuthorizerFromCLI()
	if err != nil {
		log.Fatal(err)
	}
	subscriptionsClient.Authorizer = rmAuthorizer
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
	tenantsClient.Authorizer = rmAuthorizer
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

	graphAuthorizer, err := auth.NewAuthorizerFromCLIWithResource(env.GraphEndpoint)

	spClient := graphrbac.NewServicePrincipalsClient(tId)
	spClient.Authorizer = graphAuthorizer

	spIterator, err := spClient.ListComplete(context.TODO(), "")
	if err != nil {
		log.Fatal(err)
	}
	for spIterator.NotDone() {
		sp := spIterator.Value()
		if *sp.PublisherName != "Microsoft Services" {
			log.Printf("Service Principal: %s\n", *sp.DisplayName)
		}
		err = spIterator.NextWithContext(context.TODO())
		if err != nil {
			log.Fatal(err)
		}
	}

	appClient := graphrbac.NewApplicationsClient(tId)
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
}
