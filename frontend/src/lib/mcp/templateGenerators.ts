export interface MCPTemplate {
  id: string;
  name: string;
  description: string;
  icon: string;
  category: string;
  enabled: boolean;
  comingSoon?: boolean;
}

export interface Context7Config {
  url: string;
  apiKey: string;
}

export interface SlackConfig {
  webhookUrl: string;
  channel?: string;
}

export interface DiscordConfig {
  webhookUrl: string;
}

export interface EmailConfig {
  smtpHost: string;
  smtpPort: number;
  smtpUser: string;
  smtpPassword: string;
  from: string;
  to: string;
}

export interface WebhookConfig {
  url: string;
  method?: string;
  headers?: Record<string, string>;
}

export const mcpTemplates: MCPTemplate[] = [
  {
    id: 'context7',
    name: 'Context7',
    description: 'Connect to Context7 MCP server for AI documentation access',
    icon: 'Globe',
    category: 'AI Tools',
    enabled: true,
  },
  {
    id: 'slack',
    name: 'Slack',
    description: 'Send notifications to Slack channels',
    icon: 'MessageSquare',
    category: 'Communication',
    enabled: false,
    comingSoon: true,
  },
  {
    id: 'discord',
    name: 'Discord',
    description: 'Send notifications to Discord channels',
    icon: 'MessageCircle',
    category: 'Communication',
    enabled: false,
    comingSoon: true,
  },
  {
    id: 'email',
    name: 'Email',
    description: 'Send notifications via email',
    icon: 'Mail',
    category: 'Communication',
    enabled: false,
    comingSoon: true,
  },
  {
    id: 'webhook',
    name: 'Webhook',
    description: 'Send notifications to custom webhook endpoints',
    icon: 'Webhook',
    category: 'Integration',
    enabled: false,
    comingSoon: true,
  },
  {
    id: 'opsgenie',
    name: 'OpsGenie',
    description: 'Send alerts to OpsGenie',
    icon: 'AlertTriangle',
    category: 'Monitoring',
    enabled: false,
    comingSoon: true,
  },
];

export function generateContext7Config(config: Context7Config): string {
  const mcpConfig = {
    type: "http",
    url: config.url,
    headers: {
      CONTEXT7_API_KEY: config.apiKey,
    },
  };

  return JSON.stringify(mcpConfig, null, 2);
}

export function generateSlackConfig(config: SlackConfig): string {
  const mcpConfig = {
    type: "slack",
    webhook_url: config.webhookUrl,
    ...(config.channel && { channel: config.channel }),
  };

  return JSON.stringify(mcpConfig, null, 2);
}

export function generateDiscordConfig(config: DiscordConfig): string {
  const mcpConfig = {
    type: "discord",
    webhook_url: config.webhookUrl,
  };

  return JSON.stringify(mcpConfig, null, 2);
}

export function generateEmailConfig(config: EmailConfig): string {
  const mcpConfig = {
    type: "email",
    smtp: {
      host: config.smtpHost,
      port: config.smtpPort,
      user: config.smtpUser,
      password: config.smtpPassword,
    },
    from: config.from,
    to: config.to,
  };

  return JSON.stringify(mcpConfig, null, 2);
}

export function generateWebhookConfig(config: WebhookConfig): string {
  const mcpConfig = {
    type: "webhook",
    url: config.url,
    method: config.method || "POST",
    ...(config.headers && { headers: config.headers }),
  };

  return JSON.stringify(mcpConfig, null, 2);
}

export function getTemplateById(id: string): MCPTemplate | undefined {
  return mcpTemplates.find(template => template.id === id);
}

export function getEnabledTemplates(): MCPTemplate[] {
  return mcpTemplates.filter(template => template.enabled);
}

export function getAllTemplates(): MCPTemplate[] {
  return mcpTemplates;
}