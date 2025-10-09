export interface MCPTemplate {
  id: string;
  name: string;
  description: string;
  icon: string;
  category: string;
  enabled: boolean;
}

export interface Context7Config {
  url: string;
  apiKey: string;
}

export interface DeepwikiConfig {
  url: string;
  apiKey?: string;
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
    id: 'deepwiki',
    name: 'Deepwiki',
    description: 'Connect to Deepwiki MCP server',
    icon: 'Globe',
    category: 'AI Tools',
    enabled: true,
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

export function generateDeepwikiConfig(config: DeepwikiConfig): string {
  const baseConfig = {
    type: "http",
    url: config.url,
  };

  // If apiKey is provided, add headers with Bearer token
  const mcpConfig = config.apiKey
    ? {
        ...baseConfig,
        headers: {
          Authorization: `Bearer ${config.apiKey}`,
        },
      }
    : baseConfig;

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