CREATE TABLE IF NOT EXISTS `ai_proxy_mcp_server_template`
(
    `id`         BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Primary Key',
    `mcp_name`   varchar(255)        NOT NULL COMMENT 'MCP 名称',
    `version`    varchar(128)        NOT NULL COMMENT 'MCP 版本',
    `template`   TEXT COMMENT '配置模板',

    `created_at` DATETIME            NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME            NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME            NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    PRIMARY KEY (`id`),
    UNIQUE INDEX `unique_name_version` (`mcp_name`, `version`)
    ) ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy MCP 配置模板表';

INSERT INTO `ai_proxy_mcp_server_template` (`mcp_name`, `version`, `template`, `description`)
VALUES
    ('enterprise-tools', '1.0.0', '[
  {
    "description": "天眼查 Token",
    "name": "token",
    "required": true,
    "type": "string"
  }
]', '提供用于查询企业信息和上市公司股票信息的工具'),
    ('html-generator', '2.0.0', '[
  {
    "default": "",
    "desc": "生成HTML的额外说明",
    "name": "instructions",
    "required": false,
    "type": "string"
  },
  {
    "desc": "模型名称",
    "name": "model_name",
    "required": true,
    "type": "string"
  },
  {
    "desc": "模型发布商",
    "name": "model_publisher",
    "required": true,
    "type": "string"
  }
]', '一个简单的HTML生成工具，可帮助用户生成图表等内容'),
    ('search-oil', '1.0.0', '[]', '一个搜索工具，提供汽油价格和螺纹钢价格查询功能'),
    ('pptx-server', '1.0.0', '[]', '一个PPT模板生成工具，提供多种模板构建功能'),
    ('pptx-agent', '1.0.0', '[]', '一个PPT生成工具，可根据用户提供的数据生成PPT'),
    ('mcp-timer', '1.0.0', '[]', '提供时间戳与日期时间互相转换的两个工具'),
    ('mcp-server-baidu-maps', '1.11.0', '[{
    "desc": "百度地图 AccessKey",
    "name": "ak",
    "required": true,
    "type": "string",
    "scope": "query"
  }]', '百度地图MCP服务器，符合MCP标准的开源LBS解决方案，为开发者和AI代理提供地理位置相关API与工具'),
    ('playwright', '1.0.0', '[]', '基于Playwright的浏览器自动化与测试工具'),
    ('office-word-mcp-server', '1.0.0', '[]', '用于创建、读取和操作Microsoft Word文档的MCP服务器'),
    ('mcp-milvus', '1.0.0', '[]', '提供Milvus向量数据库的访问功能'),
    ('mcp-fetch', '1.0.0', '[]', '网页内容抓取工具'),
    ('search-engine', '3.0.0', '[]', '为大语言模型（LLM）提供联网搜索功能的工具'),
    ('mcp-python-script', '1.0.0', '[]', '提供远程执行Python脚本的功能'),
    ('mcp-stock', '1.0.0', '[]', '用于查询中国上市公司股票、财务和市场信息的工具'),
    ('mcp-calculator', '1.0.0', '[]', '一个MCP计算器，可计算数学表达式，如 2+5*3-(5-2)'),
    ('search-engine', '1.0.0', '[]', '为大语言模型（LLM）提供联网搜索功能的工具'),
    ('mcp-ocr', '1.0.0', '[]', '提供快速准确的光学字符识别（OCR）功能，用于从图片中提取文字'),
    ('mcp-email', '1.0.0', '[
  {
    "description": "用户名称",
    "name": "user",
    "required": true,
    "type": "string"
  },
  {
    "description": "用户名密码",
    "name": "password",
    "required": true,
    "type": "string"
  },
  {
    "description": "IMAP 服务器地址（收件）",
    "name": "imap_host",
    "required": true,
    "type": "string"
  },
  {
    "description": "IMAP 服务器端口",
    "name": "imap_port",
    "required": true,
    "type": "integer"
  },
  {
    "description": "SMTP 服务器地址（发件）",
    "name": "smtp_host",
    "required": true,
    "type": "string"
  },
  {
    "description": "SMTP 服务器端口",
    "name": "smtp_port",
    "required": true,
    "type": "integer"
  }
]', '提供邮件发送功能的工具'),
    ('google-maps', '1.0.0', '[]', '提供谷歌地图服务的工具'),
    ('pymupdf4llm-mcp-server', '1.0.0', '[]', '用于将PDF文件转换为Markdown格式以供LLM使用的MCP服务器'),
    ('12306', '1.0.0', '[]', '12306火车票查询服务MCP服务器'),
    ('mcp-exchange-rate', '1.0.0', '[]', '提供实时汇率查询功能的工具'),
    ('pptx-server', '2.0.0', '[]', '一个PPT模板生成工具，提供多种模板构建功能');



CREATE TABLE IF NOT EXISTS `ai_proxy_mcp_server_config_instance`
(
    `id`            CHAR(36)     NOT NULL COMMENT 'Primary Key',
    `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`    DATETIME     NOT NULL DEFAULT '1970-01-01 00:00:00' COMMENT '删除时间, 1970-01-01 00:00:00 表示未删除',

    `instance_name` VARCHAR(255) NOT NULL DEFAULT 'default' COMMENT '默认实例名称',
    `client_id`     CHAR(36)     NOT NULL COMMENT '客户端 id',
    `config`        TEXT COMMENT 'MCP 配置',
    `mcp_name`      varchar(255) NOT NULL COMMENT 'MCP 名称',
    `version`       varchar(128) NOT NULL COMMENT 'MCP 版本',

    PRIMARY KEY (`id`),
    UNIQUE INDEX `unique_mcp_name_version_client_id` (`instance_name`, `mcp_name`, `version`, `client_id`, `deleted_at`)
    ) ENGINE = InnoDB
    DEFAULT CHARSET = utf8mb4 COMMENT 'ai-proxy MCP 配置实例表';
