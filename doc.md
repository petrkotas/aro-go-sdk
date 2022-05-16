# Copy & Paster how-to deploy cluster via ARO GO SDK

Documentation:
 - [How-to create ARO cluster](https://docs.microsoft.com/en-us/azure/openshift/tutorial-create-cluster)
 - [How-to create service principal](https://docs.microsoft.com/en-us/azure/openshift/howto-create-service-principal?pivots=aro-azurecli)

The excerpt with copy paste snippets bellow.

The snippets bellow prepare the subscription with minimal access granted on one resource group that hosts ARO and VNET.
Second resource group that will contain ARO resources will be auto created.


## Prepare the subscription

Setup the whereabouts.
```bash
LOCATION=eastus                 # the location of your cluster
RESOURCEGROUP=pkotas-sdk-test   # the name of the resource group where you want to create your cluster
CLUSTER=pkotas-sdk-test         # the name of your cluster
```

Create the resource group for the cluster.
```bash
az group create \
  --name $RESOURCEGROUP \
  --location $LOCATION
```

Create the VNET.
```
az network vnet create \
   --resource-group $RESOURCEGROUP \
   --name aro-vnet \
   --address-prefixes 10.0.0.0/22
```

Create master subnet.
```bash
az network vnet subnet create \
  --resource-group $RESOURCEGROUP \
  --vnet-name aro-vnet \
  --name master-subnet \
  --address-prefixes 10.0.0.0/23 \
  --service-endpoints Microsoft.ContainerRegistry
```

Create worker subnet.
```bash
az network vnet subnet create \
  --resource-group $RESOURCEGROUP \
  --vnet-name aro-vnet \
  --name worker-subnet \
  --address-prefixes 10.0.2.0/23 \
  --service-endpoints Microsoft.ContainerRegistry
```

Enable private link on masters. This enables ARO to provide managed services.
```bash
az network vnet subnet update \
  --name master-subnet \
  --resource-group $RESOURCEGROUP \
  --vnet-name aro-vnet \
  --disable-private-link-service-network-policies true
```

## Prepare service principals

Create cluster service principal that enables cluster to operate its resources, e.g. scale nodes.
```bash
az ad sp create-for-rbac --name cluster-sp
```

Copy the appId (called clientId in SDK), and find the object id by looking at the table.

```bash
az ad sp list --show-mine -o table
```
Copy the objectId and use it next.

`Network Contributor` enables cluster to manage its networking.
The `Contributor` on the cluster resource group that enabled cluster to manage its resources, is added later by the ARO during the installation.
```bash
az role assignment create --role 'Network Contributor' --assignee-object-id '<cluster-sp-objectId>' --resource-group $RESOURCEGROUP --assignee-principal-type 'ServicePrincipal'
```

Verify assigned roles.
```bash
az role assignment list --all --assignee '<cluster-sp-objectId>' -o table
```


Create SDK login credentials, that will enable SDK to communicate with ARO and create cluster in designated Resource group.
```bash
az ad sp create-for-rbac --name sdk-sp
```

Copy the appId (called clientId in SDK), and find the object id by looking at the table.

```bash
az ad sp list --show-mine -o table
```
Copy the objectId and use it next.


Assign the roles on the resource group required by ARO.
- `Contributor` enables ARO to manage cluster object and network.
- `Network Contributor` enables ARO to attach resources to the VNET.
- `User Access Administrator` enables ARO to grant cluster service principal the `Contributor` role on newly created resource group.

```bash
az role assignment create --role 'Contributor' --assignee-object-id '<sdk-sp-objectId>' --resource-group $RESOURCEGROUP --assignee-principal-type 'ServicePrincipal'
az role assignment create --role 'Network Contributor' --assignee-object-id '<sdk-sp-objectId>' --resource-group $RESOURCEGROUP --assignee-principal-type 'ServicePrincipal'
az role assignment create --role 'User Access Administrator' --assignee-object-id '<sdk-sp-objectId>' --resource-group $RESOURCEGROUP --assignee-principal-type 'ServicePrincipal'
```

Verify created resources.
```bash
az role assignment list --all --assignee '<sdk-sp-objectId>' -o table
```