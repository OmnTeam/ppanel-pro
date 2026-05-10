-- 设备准入控制功能配置
-- device_count_mode: 设备计数模式 ("ip" = 按IP计数, "connection" = 按连接计数)
-- device_admission_enabled: 实时设备准入检查全局开关

INSERT INTO `system` (`category`, `key`, `value`, `type`, `desc`, `created_at`, `updated_at`)
VALUES 
('server', 'device_count_mode', 'ip', 'string', '设备计数模式: ip=同一IP算一个设备, connection=每个活跃连接算一个设备', NOW(), NOW()),
('server', 'device_admission_enabled', 'false', 'bool', '实时设备准入检查开关: 启用后节点将在每次新连接时向面板发起准入检查，需配合 OmnXT Node 使用', NOW(), NOW())
ON DUPLICATE KEY UPDATE `updated_at` = NOW();
