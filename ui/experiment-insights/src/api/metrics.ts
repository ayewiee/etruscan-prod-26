import { api } from "./client";

export interface Metric {
  id: string;
  key: string;
  name: string;
  description?: string | null;
  type: string;
  eventTypeKey?: string;
  aggregationType?: string;
  isGuardrail: boolean;
}

export async function listMetrics(): Promise<Metric[]> {
  const { data } = await api.get<Metric[]>("/metrics");
  return data;
}

