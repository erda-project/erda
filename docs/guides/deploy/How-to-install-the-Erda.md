# How to install the Erda

### Prerequisites

- Kuberentes 1.16 +
  - Each node needs at least 4 core CPU, 16G memory
  - At least 4 Nodes (1 Master and 3 Workers)
  - **Don't Install the ingress-controller-manager component**
- Docker 19.03+
- CentOS 7.4 +
- Helm 3 +



### Install Erda

1. Download the [tarball](https://terminus-dice.oss-cn-hangzhou.aliyuncs.com/installer/erda-installer/erda-release.tar.gz) to your  **Kubernetes Master** node.

   ```shell
   tar -xzvf erda-release.tar.gz
   cd erda-release
   ```



2. Apply Erda necessary configurations on the **Kubernetes Master Node**.

   - make sure the **kubeconfig** file on the ~/.kube/config.

   - set configuration to prepare the Erda and execute the `prepare.sh` script

     - The script will do the following tasks:
       - generate etc SSL
       - generate multi-cluster manager SSL
       - set node labels which use for Erda Application
       - set Erda installer configuration   

     ```shell
     # specify the Kubernetes namespace to install Erda components, the default value is default and the Erda components are only support the default namespace
     export ERDA_NAMESPACE="default"
     
     # specify the generic domains like *.erda-demo.erda.io to visit the erda application, default values is erda-demo.erda.io, you can set owner generic domains in here
     export ERDA_GENERIC_DOMAIN="erda-demo.erda.io"
     
     # The ERDA_CLUSTER_NAME specified for Erda which will be used in cluster creating
     export ERDA_CLUSTER_NAME="erda-demo"
     
     # Execute the script to apply Erda necessary configuration
     bash scripts/prepare.sh
     ```

     

   - update `insecure-registries` in the config of the docker daemon 

     ```shell
     # edit the /etc/docker/daemon.json on each node
     ...
         "insecure-registries": [
             "0.0.0.0/0"
         ],
     ...
     
     # restart the docker daemon
     systemctl restart docker
     ```

     

   - set NFS storage as network storage to each node. 

     - if you already have share storage like AliCloud NAS, you can set them to each node with command like:

       ```shell
       mount -t <storage_type> <your-share-storage-node-ip>:<your-share-storage-dir> /netdata
       ```

       
     
     - if not, you can execute the script. It will install NFS utils, create a dir `/netdata` to the current machine, and mount `/netdata` to each node
     
       ```shell
       bash scripts/storage_prepare.sh
       ```
     
       

    - you need to open the 80, 443 ports of the **LB machine** , which will receivers all outside traffic

     

3. Install the Erda with helm package and waiting all Erda components are ready

   ```shell
   # install erda-base
   helm install package/erda-base-0.1.0.tgz --generate-name
   
   # install erda-addons
   helm install package/erda-addons-0.1.0.tgz --generate-name
   
   # install erda
   helm install package/erda-0.1.0.tgz --generate-name
   ```

   

4. After Installed the Erda

   - set admin username and password to push the Erda extensions（the extension is a plugin which uses in the pipeline）

     ```shell
     export ERDA_ADMIN_USERNAME=admin
     export ERDA_ADMIN_PASSWORD=password123456
     
     bash scripts/push-ext.sh
     ```

   - write the following URLs to `/etc/hosts` on the **machine where the browser is located**, replace the <IP> with IP of the **LB machine**

     > For example, if I have an LB machine whose IP is `10.0.0.1`, ERDA_GENERIC_DOMAIN is `erda-demo.erda.io`, org-name is `erda-test`. so I can write the following info to `/etc/hosts` 

     ```shell
     10.0.0.1 nexus.erda-demo.erda.io
     10.0.0.1 sonar.erda-demo.erda.io
     10.0.0.1 dice.erda-demo.erda.io
     10.0.0.1 uc-adaptor.erda-demo.erda.io
     10.0.0.1 soldier.erda-demo.erda.io
     10.0.0.1 gittar.erda-demo.erda.io
     10.0.0.1 collector.erda-demo.erda.io
     10.0.0.1 hepa.erda-demo.erda.io
     10.0.0.1 openapi.erda-demo.erda.io
     10.0.0.1 uc.erda-demo.erda.io
     # Note: The org-name of this example is erda-test
     10.0.0.1 erda-test-org.erda-demo.erda.io
     ```

   - set your Kubernetes nodes label with your created organization name（organization is a name for a group in Erda）

     ```shell
     kubectl label node <node_name> dice/org-<orgname>=true --overwrite
     ```

     

5. Visit the URL `http://dice.erda-demo.erda.io` on your browser machine which set the `/etc/hosts`