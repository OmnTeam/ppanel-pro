INSERT IGNORE INTO `subscribe_application` (
  `id`,
  `name`,
  `icon`,
  `description`,
  `scheme`,
  `user_agent`,
  `is_default`,
  `subscribe_template`,
  `output_format`,
  `download_link`,
  `created_at`,
  `updated_at`
) VALUES (
  1001,
  'OmnXT SimNet',
  '',
  'OmnXT SimNet JSON subscription',
  '',
  'OmnXT',
  0,
  '{{ buildOmnxtSimnetConfigs .Proxies .UserInfo .Params | toPrettyJson }}',
  'json',
  '{}',
  NOW(3),
  NOW(3)
);
