# How to install Erda

### Prerequisites

- Kubernetes 1.16 - 1.18
  - At least 4 Nodes (1 Master and 3 Workers)
  - Each node needs at least 4 core CPU, 16G memory
  - Install the [ingress controller](https://kubernetes.io/zh/docs/concepts/services-networking/ingress-controllers/) component
- Docker 19.03 +
- CentOS 7.4 +
- Helm 3 +
- Generic Domain Name (Optional, configure the domain name through Kubernetes Ingress to access the Erda platform，e.g. *.erda.io)

### Install Erda

1. Download the [tarball](https://github.com/erda-project/erda/releases) to your  **Kubernetes Master** node.

   > **Note**: Only support install on Linux currently

   ```shell
   tar -xzvf erda-linux.tar.gz
   cd erda
   ```

2. Apply Erda necessary configurations on the **Kubernetes Master Node**.

   - make sure the **kubeconfig** file on the ~/.kube/config.

     - Make sure the kubeconfig contains following configuration
       - `certificate-authority-data`
       - `client-certificate-data`
       - `client-key-data`

   - set configuration to prepare to install Erda and execute the `prepare.sh` script

     - The script will do the following tasks:
       - generate etc SSL
       - set node labels which use for Erda Application
       - set Erda installer configuration   

     ```shell
     # specify the Kubernetes namespace to install Erda components, the default value is default and the Erda components are only support the default namespace
     export ERDA_NAMESPACE="default"
     
     # specify the generic domain name like *.erda.io to visit the erda application, default values is erda.io, you can set owner generic domain name in here
     export ERDA_GENERIC_DOMAIN="erda.io"
     
     # The ERDA_CLUSTER_NAME specified for Erda which will be used in cluster creating
     export ERDA_CLUSTER_NAME="erda-demo"
     
     # Execute the script to apply Erda necessary configuration
     bash scripts/prepare.sh
     ```

   - update `insecure-registries` in the config of the docker daemon 

     ```shell
     # edit the /etc/docker/daemon.json on *each node*
     ...
         "insecure-registries": [
             "0.0.0.0/0"
         ],
     ...
     
     # restart the docker daemon
     systemctl restart docker
     ```

   - set NFS share storage as network storage to each node. 

     - if you already have share storage like AliCloud NAS, you need to set them to **each node** with command like:

       ```shell
       mount -t <storage_type> <your-share-storage-node-ip>:<your-share-storage-dir> /netdata
       
       # for example use AliCloud NAS v4 as share storage，and AliCloud NAS Host is file-system-id.region.nas.aliyuncs.com you need to mount the directory with command:
       
       mount -t nfs -o vers=4,minorversion=0,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2,noresvport file-system-id.region.nas.aliyuncs.com:/ /netdata
       ```

     - if not, you need to execute the script. It will install NFS utils, create a directory `/netdata` to the current machine, and mount `/netdata` to each node

       ```shell
       bash scripts/storage_prepare.sh
       ```

3. Install the Erda with helm package and waiting all Erda components are ready

   ```shell
   # Install erda-base, involving some dependent operators
   helm install erda-base package/erda-base-$(cat VERSION).tgz 
   
   # Install erda-addons, the middleware that the erda platform relies on
   helm install erda-addons package/erda-addons-$(cat VERSION).tgz 
   
   # install erda
   helm install erda package/erda-$(cat VERSION).tgz 
   ```

4. After Installed Erda

   - set administrator user name and password to push the Erda extensions（the extension is a plugin which uses in the pipeline）

     ```shell
     export ERDA_ADMIN_USERNAME=admin
     export ERDA_ADMIN_PASSWORD=password123456
     
     bash scripts/push-ext.sh
     ```

   - If you have a real generic domain name, you need to configure the domain name and import the access traffic of the domain name into the Ingress Controller of the Kubernetes cluster, so that the Ingress domain name configured in the cluster can be used normally.

   - If not, you need to write the following URLs to `/etc/hosts` on the **machine where the browser is located**, replace the <IP> below with the entry IP of the Ingress Controller of the Kubernetes cluster.

     > For example, suppose the IP of the ingress controller is `10.0.0.1`, ERDA_GENERIC_DOMAIN is `erda.io`, org-name is `erda-test`, write the following info to `/etc/hosts` 

     ```shell
     10.0.0.1 collector.erda.io
     10.0.0.1 openapi.erda.io
     10.0.0.1 uc.erda.io
     10.0.0.1 erda.io
     # Note: The org-name of this example is erda-test
     10.0.0.1 erda-test-org.erda.io
     ```

   - set your Kubernetes nodes label with your created organization name（organization is a name for a group in Erda）

     ```shell
     kubectl label node <node_name> dice/org-<orgname>=true --overwrite
     ```

5. Finally, you can visit `http://erda.io` through your browser and start your Erda journey

### Uninstall Erda

Uninstall Erda via Helm

   ```shell
   helm uninstall erda 
   
   helm uninstall erda-addons 
   
   helm uninstall erda-base 
   ```
