-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: MySQL
-- Generated at: 2026-09-10T02:31:17.181Z

CREATE TABLE `files` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `file_sha1` varchar(255) UNIQUE NOT NULL COMMENT '文件的sha1值',
  `file_name` varchar(255) NOT NULL COMMENT '文件名',
  `file_size` bigint NOT NULL DEFAULT 0 COMMENT '文件大小',
  `file_addr` varchar(1024) NOT NULL DEFAULT '' COMMENT '文件存储地址',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '文件状态(0: 正常,1:禁用, 2: 删除)',
  `create_at` timestamp NOT NULL DEFAULT (CURRENT_TIMESTAMP) COMMENT '创建时间',
  `update_at` timestamp NOT NULL DEFAULT (CURRENT_TIMESTAMP) COMMENT '更新时间',
  `extend` int NOT NULL DEFAULT 0 COMMENT '扩展字段',
  `extend2` text COMMENT '扩展字段2'
);

CREATE TABLE `users` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `username` varchar(255) UNIQUE NOT NULL COMMENT '用户名',
  `password` varchar(255) NOT NULL COMMENT '密码',
  `create_at` timestamp NOT NULL DEFAULT (CURRENT_TIMESTAMP) COMMENT '创建时间',
  `update_at` timestamp NOT NULL DEFAULT (CURRENT_TIMESTAMP) COMMENT '更新时间',
  `last_login_at` timestamp COMMENT '最后登录时间',
  `email` varchar(128) UNIQUE NOT NULL COMMENT '邮箱',
  `phone` varchar(128) UNIQUE NOT NULL COMMENT '手机号',
  `email_verified` bool NOT NULL DEFAULT 0 COMMENT '邮箱是否验证',
  `phone_verified` bool NOT NULL DEFAULT 0 COMMENT '手机号是否验证',
  `profile` text COMMENT '用户信息',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '用户状态(0: 正常,1:禁用, 2: 删除)'
);

CREATE INDEX `idx_status` ON `files` (`status`);

CREATE UNIQUE INDEX `idx_file_sha1` ON `files` (`file_sha1`);

CREATE UNIQUE INDEX `idx_username` ON `users` (`username`);

CREATE UNIQUE INDEX `idx_email` ON `users` (`email`);

CREATE UNIQUE INDEX `idx_phone` ON `users` (`phone`);

ALTER TABLE `files` COMMENT = '文件表';
