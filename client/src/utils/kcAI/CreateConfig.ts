import { kcAIModelResponse } from "@/types/kcAI/addConfiguration";

const formatKcAIModels = (kcAIModelsResponse: kcAIModelResponse) => {
  return kcAIModelsResponse.data.map(({ id}) => ({
    label: id,
    value: id
  }));
};

export {
  formatKcAIModels
};
