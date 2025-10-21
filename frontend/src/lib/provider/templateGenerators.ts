export interface ProviderTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  enabled: boolean;
  config: Record<string, string>;
}

export const providerTemplates: ProviderTemplate[] = [
  {
    id: "claude-official",
    name: "Claude Official Account",
    description: "Official Claude Code account with OAuth token",
    category: "Official",
    enabled: true,
    config: {
      CLAUDE_CODE_OAUTH_TOKEN: "",
    },
  },
  {
    id: "deepseek",
    name: "DeepSeek",
    description: "DeepSeek API configuration",
    category: "Third Party",
    enabled: true,
    config: {
      ANTHROPIC_BASE_URL: "https://api.deepseek.com/anthropic",
      ANTHROPIC_AUTH_TOKEN: "",
      API_TIMEOUT_MS: "600000",
      ANTHROPIC_MODEL: "deepseek-chat",
      ANTHROPIC_SMALL_FAST_MODEL: "deepseek-chat",
      CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: "1",
    },
  },
  {
    id: "moonshot-cn",
    name: "Moonshot.CN",
    description: "Moonshot.CN API configuration",
    category: "Third Party",
    enabled: true,
    config: {
      ANTHROPIC_BASE_URL: "https://api.moonshot.cn/anthropic",
      ANTHROPIC_AUTH_TOKEN: "",
      ANTHROPIC_MODEL: "kimi-k2-turbo-preview",
      ANTHROPIC_SMALL_FAST_MODEL: "kimi-k2-turbo-preview",
    },
  },
  {
    id: "moonshot-ai",
    name: "Moonshot.AI",
    description: "Moonshot.AI API configuration",
    category: "Third Party",
    enabled: true,
    config: {
      ANTHROPIC_BASE_URL: "https://api.moonshot.ai/anthropic",
      ANTHROPIC_AUTH_TOKEN: "",
      ANTHROPIC_MODEL: "kimi-k2-turbo-preview",
      ANTHROPIC_SMALL_FAST_MODEL: "kimi-k2-turbo-preview",
    },
  },
  {
    id: "bigmodel-cn",
    name: "BigModel.CN",
    description: "ZhiPu BigModel API configuration",
    category: "Third Party",
    enabled: true,
    config: {
      ANTHROPIC_AUTH_TOKEN: "",
      ANTHROPIC_BASE_URL: "https://open.bigmodel.cn/api/anthropic",
      API_TIMEOUT_MS: "3000000",
      CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC: "1",
    },
  },
  {
    id: "z-ai",
    name: "Z.AI",
    description: "Z.AI API configuration",
    category: "Third Party",
    enabled: true,
    config: {
      ANTHROPIC_AUTH_TOKEN: "",
      ANTHROPIC_BASE_URL: "https://api.z.ai/api/anthropic",
      API_TIMEOUT_MS: "3000000",
    },
  },
  {
    id: "proxy-token",
    name: "ClaudeCode Proxy Token",
    description: "Custom proxy server with authentication token",
    category: "Proxy",
    enabled: true,
    config: {
      ANTHROPIC_BASE_URL: "",
      ANTHROPIC_AUTH_TOKEN: "",
    },
  },
  {
    id: "aws-bedrock-keys",
    name: "AWS Bedrock (Access Keys)",
    description: "AWS Bedrock with access key credentials",
    category: "AWS",
    enabled: true,
    config: {
      CLAUDE_CODE_USE_BEDROCK: "1",
      AWS_ACCESS_KEY_ID: "",
      AWS_SECRET_ACCESS_KEY: "",
      AWS_SESSION_TOKEN: "",
    },
  },
  {
    id: "aws-bedrock-bearer",
    name: "AWS Bedrock (Bearer Token)",
    description: "AWS Bedrock with bearer token authentication",
    category: "AWS",
    enabled: true,
    config: {
      CLAUDE_CODE_USE_BEDROCK: "1",
      AWS_BEARER_TOKEN_BEDROCK: "",
    },
  },
];

/**
 * Get all provider templates
 */
export function getAllTemplates(): ProviderTemplate[] {
  return providerTemplates;
}

/**
 * Get enabled provider templates only
 */
export function getEnabledTemplates(): ProviderTemplate[] {
  return providerTemplates.filter((template) => template.enabled);
}

/**
 * Get a specific template by ID
 */
export function getTemplateById(id: string): ProviderTemplate | undefined {
  return providerTemplates.find((template) => template.id === id);
}

/**
 * Get templates grouped by category
 */
export function getTemplatesByCategory(): Record<string, ProviderTemplate[]> {
  const grouped: Record<string, ProviderTemplate[]> = {};

  providerTemplates.forEach((template) => {
    if (!grouped[template.category]) {
      grouped[template.category] = [];
    }
    grouped[template.category].push(template);
  });

  return grouped;
}
