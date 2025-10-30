import { API_VERSION, MCP_SERVER_ENDPOINT } from "@/constants";
import { PayloadAction, createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import { kcAIModel, kcAIModelResponse } from "@/types/kcAI/addConfiguration";
import kcFetch, { RawRequestError } from "../kcFetch";

import { formatKcAIModels } from "@/utils/kcAI/CreateConfig";
import { serializeError } from "serialize-error";

type InitialState = {
  loading: boolean;
  kcAiModel: kcAIModel[];
  error: RawRequestError | null;
};

const initialState: InitialState = {
  loading: false,
  kcAiModel: [] as kcAIModel[],
  error: null,
};

type fetchKcAIModelsProps = {
  url: string;
  apiKey: string;
  queryParams: string;
}

const kcAiModels = createAsyncThunk('kcAiModels', ({ apiKey, url, queryParams }: fetchKcAIModelsProps, thunkAPI) => {
  // TODO: Check why // is showing up in build
  const formatedUrl = `${API_VERSION}/${MCP_SERVER_ENDPOINT}`.replace('//', '/');
  return kcFetch(`${formatedUrl}/${url}/models?${queryParams}`, {
    headers: {
      'Authorization': `Bearer ${apiKey}`
    }
  })
    .then((res: kcAIModelResponse) => res ?? {})
    .catch((e: Error) => thunkAPI.rejectWithValue(serializeError(e)));
});

const kcAiModelsSlices = createSlice({
  name: 'kcAiModel',
  initialState,
  reducers: {
    resetKcAiModels: () => {
      return initialState;
    }
  },
  extraReducers: (builder) => {
    builder.addCase(kcAiModels.pending, (state) => {
      state.loading = true;
    });
    builder.addCase(
      kcAiModels.fulfilled,
      (state, action: PayloadAction<kcAIModelResponse>) => {
        state.loading = false;
        state.kcAiModel = formatKcAIModels(action.payload);
        state.error = null;
      },
    );
    builder.addCase(kcAiModels.rejected, (state, action) => {
      state.loading = false;
      state.kcAiModel = [] as kcAIModel[];
      state.error = action.payload as RawRequestError;
    });
  },
});

export default kcAiModelsSlices.reducer;
const { resetKcAiModels } = kcAiModelsSlices.actions;
export { initialState, kcAiModels, resetKcAiModels };
