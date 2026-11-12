# 阿里云 RDS MySQL SSL 配置

Erda 平台支持使用 PEM CA 证书连接开启 SSL 的阿里云 RDS MySQL。该能力默认关闭，不改变现有非 SSL 部署。

云侧配置和证书下载参见[阿里云 RDS MySQL SSL 官方文档](https://help.aliyun.com/zh/rds/apsaradb-rds-for-mysql/configure-a-cloud-certificate-to-enable-ssl-encryption)。

## 配置约定

平台服务使用以下环境变量：

| 环境变量 | 示例 | 说明 |
|---|---|---|
| `MYSQL_TLS` | `rds` | Go MySQL 驱动的命名 TLS 配置。RDS 场景使用自定义名称，不使用 `skip-verify` 或 `preferred`。 |
| `MYSQL_CACERTPATH` | `/etc/erda/mysql-tls/ApsaraDB-CA-Chain.pem` | PEM CA 在容器内的绝对路径。 |

`erda.yml` 通过部署值 `mysql_tls`、`mysql_ca_cert_path` 将这两个变量传给其中的数据库消费服务。两个值默认为空；保持为空时沿用原有非 SSL 连接。非数据库服务不接收这两个变量，避免要求它们挂载 CA。

推荐配置：

```text
MYSQL_TLS=rds
MYSQL_CACERTPATH=/etc/erda/mysql-tls/ApsaraDB-CA-Chain.pem
```

命名 TLS 会校验证书链和数据库连接主机名，等价于 `VERIFY_IDENTITY`。`MYSQL_HOST` 应使用证书覆盖的 RDS 连接地址，不应使用未包含在证书 SAN 中的 IP、任意别名或代理地址。

## CA 证书交付

从阿里云 RDS 控制台下载 PEM 格式 CA 链。不要把证书提交到 Erda 或业务代码仓库，也不要打入平台镜像。

通过 Kubernetes Secret 或等效受控配置源，把 CA 只读挂载到所有数据库消费容器的相同路径。主发布描述不自动创建 Secret，也不声明默认宿主机 bind；这样未启用 SSL 的现有部署不会因为缺少证书文件而失败。

至少覆盖：

- `erda-server`
- `cluster-manager`
- `gittar`
- `hepa`
- `monitor`
- `orchestrator`
- `pipeline`
- `streaming`
- `log-service`
- 实际发布清单中的 AI Proxy、MCP Proxy 以及其他使用平台 MySQL provider 的组件

示意 Kubernetes 配置：

```yaml
volumes:
  - name: mysql-rds-ca
    secret:
      secretName: erda-mysql-rds-ca
containers:
  - name: erda-component
    volumeMounts:
      - name: mysql-rds-ca
        mountPath: /etc/erda/mysql-tls
        readOnly: true
```

Secret 中的文件名应与 `MYSQL_CACERTPATH` 一致。实际 Erda CR、Helm values 或发布系统配置不在本仓库时，需要在对应安装配置仓库同步完成挂载。

平台进程在启动时读取 CA 并注册 TLS 配置。更新挂载的 CA 后，需要滚动重启数据库消费服务以重新加载证书。启用、关闭或更新 RDS SSL 可能触发实例重启或主备切换，应按阿里云变更窗口执行并验证平台连接恢复。

## 数据库迁移

Erda CLI：

```sh
erda-cli migrate \
  --mysql-host example.mysql.rds.aliyuncs.com \
  --mysql-port 3306 \
  --mysql-username erda \
  --mysql-password '<password>' \
  --database erda \
  --mysql-tls rds \
  --mysql-ca-cert-path /etc/erda/mysql-tls/ApsaraDB-CA-Chain.pem
```

`migrate record` 使用相同参数。

`erda-mysql-migration` Action 使用：

```yaml
params:
  mysql_tls: rds
  mysql_ca_cert_path: /etc/erda/mysql-tls/ApsaraDB-CA-Chain.pem
```

运行 Action 前，流水线执行器必须把 CA 挂载到参数指定路径。迁移主库使用 TLS，本地 sandbox 不继承 RDS TLS 配置。

本仓库的默认 CI 流水线不引用 TLS 配置项，以免未启用 SSL 的环境因缺失 `${{ configs.* }}` 而解析失败。需要对 CI 迁移启用 SSL 时，在对应环境创建例如 `legacy_migration_mysql_tls`、`legacy_migration_mysql_ca_cert_path` 的配置，并在迁移 Action 中显式透传：

```yaml
params:
  mysql_tls: ${{ configs.legacy_migration_mysql_tls }}
  mysql_ca_cert_path: ${{ configs.legacy_migration_mysql_ca_cert_path }}
```

Action 支持需要先发布包含 TLS migrator API 的 Erda 主仓版本，再更新并发布 `erda-actions` 中 1.0-56、1.0-57、1.0-80 对应镜像和规格。

## 验收

1. 在预发布环境挂载 CA，并设置两个环境变量。
2. 使用证书覆盖的 RDS 域名启动全部数据库消费服务，确认不存在 CA 读取、PEM 解析或主机名校验错误。
3. 在 RDS 会话执行：

   ```sql
   SHOW STATUS LIKE 'Ssl_cipher';
   ```

   返回值应为非空 TLS cipher。
4. 使用错误 CA、缺失 CA 和证书不匹配的连接地址分别验证启动或迁移明确失败，且不会回退明文。
5. 按阿里云变更窗口启用或更新 RDS SSL，验证实例重启/切换后平台服务可依赖现有连接池与重连机制恢复。

## Pipeline Action 元数据

Pipeline 页面会读取公开 Action 元数据来展示节点名称、图标和描述。完成平台部署后，先确认 Action 目录不是空的：

```sh
curl 'http://<erda-server>/api/extensions?public=true&type=action'
```

如果某个 Pipeline 引用的 Action 尚未导入，应通过 Erda Extension API 或 `erda-cli ext push` 导入对应 `spec.yml`、`dice.yml` 和 README。不要直接写扩展数据表。

后端在 Action 元数据缺失或 `spec.yml` 无效时仍会保留已解析的 Pipeline Action，并使用 alias 或 Action type 作为显示名，避免流程图把所有 stage 变成空数组。Action 元数据本身仍应补齐，否则页面可以渲染，但对应任务可能无法执行。

安装后还应确认标准运行依赖已经导入：

- Action Agent：`agent@1.0`
- 应用 addon：`redis@6.2.10`、`monitor@3.6`
- addon Operator：`redis-operator@6.2.10`

这些元数据应通过 Extension API 或 `erda-cli ext push` 导入。Orchestrator 会在启动时加载 addon 类型并在内存中维护索引；导入后如果当前进程未及时刷新，可滚动重启 Orchestrator。

如果 addon 调度记录已经丢失，删除 addon 时不能因为 ServiceGroup 返回 `not found` 而保留数据库中的孤儿路由。Orchestrator 应将缺失的调度记录视为已清理，并继续删除 addon routing/instance 记录，否则后续部署会持续复用损坏的 addon。

## 回滚

先将 RDS 侧策略恢复为允许非 SSL 连接，再清空 `MYSQL_TLS` 和 `MYSQL_CACERTPATH`，滚动重启平台服务并移除 CA 挂载。若 RDS 仍强制 SSL，提前清空客户端配置会导致平台数据库连接失败。
