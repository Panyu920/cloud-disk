-- SQL dump generated using DBML (dbml.dbdiagram.io)
-- Database: MySQL
-- Generated at: 2026-09-09T07:40:19.194Z

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

CREATE INDEX `idx_status` ON `files` (`status`);

CREATE UNIQUE INDEX `idx_file_sha1` ON `files` (`file_sha1`);

ALTER TABLE `files` COMMENT = '文件表';
