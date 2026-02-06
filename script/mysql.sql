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