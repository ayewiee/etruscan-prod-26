import { api } from "./client";

export interface Variant {
  id: string;
  name: string;
  weight: number;
  isControl: boolean;
}

export interface Guardrail {
  id: string;
  metricKey: string;
  threshold: number;
  thresholdDirection: string;
  windowSeconds: number;
  action: string;
}

export interface ExperimentSummary {
  id: string;
  flagId: string;
  name: string;
  status: string;
  audiencePercentage: number;
  primaryMetricKey?: string | null;
  outcome?: string | null;
  createdAt: string;
  updatedAt: string;
}

export interface ExperimentDetail extends ExperimentSummary {
  description?: string | null;
  metricKeys: string[];
  metrics: { metricKey: string; isPrimary: boolean }[];
  guardrails?: Guardrail[] | null;
  variants: Variant[];
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  size: number;
}

export interface ExperimentListFilters {
  status?: string;
  outcome?: string;
  page?: number;
  size?: number;
}

export async function listExperiments(
  filters: ExperimentListFilters
): Promise<PaginatedResponse<ExperimentSummary>> {
  const params: Record<string, unknown> = {};
  if (filters.status) params.status = filters.status;
  if (filters.outcome) params.outcome = filters.outcome;
  if (filters.page != null) params.page = filters.page;
  if (filters.size != null) params.size = filters.size;

  const { data } = await api.get<PaginatedResponse<ExperimentSummary>>(
    "/experiments",
    { params }
  );
  return data;
}

export async function getExperiment(id: string): Promise<ExperimentDetail> {
  const { data } = await api.get<ExperimentDetail>(`/experiments/${id}`);
  return data;
}

