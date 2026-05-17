# Nuclei Platform

基于 [Nuclei](https://github.com/projectdiscovery/nuclei) 的分布式漏洞扫描平台，提供 Web 管理界面、REST API、任务调度和水平扩展的 Worker 节点。

## 架构

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│  Web UI     │────▶│  API Server  │────▶│    NATS     │
│ (React/antd)│     │  (Go/chi)    │     │  (消息队列)  │
└─────────────┘     └──────┬───────┘     └──────┬──────┘
                           │                     │
                     ┌─────┴─────┐         ┌─────┴─────┐
                     │  SQLite   │         │  Workers  │
                     │  (数据存储) │         │ (可水平扩展)│
                     └───────────┘         └───────────┘
```

- **API Server**: 接收 HTTP 请求，管理任务/结果/Worker，通过 NATS 分发任务
- **Worker**: 订阅 NATS 任务队列，执行 Nuclei 扫描，回传结果
- **NATS**: 轻量消息队列，实现任务分发和心跳管理
- **SQLite**: 持久化存储任务、结果和 Worker 状态

## 快速开始

### Docker Compose（推荐）

```bash
# 一键启动（NATS + API Server + 3 Worker 副本）
make start

# 或手动操作
docker compose build
docker compose up -d

# 扩展 Worker 数量
make docker-scale N=5
```

启动后访问：
- Web 管理界面: http://localhost:8080
- NATS 监控: http://localhost:8222

### 本地开发

**环境要求**: Go 1.24+, Node.js 22+, CGO (SQLite 需要)

```bash
# 终端 1: 启动前端开发服务器（热更新，API 代理到 :8080）
make dev-frontend

# 终端 2: 启动 NATS
docker run -d --name nats -p 4222:4222 -p 8222:8222 nats:2.10-alpine --jetstream=false

# 终端 3: 启动 API Server
make run-server

# 终端 4: 启动 Worker
make run-worker
```

前端开发时访问 http://localhost:5173，API 请求自动代理到后端。

## 构建

```bash
# 完整构建（前端 + Go 二进制，前端内嵌到二进制中）
make build

# 分步构建
make build-frontend   # 构建前端到 web/dist/
make build-server     # 构建 API Server（内嵌前端）
make build-worker     # 构建 Worker

# 仅构建前端
make build-frontend

# 清理
make clean
```

构建产物:
- `bin/server` — API Server（内嵌前端静态文件，单文件部署）
- `bin/worker` — Worker 节点

## 项目结构

```
nuclei-platform/
├── cmd/
│   ├── server/main.go          # API Server 入口，内嵌前端
│   └── worker/main.go          # Worker 入口
├── internal/
│   ├── config/                  # 配置加载 (Viper)
│   ├── model/                   # 数据模型 (GORM)
│   └── server/
│       ├── server.go            # HTTP 路由 + 静态文件服务
│       ├── nats.go              # NATS 订阅/发布
│       └── handler/             # API Handler
├── web/                         # 前端项目
│   ├── src/
│   │   ├── api/                 # API 客户端 + 类型定义
│   │   ├── i18n/                # 中英双语
│   │   ├── pages/               # 页面组件
│   │   ├── components/          # 通用组件
│   │   └── layouts/             # 布局
│   └── vite.config.ts           # Vite 配置 + API 代理
├── configs/                     # 配置文件模板
├── docker/                      # Dockerfile
├── docker-compose.yaml          # 完整编排
└── Makefile                     # 构建命令
```

## API 文档

所有 API 前缀为 `/api/v1`。

### 扫描任务

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/v1/scans/` | 创建扫描任务 |
| `GET` | `/api/v1/scans/` | 列表（分页，支持 `status` 筛选） |
| `GET` | `/api/v1/scans/{id}` | 获取详情 |
| `DELETE` | `/api/v1/scans/{id}` | 取消任务 |
| `POST` | `/api/v1/scans/{id}/retry` | 重试失败/取消的任务 |
| `GET` | `/api/v1/scans/{id}/results` | 获取任务结果（支持 `severity` 筛选） |

**创建扫描请求体**:

```json
{
  "name": "扫描任务名称",
  "targets": ["192.168.1.0/24", "https://example.com"],
  "template_filters": {"tags": ["cve"]},
  "concurrency": {"template": 25, "host": 10},
  "rate_limit": 100,
  "headers": ["Authorization: Bearer xxx"]
}
```

### 扫描结果

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/results/` | 全局结果列表（支持 `severity`/`host`/`template_id` 筛选） |
| `GET` | `/api/v1/results/stats` | 统计（按严重程度、Top 主机） |

### Worker 管理

| 方法 | 路径 | 说明 |
|------|------|------|
| `GET` | `/api/v1/workers/` | 列表 |
| `GET` | `/api/v1/workers/{id}` | 获取详情 |
| `POST` | `/api/v1/workers/{id}/disable` | 禁用 |
| `POST` | `/api/v1/workers/{id}/enable` | 启用 |

### 健康检查

- `GET /healthz`

## 配置

### 环境变量

所有配置项可通过环境变量覆盖，前缀为 `NUCLEI_PLATFORM_`，使用 `_` 分隔层级：

```bash
NUCLEI_PLATFORM_SERVER_PORT=8080
NUCLEI_PLATFORM_DATABASE_DSN=/data/nuclei-platform.db
NUCLEI_PLATFORM_NATS_URL=nats://nats:4222
NUCLEI_PLATFORM_LOG_LEVEL=info
NUCLEI_PLATFORM_DEV=true   # 开发模式，跳过内嵌前端
```

### 配置文件

参考 [configs/server.yaml](configs/server.yaml) 和 [configs/worker.yaml](configs/worker.yaml)。

## 部署

### Docker Compose（生产环境）

```bash
# 构建镜像
docker compose build

# 启动所有服务
docker compose up -d

# 查看日志
docker compose logs -f api

# 扩展 Worker
docker compose up -d --scale worker=5

# 停止
docker compose down
```

### 单机部署

```bash
# 构建
make build

# 复制二进制和配置到目标机器
scp bin/server user@host:/opt/nuclei-platform/
scp configs/server.yaml user@host:/opt/nuclei-platform/

# 启动（需要外部 NATS 服务）
/opt/nuclei-platform/server -config /opt/nuclei-platform/server.yaml
```

Server 二进制内嵌了前端静态文件，无需额外部署 Web 服务器。

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go, chi, GORM, NATS, zerolog |
| 前端 | React, TypeScript, Ant Design, Vite |
| 数据库 | SQLite |
| 消息队列 | NATS |
| 扫描引擎 | Nuclei v3 |
| 容器化 | Docker, Docker Compose |

## 许可证

MIT
