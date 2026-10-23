-- ----------------------------------------------------------------------------
-- news_ai_agent 数据库初始化脚本
-- 新闻 AI Agent: 检索 -> 抓取 -> GLM 生成 -> 公众号草稿 -> 人工确认发布
-- 与 GORM AutoMigrate 保持一致, 可任选一种方式建表
-- ----------------------------------------------------------------------------
CREATE DATABASE IF NOT EXISTS `news_ai_agent` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
USE `news_ai_agent`;

-- ----------------------------------------------------------------------------
-- 新闻任务表(任务状态机)
-- task_status: 0=待执行 10=搜索中 20=抓取正文中 30=AI生成中 40=创建草稿中
--              50=待人工确认 60=发布中 70=已发布 80=失败 90=已取消
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `news_task` (
  `id`                      BIGINT       NOT NULL AUTO_INCREMENT COMMENT '主键',
  `task_no`                 VARCHAR(64)  NOT NULL DEFAULT ''     COMMENT '任务编号(NT+日期+序号)',
  `topic`                   VARCHAR(32)  NOT NULL DEFAULT ''     COMMENT '新闻领域(ai=人工智能 tech=科技 internet=互联网)',
  `keywords`                VARCHAR(512) NOT NULL DEFAULT ''     COMMENT '本次检索关键词, 逗号分隔',
  `task_status`             TINYINT      NOT NULL DEFAULT 0      COMMENT '任务状态, 见状态机定义',
  `stage_msg`               VARCHAR(1024) NOT NULL DEFAULT ''    COMMENT '当前阶段说明',
  `err_msg`                 VARCHAR(1024) NOT NULL DEFAULT ''    COMMENT '失败原因',
  `retry_count`             INT          NOT NULL DEFAULT 0      COMMENT '已重试次数',
  `llm_model`               VARCHAR(64)  NOT NULL DEFAULT ''     COMMENT '生成使用的模型',
  `llm_prompt_tokens`       INT          NOT NULL DEFAULT 0      COMMENT '提示词token',
  `llm_completion_tokens`   INT          NOT NULL DEFAULT 0      COMMENT '补全token',
  `created_at`              TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`              TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_no` (`task_no`),
  KEY `idx_status` (`task_status`),
  KEY `idx_topic_status` (`topic`, `task_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='新闻任务表';

-- ----------------------------------------------------------------------------
-- 新闻源文章表(检索+正文抓取结果, url_md5 全局去重)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `news_source_article` (
  `id`            BIGINT        NOT NULL AUTO_INCREMENT COMMENT '主键',
  `task_id`       BIGINT        NOT NULL DEFAULT 0      COMMENT '所属任务id',
  `provider`      VARCHAR(32)   NOT NULL DEFAULT ''     COMMENT '检索来源(google/bing)',
  `url`           VARCHAR(1024) NOT NULL DEFAULT ''     COMMENT '原文链接',
  `url_md5`       CHAR(32)      NOT NULL DEFAULT ''     COMMENT '链接md5, 全局去重键',
  `title`         VARCHAR(512)  NOT NULL DEFAULT ''     COMMENT '标题',
  `snippet`       VARCHAR(2048) NOT NULL DEFAULT ''     COMMENT '摘要(RSS description)',
  `author`        VARCHAR(128)  NOT NULL DEFAULT ''     COMMENT '作者',
  `source_name`   VARCHAR(128)  NOT NULL DEFAULT ''     COMMENT '媒体名',
  `published_at`  DATETIME      NULL                    COMMENT '发布时间',
  `content`       MEDIUMTEXT                             COMMENT '抓取的正文(纯文本)',
  `content_status` TINYINT      NOT NULL DEFAULT 0      COMMENT '正文状态 0=未抓取 1=成功 2=失败',
  `content_len`   INT           NOT NULL DEFAULT 0      COMMENT '正文字数',
  `selected`      TINYINT       NOT NULL DEFAULT 0      COMMENT '是否入选生成素材 0=否 1=是',
  `created_at`    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`    TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_url_md5` (`url_md5`),
  KEY `idx_task_id` (`task_id`),
  KEY `idx_published_at` (`published_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='新闻源文章表';

-- ----------------------------------------------------------------------------
-- AI 生成文章表(GLM 产出, 一条任务一条)
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `news_article` (
  `id`         BIGINT        NOT NULL AUTO_INCREMENT COMMENT '主键',
  `task_id`    BIGINT        NOT NULL DEFAULT 0      COMMENT '所属任务id',
  `title`      VARCHAR(256)  NOT NULL DEFAULT ''     COMMENT '文章标题',
  `digest`     VARCHAR(512)  NOT NULL DEFAULT ''     COMMENT '摘要(公众号 digest)',
  `content`    MEDIUMTEXT                             COMMENT '公众号正文HTML',
  `tags`       VARCHAR(256)  NOT NULL DEFAULT ''     COMMENT '标签, 逗号分隔',
  `cover_url`  VARCHAR(1024) NOT NULL DEFAULT ''     COMMENT '封面图URL',
  `llm_model`  VARCHAR(64)   NOT NULL DEFAULT ''     COMMENT '生成模型',
  `created_at` TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_id` (`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI生成文章表';

-- ----------------------------------------------------------------------------
-- 公众号草稿表(草稿创建+发布状态跟踪)
-- publish_status: 0=未发布 1=草稿已创建 2=发布中 3=已发布 4=发布失败
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `news_draft` (
  `id`              BIGINT        NOT NULL AUTO_INCREMENT COMMENT '主键',
  `task_id`         BIGINT        NOT NULL DEFAULT 0      COMMENT '所属任务id',
  `article_id`      BIGINT        NOT NULL DEFAULT 0      COMMENT '文章id',
  `media_id`        VARCHAR(128)  NOT NULL DEFAULT ''     COMMENT '草稿media_id',
  `publish_id`      VARCHAR(128)  NOT NULL DEFAULT ''     COMMENT '发布任务id',
  `publish_status`  TINYINT       NOT NULL DEFAULT 0      COMMENT '发布状态, 见表注释',
  `article_url`     VARCHAR(1024) NOT NULL DEFAULT ''     COMMENT '发布后文章链接',
  `err_msg`         VARCHAR(1024) NOT NULL DEFAULT ''     COMMENT '失败原因',
  `is_mock`         TINYINT       NOT NULL DEFAULT 0      COMMENT '1=未配置公众号时的mock草稿',
  `created_at`      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at`      TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_task_id` (`task_id`),
  KEY `idx_publish_status` (`publish_status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='公众号草稿表';
