-- 初始化基础数据
-- 认证方法数据
INSERT INTO proxy_auth_method (method, config, enabled, created_at, updated_at) VALUES
('email', '{"platform":"smtp","platform_config":{"host":"","port":0,"user":"","pass":"","from":"","ssl":false},"enable_verify":false,"enable_notify":false,"enable_domain_suffix":false,"domain_suffix_list":"","verify_email_template":"","expiration_email_template":"","maintenance_email_template":"","traffic_exceed_email_template":""}', 1, NOW(3), NOW(3)),
('mobile', '{"platform":"AlibabaCloud","platform_config":{"access":"","secret":"","sign_name":"","endpoint":"","template_code":""},"enable_whitelist":false,"whitelist":[]}', 0, NOW(3), NOW(3)),
('apple', '{"team_id":"","key_id":"","client_id":"","client_secret":"","redirect_url":""}', 0, NOW(3), NOW(3)),
('google', '{"client_id":"","client_secret":"","redirect_url":""}', 0, NOW(3), NOW(3)),
('github', '{"client_id":"","client_secret":"","redirect_url":""}', 0, NOW(3), NOW(3)),
('telegram', '{"bot_token":"","enable_notify":false,"webhook_domain":""}', 0, NOW(3), NOW(3)),
('device', '{"show_ads":false,"only_real_device":false,"enable_security":false,"security_secret":""}', 0, NOW(3), NOW(3));

-- 默认支付方式
INSERT INTO proxy_payment (name, platform, description, icon, domain, config, fee_mode, fee_percent, fee_amount, enable, token) VALUES
('Balance', 'balance', '', '', '', '', 0, 0, 0, 1, '');