CREATE TABLE `subscribe` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `date_time` datetime NOT NULL DEFAULT '2020-01-01 00:00:00' COMMENT '订阅时间',
  `strategy` text DEFAULT NULL COMMENT '订阅策略',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='股票订阅数据';