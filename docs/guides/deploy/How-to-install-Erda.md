# How to install Erda

### Prerequisites

- Resource configuration（Does not contain the resources required to run Kubernetes components）
  
  > **Note:** Currently Erda only supports Demo scale deployment
  
  | Size | CPU（c） | Memory（GB） | Storage（GB） | Recommended configuration                |
  | ---- | -------- | ------------ | ------------- | ---------------------------------------- |
  | Demo | 8        | 32           | 400           | Node count: 2<br/>Size: 4 c/16 GB/200 GB |
  
- Kubernetes 1.16 - 1.20 (Install the [ingress controller](https://kubernetes.io/zh/docs/concepts/services-networking/ingress-controllers/) component)

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

   - make sure the **kubeconfig** file on the ~/.kube/config, can use `kubectl` to visit kubernetes api

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

   > **Note:** The current version of Erda only supports installation under `default namespace` 

   ```shell
   cd erda-helm
   
   # Specify Erda cluster name, erda.clusterName=erda-test
   # Specify the pan-domain name of the Erda platform, erda.domain=erda.io
   helm install erda erda-$(cat VERSION).tgz --set erda.clusterName="erda-demo",erda.domain="erda.io"
   ```

   > **Note:**  If you cannot directly access the Kubernetes internal domain name (for example, *kubernetes.default.svc.cluster.local*) on the Kubernetes node, you need to specify a Node when installing Erda to install the Registry with `hostNework`, and `--set registry.custom. nodeIP="",registry.custom.nodeName=""` parameter, otherwise you will not be able to use the pipeline function

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
     ```
     
   - set your Kubernetes nodes label with your created organization name（organization is a name for a group in Erda）

     ```shell
     kubectl label node <node_name> dice/org-<orgname>=true --overwrite
     ```

5. Finally, you can visit `http://erda.io` through your browser and start your Erda journey

### Uninstall Erda

1. Uninstall Erda via Helm

   ```shell
   helm uninstall erda 
   ```
   
2. By default, uninstalling via Helm will not delete the `pvc` of the middleware that Erda relies on, please clean it up manually as needed
