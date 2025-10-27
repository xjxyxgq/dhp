-- 用户认证相关表结构

-- ----------------------------
-- 用户表
-- ----------------------------
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `deleted_at` datetime DEFAULT NULL,
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `password_hash` varchar(255) DEFAULT NULL COMMENT '密码哈希（CAS模式下为空）',
  `email` varchar(100) DEFAULT NULL COMMENT '邮箱',
  `display_name` varchar(100) DEFAULT NULL COMMENT '显示名称',
  `is_active` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否激活',
  `is_admin` tinyint(1) NOT NULL DEFAULT '0' COMMENT '是否管理员',
  `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
  `login_source` enum('local', 'cas') NOT NULL DEFAULT 'local' COMMENT '登录来源',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  KEY `idx_email` (`email`),
  KEY `idx_is_active` (`is_active`),
  KEY `idx_login_source` (`login_source`),
  KEY `idx_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

-- ----------------------------
-- 用户会话表
-- ----------------------------
DROP TABLE IF EXISTS `user_sessions`;
CREATE TABLE `user_sessions` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  `user_id` int(10) unsigned NOT NULL COMMENT '用户ID',
  `session_token` varchar(255) NOT NULL COMMENT '会话令牌',
  `cas_ticket` varchar(255) DEFAULT NULL COMMENT 'CAS票据',
  `expires_at` datetime NOT NULL COMMENT '过期时间',
  `ip_address` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` text COMMENT '用户代理',
  `is_active` tinyint(1) NOT NULL DEFAULT '1' COMMENT '是否有效',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_session_token` (`session_token`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_expires_at` (`expires_at`),
  KEY `idx_is_active` (`is_active`),
  CONSTRAINT `fk_user_sessions_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户会话表';

-- 插入默认管理员用户（密码：admin123）
INSERT INTO `users` (`username`, `password_hash`, `email`, `display_name`, `is_admin`) 
VALUES ('admin', '$2a$10$rV8O4w3KUmFE3/W4zz2YBOuyD96FYtDaR4Oa4IB/piNEb0QCl6XhW', 'admin@example.com', '系统管理员', 1);

-- 插入测试用户
INSERT INTO `users` (`username`, `password_hash`, `email`, `display_name`) 
VALUES ('test', '$2a$10$rV8O4w3KUmFE3/W4zz2YBOuyD96FYtDaR4Oa4IB/piNEb0QCl6XhW', 'test@example.com', '测试用户');