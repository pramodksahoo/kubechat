type kcAIConfiguration = { provider, model, url, apiKey, alias };

type kcAIConfigurations = 'provider' | 'model' | 'url' | 'apiKey' | 'alias';

type kcAIModelResponse = {
  data: {
    id: string;
  }[];
};

type kcAIStoredModel = {
  provider: string;
  url: string;
  apiKey: string;
  model: string;
  apiVersion: string;
  alias: string;
}

type kcAIStoredModels = {
  defaultProvider: string;
  providerCollection: {
    [uuid: string]: kcAIStoredModel;
  }
};

interface ChatMessage {
  id: string
  content: string
  role: "user" | "assistant" | "system"
  timestamp: Date,
  isNotVisible?: boolean,
  completionTokens?: number;
  promptTokens?: number;
  totalTokens?: number;
  error?: boolean;
  reasoning?: string;
  isReasoning?: boolean;
}

type kcAIStoredChatHistory = {
  [clusterConfig: string]: {
    [key: string]: {
      messages: ChatMessage[];
      provider: string;
    };
  }
};

type kcAIModel = {
  label: string;
  value: string;
};

export {
  kcAIModel,
  kcAIConfiguration,
  kcAIConfigurations,
  kcAIModelResponse,
  kcAIStoredModel,
  kcAIStoredModels,
  kcAIStoredChatHistory,
  ChatMessage
};
