# VC Lab Platform

[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

一个功能完善的一体化 DevOps 管理平台，专为实验室和开发团队设计，提供从资源申请到服务运维的全生命周期管理能力。

## ✨ 核心特性

### 🚀 平台能力
- **一体化 DevOps 平台** - 整合开发、测试、部署、运维全流程
- **高性能后端** - 基于 Go 语言构建的轻量级后端服务
- **离线部署支持** - 完全支持纯离线环境部署，无需外网依赖
- **前后端分离** - 现代化架构设计，易于扩展和维护

### 🔐 安全与权限
- **SSO 单点登录** - 统一身份认证，一次登录全平台通行
- **细粒度权限管理** - 基于 RBAC 的权限控制体系
- **安全审计** - 完整的操作审计日志记录

### 📊 监控与运维
- **监控告警** - 对接 Prometheus + Alertmanager，实时监控系统状态
- **日志查询** - 对接 Elasticsearch，支持全文检索和日志分析
- **WebShell** - 集成 vc-jump，安全便捷的远程终端访问
- **CMDB** - 配置管理数据库，统一管理所有资源

### 🔄 流程与自动化
- **工作流引擎** - 灵活可定制的审批和自动化流程
- **服务管理** - 应用服务的全生命周期管理
- **流程编排** - 可视化流程设计，支持复杂业务场景

### 💻 基础设施管理
- **自动化资源申请** - 一键申请各种环境的计算资源
- **多云支持** - 通过 Terraform Provider 对接多种虚拟化平台
  - PVE (Proxmox Virtual Environment)
  - VMware vSphere
  - OpenStack
  - 公有云（AWS、阿里云、腾讯云等）
- **自定义规格** - 灵活配置 CPU、内存、存储等资源规格

## 🏗️ 技术架构

### 后端技术栈
- **语言框架**: Go 1.21+
- **数据库**: MySQL 8.0+
- **缓存**: Redis
- **消息队列**: RabbitMQ / Kafka（可选）

### 前端技术栈
- **框架**: React 18+
- **样式方案**: TailwindCSS 3.0+
- **状态管理**: Redux Toolkit / Zustand
- **构建工具**: Vite
- **UI 组件**: Headless UI / Radix UI

### 集成组件
- **监控**: Prometheus + Grafana
- **告警**: Alertmanager
- **日志**: Elasticsearch + Kibana
- **堡垒机**: [vc-jump](https://github.com/Veritas-Calculus/vc-jump) - WebShell 安全访问
- **基础设施即代码**: Terraform

## 📦 快速开始

### 环境要求
- Go 1.21 或更高版本
- MySQL 8.0+
- Redis 6.0+
- Node.js 18+ (前端开发)

### 安装部署

#### 1. 克隆项目
```bash
git clone https://github.com/Veritas-Calculus/vc-lab-platform.git
cd vc-lab-platform

# 初始化并更新 submodules（包含 vc-jump）
git submodule init
git submodule update --recursive
```

#### 2. 配置数据库
```bash
# 创建数据库
mysql -u root -p
CREATE DATABASE vc_lab CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

# 导入初始化脚本
mysql -u root -p vc_lab < deploy/sql/init.sql
```

#### 3. 配置文件
```bash
# 复制配置文件模板
cp config/config.example.yaml config/config.yaml

# 编辑配置文件，填写数据库连接等信息
vim config/config.yaml
```

#### 4. 启动后端服务
```bash
# 安装依赖
go mod download

# 编译
go build -o vc-lab-platform cmd/main.go

# 运行
./vc-lab-platform
```

#### 5. 启动前端服务
```bash
cd web
npm install
npm run dev
```

### Docker 部署

```bash
# 使用 docker-compose 一键启动
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 离线部署

详见 [离线部署文档](docs/offline-deployment.md)

## 📖 使用示例

### 申请虚拟机资源

1. 登录平台后，进入「资源管理」模块
2. 点击「申请资源」，选择环境和规格
3. 填写申请信息并提交审批
4. 审批通过后，系统自动通过 Terraform 在 PVE 上创建虚拟机
5. 创建完成后，可通过 WebShell 直接访问

### 配置监控告警

1. 进入「监控告警」模块
2. 配置 Prometheus 数据源
3. 创建告警规则，设置触发条件
4. 配置告警通知渠道（邮件、钉钉、企业微信等）

### 查询日志

1. 进入「日志查询」模块
2. 选择时间范围和服务
3. 输入关键词进行全文检索
4. 支持 Lucene 查询语法

## 📚 文档

- [快速开始](docs/quick-start.md)
- [部署指南](docs/deployment.md)
- [API 文档](docs/api.md)
- [开发指南](docs/development.md)
- [常见问题](docs/faq.md)

## 🤝 贡献指南

欢迎贡献代码、提出问题和功能建议！

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

详见 [贡献指南](CONTRIBUTING.md)

## 📝 开发路线图

- [ ] 支持 Kubernetes 集群管理
- [ ] 集成 CI/CD 流水线
- [ ] 多租户隔离
- [ ] 成本分析与账单管理
- [ ] AI 运维助手
- [ ] 更多云平台对接

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 👥 联系方式

- 项目主页: [https://github.com/Veritas-Calculus/vc-lab-platform](https://github.com/Veritas-Calculus/vc-lab-platform)
- 组织主页: [https://github.com/Veritas-Calculus](https://github.com/Veritas-Calculus)
- 问题反馈: [Issues](https://github.com/Veritas-Calculus/vc-lab-platform/issues)
- 邮件联系: support@vc-lab.com

## 🙏 致谢

感谢以下开源项目的支持：
- [Gin](https://github.com/gin-gonic/gin) - Go Web 框架
- [GORM](https://gorm.io/) - Go ORM 库
- [Terraform](https://www.terraform.io/) - 基础设施即代码工具
- [Prometheus](https://prometheus.io/) - 监控告警系统
- [Elasticsearch](https://www.elastic.co/) - 搜索和分析引擎
- [vc-jump](https://github.com/Veritas-Calculus/vc-jump) - WebShell 堡垒机解决方案
- [React](https://react.dev/) - 用户界面构建库
- [TailwindCSS](https://tailwindcss.com/) - 实用优先的 CSS 框架

---

⭐ 如果这个项目对你有帮助，请给我们一个 Star！
