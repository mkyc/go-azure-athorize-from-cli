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
	results, err := subscriptionsClient.List(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	for _, sub := range results.Values() {
		log.Printf("Subscription name: %s and ID: %s", *sub.DisplayName, *sub.ID)
	}
}
