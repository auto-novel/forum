# Forum 论坛服务

[![GPL-3.0](https://img.shields.io/github/license/auto-novel/forum)](https://github.com/auto-novel/forum#license)
[![cd-web](https://github.com/auto-novel/forum/actions/workflows/cd-web.yml/badge.svg)](https://github.com/auto-novel/forum/actions/workflows/cd-web.yml)
[![cd-api](https://github.com/auto-novel/forum/actions/workflows/cd-api.yml/badge.svg)](https://github.com/auto-novel/forum/actions/workflows/cd-api.yml)

提供论坛服务，支持分类、帖子、评论、收藏和外部资源评论等功能。

## 贡献

请务必在编写代码前阅读[贡献指南](https://github.com/auto-novel/forum/blob/main/CONTRIBUTING.md)，感谢所有为本项目做出贡献的人们！

## 部署

> [!WARNING]
> 注意：本项目并不是为了个人部署设计的，不保证所有功能可用和前向兼容。

```bash
# 1. 克隆仓库
git clone https://github.com/auto-novel/forum.git
cd forum

# 2. 生成环境变量配置
cat > .env << EOF
ACCESS_TOKEN_SECRET=$(openssl rand -base64 48)
POSTGRES_PASSWORD=$(openssl rand -base64 48)
EOF

# 3. 启动服务
docker compose up -d
```

启动后，可以通过以下地址访问：

- 用户端：http://localhost:5000/
- 管理端：http://localhost:5000/admin/
