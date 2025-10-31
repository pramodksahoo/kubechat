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
  updateInFlight: boolean;
  lastPlanSnapshot?: PlanRecord | null;
}

const initialState: PlanPreviewState = {
  isOpen: false,
  loading: false,
  error: null,
  plan: null,
  dryRunInFlight: false,
  executeInFlight: false,
  updateInFlight: false,
  lastPlanSnapshot: undefined,
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
        revisions: data.revisions,
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

interface UpdatePlanPayload {
  planId: string;
  targetNamespace?: string;
  labels?: Record<string, string>;
  replicaOverrides?: Record<string, number>;
  updatedBy?: string;
}

const updatePlanParameters = createAsyncThunk<PlanRecord, UpdatePlanPayload>(
  "planPreview/updatePlanParameters",
  async ({ planId, targetNamespace, labels, replicaOverrides, updatedBy }, thunkAPI) => {
    try {
      const payload: Record<string, unknown> = {};
      if (typeof targetNamespace === "string") {
        payload.targetNamespace = targetNamespace;
      }
      if (labels) {
        payload.labels = labels;
      }
      if (replicaOverrides) {
        payload.replicaOverrides = replicaOverrides;
      }
      if (updatedBy) {
        payload.updatedBy = updatedBy;
      }

      const response = await kcFetch(`${API_VERSION}/${PLANS_ENDPOINT}/${planId}`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

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
      state.lastPlanSnapshot = undefined;
      state.updateInFlight = false;
    },
    setPlanRecord: (state, action: PayloadAction<PlanRecord | null>) => {
      state.plan = action.payload;
      state.lastPlanId = action.payload?.plan.id;
      state.lastPlanSnapshot = undefined;
      state.updateInFlight = false;
    },
    applyOptimisticPlanUpdate: (state, action: PayloadAction<PlanRecord>) => {
      if (state.plan) {
        state.lastPlanSnapshot = JSON.parse(JSON.stringify(state.plan)) as PlanRecord;
      } else {
        state.lastPlanSnapshot = null;
      }
      state.plan = action.payload;
      state.lastPlanId = action.payload.plan.id;
      state.updateInFlight = true;
    },
    rollbackPlanUpdate: (state) => {
      if (state.lastPlanSnapshot) {
        state.plan = state.lastPlanSnapshot;
        state.lastPlanId = state.lastPlanSnapshot.plan.id;
      }
      state.lastPlanSnapshot = undefined;
      state.updateInFlight = false;
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
      state.lastPlanSnapshot = undefined;
      state.updateInFlight = false;
    });
    builder.addCase(fetchPlanById.rejected, (state, action) => {
      state.loading = false;
      state.error = action.payload as RawRequestError;
      state.plan = null;
      state.isOpen = true;
      state.lastPlanSnapshot = undefined;
      state.updateInFlight = false;
    });

    builder.addCase(updatePlanParameters.pending, (state) => {
      state.updateInFlight = true;
      state.error = null;
    });
    builder.addCase(updatePlanParameters.fulfilled, (state, action) => {
      state.plan = action.payload;
      state.lastPlanId = action.payload.plan.id;
      state.updateInFlight = false;
      state.lastPlanSnapshot = undefined;
      state.error = null;
    });
    builder.addCase(updatePlanParameters.rejected, (state, action) => {
      if (state.lastPlanSnapshot) {
        state.plan = state.lastPlanSnapshot;
        state.lastPlanId = state.lastPlanSnapshot.plan.id;
      }
      state.lastPlanSnapshot = undefined;
      state.updateInFlight = false;
      state.error = action.payload as RawRequestError;
    });
  },
});

export default planPreviewSlice.reducer;
export const { closePlan, setPlanRecord, applyOptimisticPlanUpdate, rollbackPlanUpdate } = planPreviewSlice.actions;
export { createPlanFromPrompt, fetchPlanById, updatePlanParameters, initialState };
