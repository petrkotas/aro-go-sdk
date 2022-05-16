package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redhatopenshift/armredhatopenshift"
	"github.com/Azure/go-autorest/autorest/to"
)

func main() {
	log.Println("Starting the cluster deployment")

	// SDK login information
	tenantID := "64dc69e4-d083-49fc-9569-ebece1dd1408"
	// clientId == appId of the created service principal
	clientID := "7ac602cf-a265-4922-8f89-f48cd5cec6f6"
	clientSecret := "7wC8r7Ue1NR29TixxrYg11a_-0EtlxNc2i"

	// cluster details
	subscriptionID := "fe16a035-e540-4ab7-80d9-373fa9a3d6ae"
	clusterResourceGroup := "pkotas-sdk-test"
	clusterName := "pkotas-sdk-test"
	clusterLocation := "eastus"
	// clusterServicePrincipalClientID == appId of the created service principal
	clusterServicePrincipalClientID := "954303c8-9913-496d-92ea-f1cc0bedaae2"
	clusterServicePrincipalClientSecret := "8~LwbDZZcD5EJVihe3QIavbPLt-T42DGdQ"
	// network configuration
	masterSubnetResourceID := "/subscriptions/fe16a035-e540-4ab7-80d9-373fa9a3d6ae/resourceGroups/pkotas-sdk-test/providers/Microsoft.Network/virtualNetworks/aro-vnet/subnets/master-subnet"
	workerSubnetResourceID := "/subscriptions/fe16a035-e540-4ab7-80d9-373fa9a3d6ae/resourceGroups/pkotas-sdk-test/providers/Microsoft.Network/virtualNetworks/aro-vnet/subnets/worker-subnet"
	// default used by `az aro` client
	podCIDR := "10.128.0.0/14"
	serviceCIDR := "172.30.0.0/16"

	cluster := armredhatopenshift.OpenShiftCluster{
		Location: to.StringPtr(clusterLocation),
		Properties: &armredhatopenshift.OpenShiftClusterProperties{
			ApiserverProfile: &armredhatopenshift.APIServerProfile{
				Visibility: (*armredhatopenshift.Visibility)(to.StringPtr("Public")),
			},
			IngressProfiles: []*armredhatopenshift.IngressProfile{
				{
					Name:       to.StringPtr("default"),
					Visibility: (*armredhatopenshift.Visibility)(to.StringPtr("Public")),
				},
			},
			ClusterProfile: &armredhatopenshift.ClusterProfile{
				Domain:          to.StringPtr(clusterName),
				PullSecret:      to.StringPtr(""),
				ResourceGroupID: to.StringPtr(fmt.Sprintf("/subscriptions/%s/resourceGroups/%s", subscriptionID, "aro-"+strings.ToLower(clusterName))),
			},
			ServicePrincipalProfile: &armredhatopenshift.ServicePrincipalProfile{
				ClientID:     to.StringPtr(clusterServicePrincipalClientID),
				ClientSecret: to.StringPtr(clusterServicePrincipalClientSecret),
			},
			NetworkProfile: &armredhatopenshift.NetworkProfile{
				PodCidr:     &podCIDR,
				ServiceCidr: &serviceCIDR,
			},
			MasterProfile: &armredhatopenshift.MasterProfile{
				SubnetID: to.StringPtr(masterSubnetResourceID),
				VMSize:   (*armredhatopenshift.VMSize)(to.StringPtr("Standard_D8s_v3")),
			},
			WorkerProfiles: []*armredhatopenshift.WorkerProfile{
				{
					Name:       to.StringPtr("worker"),
					Count:      to.Int32Ptr(3),
					DiskSizeGB: to.Int32Ptr(128),
					VMSize:     (*armredhatopenshift.VMSize)(to.StringPtr("Standard_D4s_v3")),
					SubnetID:   to.StringPtr(workerSubnetResourceID),
				},
			},
		},
	}

	// read the credentials from azidentity
	credentials, err := azidentity.NewClientSecretCredential(
		tenantID,
		clientID,
		clientSecret,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}

	// create the client
	client, err := armredhatopenshift.NewOpenShiftClustersClient(subscriptionID, credentials, nil)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// create the cluster
	createPoller, err := client.BeginCreateOrUpdate(ctx, clusterResourceGroup, clusterName, cluster, nil)
	if err != nil {
		log.Fatalf("Failed to create poller: %v", err)
	}

	createResult, err := createPoller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		log.Fatalf("Could not create the cluster: %v", err)
	}

	log.Print(createResult.ID)

	// Delete the cluster
	deletePoller, err := client.BeginDelete(ctx, clusterResourceGroup, clusterName, nil)
	if err != nil {
		log.Fatalf("Failed to create poller: %v", err)
	}

	deleteResult, err := deletePoller.PollUntilDone(ctx, 30*time.Second)
	if err != nil {
		log.Fatalf("Could not delete the cluster: %v", err)
	}

	log.Print(deleteResult)

	return
}
