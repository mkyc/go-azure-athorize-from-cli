package main

import (
	"context"
	"log"

	"github.com/Azure/azure-sdk-for-go/profiles/2019-03-01/resources/mgmt/subscriptions"
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
	//var tId string
	for _, ten := range tenantsPage.Values() {
		log.Printf("Some id: %s and some other id: %s\n", *ten.ID, *ten.TenantID)
		//tId = *ten.TenantID
	}
	if len(subscriptionsPage.Values()) != 1 || len(tenantsPage.Values()) != 1 {
		log.Fatal("There is more than 1 subscription or tenant")
	}
}
