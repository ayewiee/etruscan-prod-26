import { api } from "./client";

export interface VariantMetricValues {
  id: string;
  name: string;
  value: string;
  metrics: Record<string, number>;
}

export interface ExperimentReport {
  experimentId: string;
  from: string;
  to: string;
  variants: VariantMetricValues[];
}

export async function getExperimentReport(params: {
  id: string;
  from: string;
  to: string;
}): Promise<ExperimentReport> {
  const { data } = await api.get<ExperimentReport>(
    `/experiments/${params.id}/report`,
    {
      params: {
        from: params.from,
        to: params.to
      }
    }
  );
  return data;
}

