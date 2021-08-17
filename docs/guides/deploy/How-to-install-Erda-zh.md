# 如何安装 Erda 

### 先决条件

- 硬件资源配置（不含运行 Kubernetes 组件所需资源）

  > **Note:** 当前 Erda 只支持 Demo 规模部署

  | 规模 | CPU（核） | Memory（GB） | Storage（GB） | 推荐配置                                   |
  | ---- | --------- | ------------ | ------------- | ------------------------------------------ |
  | Demo | 8         | 32           | 400           | 规模：2 节点 <br/>规格： 4 核/16 GB/200 GB |

- Kubernetes 1.16 - 1.20 (安装 [ingress controller](https://kubernetes.io/zh/docs/concepts/services-networking/ingress-controllers/) 组件)

- Docker 19.03 及以上

- CentOS 7.4 及以上

- Helm 3 及以上

- 泛域名(可选项，通过 Kubernetes Ingress 配置域名来访问 Erda 平台，如 *.erda.io)



### 安装 Erda

1. 在您的 Kubernetes Master 节点上下载 [压缩包](https://github.com/erda-project/erda/releases) 并解压
	
   > **Note**: 当前仅支持 Linux 系统
   
   ```shell
   tar -xzvf erda-linux.tar.gz
   cd erda
   ```

2. 在 Kubernetes Master 节点上设置安装 Erda 时的必要配置

   - 确认 Master 节点的 `~/.kube/` 路径下有 kubeconfig 文件，并且可以使用 `kubectl` 访问集群

   - 确认 Master 节点下已安装 Helm（以 3.5.2 版本为例）。

     ```shell
     # 下载 Helm 安装包
     wget https://get.helm.sh/helm-v3.5.2-linux-amd64.tar.gz
     
     # 解压安装包
     tar -xvf helm-v3.5.2-linux-amd64.tar.gz
     
     # 安装 Helm 3，在解压后的目录 linux-amd64 中找到 Helm 二进制文件，然后将其移至所需的目标位置
     mv linux-amd64/helm /usr/local/bin/helm
     
     # Erda Chart 包直接在本地解压文件中，无需添加 Repo， Helm 添加 Repo 等操作请参考官方文档  
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

3. 通过 Helm 安装 Erda ，并且等待所有的 Erda 组件准备就绪

   > **Note:**  当前版本 Erda 仅支持安装在 `default namespace` 下

   ```shell
   cd erda-helm
   
   # 指定 Erda 集群名称, erda.clusterName=erda-test
   # 指定 Erda 平台的泛域名, erda.domain=erda.io
   helm install erda erda-$(cat VERSION).tgz --set erda.clusterName="erda-demo",erda.domain="erda.io"
   ```

   > 如果您在 Kubernetes 节点上无法直接访问 Kubernetes 内部域名 （例如 *kubernetes.default.svc.cluster.local*），安装 Erda 时需指定一个 Node 以 `hostNework` 安装 Registry，并且 `--set registry.custom.nodeIP="",registry.custom.nodeName=""`  参数，否则您将无法使用流水线功能

4. 安装 Erda 平台组件之后

   - 设置管理员用户名和密码，用于推送 Erda 扩展组件（扩展组件将作为一种插件被用于流水线）

     ```shell
     export ERDA_ADMIN_USERNAME=admin
     export ERDA_ADMIN_PASSWORD=password123456
     
     bash scripts/push-ext.sh
     ```
     
   - 如果有真实的泛域名，您需要进行域名配置，将该域名的访问流量导入 Kubernetes 集群的 Ingress Controller，让集群中配置的 Ingress 域名能正常访问

   - 如果没有真实的泛域名, 您需要在浏览器所在的机器上将下列的 URL 写到 `etc/hosts` 文件中，请将下面的示例 IP 替换为该 Kubernetes 集群的 Ingress Controller 的入口 IP

     > 举个例子，假设您的 Kubernetes 集群的 Ingress Controller 的入口 IP 为 10.0.0.1，泛域名(ERDA_GENERIC_DOMAIN 变量中设置)为 `erda.io`, org-name 为 `erda-test`, 需要将下列的信息写入到 `/etc/hosts` 文件中

     ```shell
     10.0.0.1 collector.erda.io
     10.0.0.1 openapi.erda.io
     10.0.0.1 uc.erda.io
     10.0.0.1 erda.io
     ```
     
   - 将您创建好的组织名称作为标签设置到您的 Kubernetes 节点上（组织名称是 Erda 中的一种组）

     ```shell
     kubectl label node <node_name> dice/org-<orgname>=true --overwrite
     ```

5. 最后您可以通过浏览器访问 `http://erda.io`，并开始您的 Erda 之旅

## 卸载 Erda

1. 通过 Helm 卸载 Erda

   ```shell
   # 卸载 erda
   helm uninstall erda 
   ```
   
2. 默认情况下，通过 Helm 卸载并不会删除 Erda 所依赖中间件的 `pvc`, 请您按需手动清理





