import { API_VERSION, PLANS_ENDPOINT, PROMPTS_ENDPOINT } from "@/constants";
import { PlanPromptResponse, PlanRecord } from "@/types";
import { PayloadAction, createAsyncThunk, createSlice } from "@reduxjs/toolkit";
import kcFetch, { RawRequestError } from "../kcFetch";
import { serializeError } from "serialize-error";

interface CreatePlanPayload {
  prompt: string;
  clusterHint?: string;
  namespaceHint?: string;
  metadata?: Record<string, string>;
}

interface PlanPreviewState {
  isOpen: boolean;
  loading: boolean;
  error: RawRequestError | null;
  plan: PlanRecord | null;
  lastPlanId?: string;
  dryRunInFlight: boolean;
  executeInFlight: boolean;
}

const initialState: PlanPreviewState = {
  isOpen: false,
  loading: false,
  error: null,
  plan: null,
  dryRunInFlight: false,
  executeInFlight: false,
};

const createPlanFromPrompt = createAsyncThunk<PlanRecord, CreatePlanPayload>(
  "planPreview/createPlanFromPrompt",
  async (payload, thunkAPI) => {
    try {
      const response = await kcFetch(`${API_VERSION}/${PROMPTS_ENDPOINT}`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      const data = (response ?? {}) as PlanPromptResponse;

      const record: PlanRecord = {
        plan: data.plan,
        storedAt: data.storedAt ?? new Date().toISOString(),
        expiresAt: data.expiresAt,
      };

      return record;
    } catch (error) {
      return thunkAPI.rejectWithValue(serializeError(error));
    }
  },
);

const fetchPlanById = createAsyncThunk<PlanRecord, string>(
  "planPreview/fetchPlanById",
  async (planId, thunkAPI) => {
    try {
      const response = await kcFetch(`${API_VERSION}/${PLANS_ENDPOINT}/${planId}`);
      return response as PlanRecord;
    } catch (error) {
      return thunkAPI.rejectWithValue(serializeError(error));
    }
  },
);

const planPreviewSlice = createSlice({
  name: "planPreview",
  initialState,
  reducers: {
    closePlan: (state) => {
      state.isOpen = false;
      state.loading = false;
      state.error = null;
    },
    setPlanRecord: (state, action: PayloadAction<PlanRecord | null>) => {
      state.plan = action.payload;
      state.lastPlanId = action.payload?.plan.id;
    },
  },
  extraReducers: (builder) => {
    builder.addCase(createPlanFromPrompt.pending, (state) => {
      state.loading = true;
      state.isOpen = true;
      state.error = null;
    });
    builder.addCase(createPlanFromPrompt.fulfilled, (state, action) => {
      state.loading = false;
      state.plan = action.payload;
      state.lastPlanId = action.payload.plan.id;
      state.error = null;
      state.isOpen = true;
    });
    builder.addCase(createPlanFromPrompt.rejected, (state, action) => {
      state.loading = false;
      state.error = action.payload as RawRequestError;
      state.plan = null;
      state.isOpen = true;
    });

    builder.addCase(fetchPlanById.pending, (state) => {
      state.loading = true;
      state.isOpen = true;
      state.error = null;
    });
    builder.addCase(fetchPlanById.fulfilled, (state, action) => {
      state.loading = false;
      state.plan = action.payload;
      state.lastPlanId = action.payload.plan.id;
      state.error = null;
      state.isOpen = true;
    });
    builder.addCase(fetchPlanById.rejected, (state, action) => {
      state.loading = false;
      state.error = action.payload as RawRequestError;
      state.plan = null;
      state.isOpen = true;
    });
  },
});

export default planPreviewSlice.reducer;
export const { closePlan, setPlanRecord } = planPreviewSlice.actions;
export { createPlanFromPrompt, fetchPlanById, initialState };
