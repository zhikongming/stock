CREATE TABLE `subscribe` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `date_time` datetime NOT NULL DEFAULT '2020-01-01 00:00:00' COMMENT '订阅时间',
  `strategy` text DEFAULT NULL COMMENT '订阅策略',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='股票订阅数据';

ALTER TABLE `stock_price`
  ADD COLUMN `main_inflow_amount` bigint NOT NULL DEFAULT 0 COMMENT '主力净流入',
  ADD COLUMN `extreme_large_inflow_amount` bigint NOT NULL DEFAULT 0 COMMENT '超大单净流入',
  ADD COLUMN `large_inflow_amount` bigint NOT NULL DEFAULT 0 COMMENT '大单净流入',
  ADD COLUMN `medium_inflow_amount` bigint NOT NULL DEFAULT 0 COMMENT '中单净流入',
  ADD COLUMN `small_inflow_amount` bigint NOT NULL DEFAULT 0 COMMENT '小单净流入';

CREATE TABLE `watcher` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '盯盘名称',
  `stocks` text DEFAULT NULL COMMENT '股票列表',
  `stock_type` tinyint NOT NULL DEFAULT '0' COMMENT '股票类型 0: 东方财富',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '更新时间',
  `status` int NOT NULL DEFAULT '0' COMMENT '状态',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='盯盘配置数据';

CREATE TABLE `cache` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `data_key` varchar(255) NOT NULL DEFAULT '' COMMENT '缓存key',
  `data_type` tinyint NOT NULL DEFAULT '0' COMMENT '缓存类型',
  `date` varchar(255) NOT NULL DEFAULT '2020-01-01' COMMENT '缓存日期',
  `data_value` text DEFAULT NULL COMMENT '缓存值',
  PRIMARY KEY (`id`),
  KEY `idx_type_date` (`data_type`, `date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='缓存数据';

ALTER TABLE `stock_code` 
  ADD COLUMN `is_parsed_price` tinyint NOT NULL DEFAULT '0' COMMENT '是否开启量价分析';

ALTER TABLE `stock_code`
  ADD COLUMN `bd_company_code` varchar(255) NOT NULL DEFAULT '' COMMENT '百度公司代码';

ALTER TABLE `subscribe`
  ADD COLUMN `last_result` tinyint NOT NULL DEFAULT '0' COMMENT '最后分析结果',
  ADD COLUMN `count` int NOT NULL DEFAULT '0' COMMENT '满足最后结果的连续次数';
