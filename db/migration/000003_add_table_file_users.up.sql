CREATE TABLE `file_users` (
  `id` bigint PRIMARY KEY AUTO_INCREMENT,
  `username` varchar(255) NOT NULL COMMENT '用户名',
  `file_sha1` varchar(255) NOT NULL COMMENT '文件的sha1值',
  `file_size` bigint NOT NULL DEFAULT 0 COMMENT '文件大小',
  `filename` varchar(255) NOT NULL COMMENT '文件名',
  `create_at` timestamp NOT NULL DEFAULT (CURRENT_TIMESTAMP) COMMENT '创建时间',
  `update_at` timestamp NOT NULL DEFAULT (CURRENT_TIMESTAMP) COMMENT '更新时间',
  `status` tinyint NOT NULL DEFAULT 0 COMMENT '文件用户状态(0: 正常,1:禁用, 2: 删除)'
);

CREATE UNIQUE INDEX `idx_username_filename_file_sha1` ON `file_users` (`username`, `filename`, `file_sha1`);

ALTER TABLE `file_users` ADD FOREIGN KEY (`username`) REFERENCES `users` (`username`);

ALTER TABLE `file_users` ADD FOREIGN KEY (`file_sha1`) REFERENCES `files` (`file_sha1`);
