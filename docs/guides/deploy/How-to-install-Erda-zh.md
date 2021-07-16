# 如何安装 Erda 

### 先决条件

- Kuberentes 1.16 +
  - 至少需要 4 个节点 (1 个 Master 和 3 个 Worker)
  - 每个节点 4 CPU 16 G 内存
  - **不要在集群上安装 ingress controller 组件**
- Docker 19.03 +
- CentOS 7.4 +
- Helm 3 +
- 泛域名(可选项，用于访问 Erda 平台，如 *.erda.io)



### 安装 Erda

1. 在您的 Kubernetes Master 节点上下载 [压缩包](https://github.com/erda-project/erda/releases) 并解压
	
   > **Note**: 当前仅支持 Linux 系统
   
   ```shell
   tar -xzvf erda-linux.tar.gz
   cd erda
   ```



2. 在 Kubernetes Master 节点上设置安装 Erda 时的必要配置

   - 请确保 `~/.kube/` 路径下有 **kubeconfig** 文件
      - 请确保 kubeconfig 文件中有如下配置
      	- `certificate-authority-data`
      	- `client-certificate-data`
      	- `client-key-data`

      
      
   - 设置 Erda 安装前的配置并且执行 `prepare.sh` 脚本
   
     - 脚本中会执行如下操作:
       - 生成 ETCD 的 SSL
       - 生成多集群管理的 SSL
       - 为节点设置上 Erda 应用所需要用的标签
       - Erda 安装工具中需要的一些配置
   
     ```shell
     # 可以在此处指定 Erda 组件所在的命名空间，默认为 default 且当前仅支持 default 命名空间
     export ERDA_NAMESPACE="default"
     
     # 可以在此处指定 Erda 平台所用的泛域名，如 *.erda.io，默认值为 erda.io
     export ERDA_GENERIC_DOMAIN="erda.io"
     
     # 可以在此处指定 Erda 平台所用的集群名称，默认为 erda-demo
     export ERDA_CLUSTER_NAME="erda-demo"
     
     # 执行 prepare.sh 脚本，用于设置 Erda 平台安装时必要的配置
     bash scripts/prepare.sh
     ```

     

   - 修改 docker daemon 文件中的 `insecure-registries` 字段
   
      ```shell
      # 在*每台节点*上编辑 /etc/docker/daemon.json 文件
      ...
          "insecure-registries": [
              "0.0.0.0/0"
          ],
      ...
      
      # 重启 docker daemon
      systemctl restart docker
      ```

      
   
   - 在每个节点上设置 NFS 作为网络共享存储
   
      - 如您有如阿里云的网络共享存储您可以用如下命令将其设置在**每台节点**上:
      
        ```shell
        mount -t <storage_type> <your-share-storage-node-ip>:<your-share-storage-dir> /netdata
        
        # 举例如下：假设您拥有阿里云 NAS v4 服务作为共享网络存储，阿里云 NAS 的 Host 为 file-system-id.region.nas.aliyuncs.com 您需要通过如下命令挂载目录:
        
        mount -t nfs -o vers=4,minorversion=0,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2,noresvport file-system-id.region.nas.aliyuncs.com:/ /netdata  
        ```
   
      - 否则您需要执行如下脚本，它会协助安装 NFS 组件，在当前节点上创建 `/netdata` 文件夹并将其挂载到剩余的节点上

        ```shell
        bash scripts/storage_prepare.sh
        ```
      
      
      
      
   - 您需要在 LB 机器上开放 80， 443 端口，这台机器将承载所有的外部流量
   
      ```shell
      # 您可以使用如下命令在您的 Kubernetes 集群上找到 LB 节点：
      
      kubectl get node -o wide --show-labels | grep dice/lb
      ```
   
      - 记住节点的 IP，您将会在后续设置**泛域名**时用到该节点 IP



3. 通过 Helm 安装 Erda Helm 包，并且等待所有的 Erda 组件准备就绪

   ```shell
   # 安装 erda-base
   helm install package/erda-base-$(cat VERSION).tgz --generate-name
   
   # 安装 erda-addons
   helm install package/erda-addons-$(cat VERSION).tgz --generate-name
   
   # 安装 erda
   helm install package/erda-$(cat VERSION).tgz --generate-name
   ```

   

4. 安装 Erda 平台组件之后

   - 设置管理员用户名和密码，用于推送 Erda 扩展组件（扩展组件将作为一种插件被用于流水线）

     ```shell
     export ERDA_ADMIN_USERNAME=admin
     export ERDA_ADMIN_PASSWORD=password123456
     
     bash scripts/push-ext.sh
     ```

     

   - 如果您购买了真实的泛域名，您需要将其与上述获取到的 LB 节点 IP 绑定

     > 举例如下，假设您有 LB 节点的 IP 为 10.0.0.1，泛域名(ERDA_GENERIC_DOMAIN 变量中设置)为 `erda.io`, 您需要将 LB 节点 IP 与泛域名在您的解析器（如 DNS 或 F5 服务器）上绑定

     

   - 如果没有真实的泛域名地址, 您需要在您浏览器所在的机器上将下列的 URL 写到 `etc/hosts` 文件中，请替换 IP 为 **LB 节点**的 IP

     > 举例如下:  假设您有 LB 节点的 IP 为 10.0.0.1，泛域名(ERDA_GENERIC_DOMAIN 变量中设置)为 `erda.io`, org-name 为 `erda-test`. 所以您需要将下列的信息写入到 `/etc/hosts` 文件中

     ```shell
     10.0.0.1 collector.erda.io
     10.0.0.1 openapi.erda.io
     10.0.0.1 uc.erda.io
     10.0.0.1 erda.io
     # 注意: org-name 举例为 erda-test
     10.0.0.1 erda-test-org.erda.io
     ```

     

   - 将您创建好的组织名称作为标签设置到您的 Kubernetes 节点上（组织名称是 Erda 中的一种组）

     ```shell
     kubectl label node <node_name> dice/org-<orgname>=true --overwrite
     ```

     

5. 在您设置过 `/etc/hosts` 文件的机器上用浏览器访问 `http://erda.io`，并开始您的 Erda 之旅