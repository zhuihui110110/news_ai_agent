# 新闻 AI Agent (newsserver)

工程结构实现的新闻 AI Agent：**直接调用大模型 API（智谱 GLM，默认模型 `glm-5.3`）**，由 Go 编排全流程：

```
新闻检索(Google/Bing RSS) -> 正文抓取 -> GLM 生成公众号文章 -> 自动创建草稿 -> 人工确认发布(freepublish)
```

新闻领域：`livelihood`(民生) / `economy`(经济) / `ai`(人工智能) / `tech`(科技)。

## 目录结构

```
cmd/newsserver/                 程序入口
configs/
  ├── newsserver.yml            服务配置(所有项可被 NEWS server 前缀环境变量覆盖)
  ├── schema/news_ai_agent.sql  建库建表 SQL(也可由 mysql.migrate=true 自动建表)
  └── assets/default_cover.jpg  默认封面素材
internal/
  ├── newsserver/               服务层
  │   ├── app.go/run.go/server.go/task.go   应用生命周期/装配/后台任务启动
  │   ├── soptions/options.go   配置聚合
  │   ├── sconfig/config.go     配置包装
  │   ├── application/          后台 Monitor: 任务状态机流水线 + 发布轮询
  │   ├── app/message/          Req/Resp DTO
  │   ├── app/service/          业务 Service(任务管理/草稿发布)
  │   └── controller/v1/        REST API(task/draft 两组路由, base 公共应答)
  ├── domain/infrastructure/mysql/
  │   ├── constants/            表名 + 任务状态机定义
  │   ├── model/                GORM 模型(4 张表)
  │   ├── repository/           仓储接口
  │   └── factory/              仓储实现(单例工厂, CAS 状态流转)
  └── innerpkg/
      ├── llm(->pkg/llm)        GLM 直连客户端(pkg/llm)
      ├── news_client/          新闻检索(google/bing RSS) + 正文抓取(编码自适应)
      ├── wechat_client/        公众号 API(token/素材/草稿/freepublish)
      ├── options/              各组件配置项
      └── server/               通用 HTTP 服务器
pkg/
  ├── app/      cobra+viper 应用框架
  ├── llm/      GLM(OpenAI 兼容 /chat/completions)客户端
  └── log/      zap 日志封装
```

## 任务状态机

```
0 pending(待执行)
 └> 10 searching(搜索新闻) -> 20 fetching(抓取正文) -> 30 generating(GLM生成)
     -> 40 draft_creating(创建草稿) -> 50 awaiting_confirm(待人工确认)
     -> 60 publishing(发布中) -> 70 published(已发布, 终态)

任意执行态 -> 80 failed(失败, 冷却后自动重试回 pending, 超过 task.max-retries 停止)
任意未终态 -> 90 cancelled(人工取消, 终态)
```

流转全部通过 `WHERE id=? AND task_status=from` 的 CAS 更新（`factory/news_task.go TransStatus`），多实例部署下不会重复执行。

## 快速开始

```bash
# 1. 建库(或依赖启动时自动建表)
mysql -h127.0.0.1 -uroot -p < configs/schema/news_ai_agent.sql

# 2. 配置(改 configs/newsserver.yml 的 mysql.password; GLM Key 建议走环境变量)
export LLM_API_KEY="你的智谱API Key"

# 3. 运行
make run     # 等价 go run ./cmd/newsserver --config=configs/newsserver.yml
```

说明：
- `wechat.enabled=false` 时草稿走 **mock**（不出真实草稿），全流程可本地跑通；
- 启用公众号需填 `app-id/app-secret`，并在公众号后台把服务器 IP 加入白名单，封面图为
  `configs/assets/default_cover.jpg`（可自行替换，需为正式图片）。

## REST API（前缀 /newsserver/api）

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| POST | /v1/task/create | 创建任务 `{"topics":["economy"],"keywords":["宏观经济"]}` |
| GET | /v1/task/list?topic=economy&status=50&page=1 | 任务分页列表 |
| GET | /v1/task/detail?id=1 | 任务详情(源新闻+生成文章+草稿) |
| POST | /v1/task/retry | 人工重试 `{"id":1}` |
| POST | /v1/task/cancel | 人工取消 `{"id":1}` |
| GET | /v1/draft/list?publish_status=1 | 草稿列表(人工审核) |
| POST | /v1/draft/publish | **人工确认发布** `{"id":1}` |

## 人工确认发布流程

1. 任务流转到 `50 awaiting_confirm`，草稿箱已出现草稿（公众号后台可见）；
2. 人工审核：`GET /v1/draft/list` 查看标题/摘要，或到公众号后台预览全文；
3. 确认：`POST /v1/draft/publish {"id":1}` 提交 freepublish；
4. 发布 Monitor 轮询结果，成功后任务流转 `70 published` 并回填文章链接。

## 大模型直连说明

- 协议：OpenAI 兼容 `POST {llm.base-url}/chat/completions`，默认 `https://open.bigmodel.cn/api/paas/v4`；
- 模型：`llm.model` 默认 `glm-5.3`，可随时切换其它 GLM/OpenAI 兼容端点；
- 鉴权：`Authorization: Bearer $LLM_API_KEY`（或配置 `llm.api-key`）；
- 输出：要求模型返回严格 JSON（title/digest/tags/content_html），客户端做 `<think>` 清洗与围栏容错解析。
