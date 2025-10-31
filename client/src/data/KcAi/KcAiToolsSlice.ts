import { type ToolSet } from "ai";
import { experimental_createMCPClient } from "@ai-sdk/mcp";
import { createAsyncThunk, createSlice } from "@reduxjs/toolkit";

import { RawRequestError } from "../kcFetch";
import { serializeError } from "serialize-error";

// Module-level variable to store the full tools object (including functions)
let fullTools: ToolSet = {};

type SerializableTool = Record<string, unknown>;
type SerializableToolSet = Record<string, SerializableTool>;

type InitialState = {
  loading: boolean;
  tools: SerializableToolSet;
  error: RawRequestError | null;
};

type FetchKcAiToolsProps = {
  isDev: boolean;
  config: string;
  cluster: string;
};

const initialState: InitialState = {
  loading: false,
  tools: {},
  error: null,
};

const fetchKcAiTools = createAsyncThunk<SerializableToolSet, FetchKcAiToolsProps, { rejectValue: RawRequestError }>('kcAiTools', async ({isDev, config, cluster}: FetchKcAiToolsProps, thunkAPI) => {
  try {
    const hostName = isDev ? 'http://localhost:7080' : window.location.origin;
    const client = await experimental_createMCPClient({
      transport: {
        type: 'sse',
        url: `${hostName}/api/v1/mcp/sse?cluster=${cluster}&config=${config}`,
      },
    });
    const tools = await client.tools();

    // Store the full tools object (with functions) outside Redux
    fullTools = tools as ToolSet;

    // Only store serializable data in Redux
    const serializableTools: SerializableToolSet = {};
    for (const [key, tool] of Object.entries(fullTools)) {
      const serializableTool: SerializableTool = {};
      for (const [prop, value] of Object.entries(tool as Record<string, unknown>)) {
        if (typeof value !== "function") {
          serializableTool[prop] = value;
        }
      }
      serializableTools[key] = serializableTool;
    }

    return serializableTools;
  } catch (e) {
    const serialized = serializeError(e) as { message?: string; code?: number; stack?: string };
    return thunkAPI.rejectWithValue({
      message: serialized.message ?? 'Failed to fetch kcAI tools',
      code: typeof serialized.code === 'number' ? serialized.code : undefined,
      details: serialized.stack,
    });
  }
});

const kcAiToolsSlice = createSlice({
  name: 'tools',
  initialState,
  reducers: {},
  extraReducers: (builder) => {
    builder.addCase(fetchKcAiTools.pending, (state) => {
      state.loading = true;
    });
    builder.addCase(fetchKcAiTools.fulfilled, (state, action) => {
      state.loading = false;
      state.tools = action.payload;
      state.error = null;
    });
    builder.addCase(fetchKcAiTools.rejected, (state, action) => {
      state.loading = false;
      state.tools = {};
      state.error = action.payload ?? { message: 'Failed to fetch kcAI tools' };
    });
  },
});

// Export a getter for the full tools object (with functions)
export const getFullTools = () => fullTools;

export default kcAiToolsSlice.reducer;
export { initialState, fetchKcAiTools };
