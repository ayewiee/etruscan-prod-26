<template>
  <div v-if="loading" class="muted">Loading experiment…</div>
  <div v-else-if="error" class="muted">Error: {{ error }}</div>
  <div v-else-if="experiment && report" class="layout">
    <section class="layout-main">
      <div class="card hero-card">
        <div class="header">
          <div>
            <div class="card-title">{{ experiment.name }}</div>
            <div class="muted">
              Status: <span class="tag">{{ experiment.status }}</span>
              <span v-if="experiment.outcome" class="tag" style="margin-left: 0.5rem">
                Outcome: {{ experiment.outcome }}
              </span>
            </div>
            <div class="muted" style="margin-top: 0.25rem">
              Audience: {{ experiment.audiencePercentage }}% | Primary metric:
              {{ primaryMetricLabel || "—" }}
            </div>
          </div>
          <div class="time-controls">
            <label>
              <span class="muted">From</span>
              <input type="datetime-local" v-model="fromLocal" />
            </label>
            <label>
              <span class="muted">To</span>
              <input type="datetime-local" v-model="toLocal" />
            </label>
            <button class="btn" @click="reloadReport">Apply</button>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-title">Metrics</div>
        <div class="metric-selector">
          <label v-for="m in experimentMetrics" :key="m.metricKey">
            <input
              type="radio"
              name="metric"
              :value="m.metricKey"
              v-model="selectedMetricKey"
            />
            <span>
              {{ metricLabel(m.metricKey) }}
              <span v-if="m.isPrimary" class="tag" style="margin-left: 0.25rem">
                primary
              </span>
            </span>
          </label>
        </div>

        <div v-if="currentMetricValues" class="metric-layout">
          <div class="metric-table">
            <table>
              <thead>
                <tr>
                  <th>Variant</th>
                  <th>Value</th>
                  <th>Weight</th>
                  <th>Control</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in currentMetricValues" :key="row.id">
                  <td>{{ row.name }}</td>
                  <td>{{ formatMetricValue(row.value, selectedMetricKey) }}</td>
                  <td>{{ row.weight }}%</td>
                  <td>{{ row.isControl ? "Yes" : "No" }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="metric-chart">
            <Bar :data="chartData" :options="chartOptions" />
          </div>
        </div>
        <div v-else class="muted">No metric selected or no data.</div>
      </div>
    </section>

    <aside class="layout-aside">
      <div class="card">
        <div class="card-title">Guardrails</div>
        <div v-if="experiment.guardrails === null || experiment.guardrails?.length == 0" class="muted">
          No guardrails configured.
        </div>
        <ul v-else class="guardrails">
          <li v-for="g in experiment.guardrails" :key="g.id">
            <div>
              <strong>{{ g.metricKey }}</strong>
              <div class="muted">
                {{ g.thresholdDirection }} {{ g.threshold }} over last
                {{ g.windowSeconds }}s → {{ g.action }}
              </div>
            </div>
          </li>
        </ul>
      </div>

      <div class="card">
        <div class="card-title">Export & sharing</div>
        <button class="btn" @click="downloadCsv">Download CSV</button>
        <button class="btn" style="margin-left: 0.5rem" @click="downloadChart">
          Download chart PNG
        </button>
        <div class="muted" style="margin-top: 0.5rem; font-size: 0.75rem">
          Share this URL to keep the same experiment, window, and metric.
        </div>
      </div>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { Bar } from "vue-chartjs";
import {
  Chart as ChartJS,
  BarElement,
  CategoryScale,
  LinearScale,
  Tooltip,
  Legend
} from "chart.js";
import {
  getExperiment,
  type ExperimentDetail
} from "../api/experiments";
import {
  getExperimentReport,
  type ExperimentReport
} from "../api/reports";
import { listMetrics, type Metric } from "../api/metrics";

ChartJS.register(BarElement, CategoryScale, LinearScale, Tooltip, Legend);

const route = useRoute();
const router = useRouter();
const id = route.params.id as string;

const experiment = ref<ExperimentDetail | null>(null);
const report = ref<ExperimentReport | null>(null);
const metricsCatalog = ref<Metric[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);

const selectedMetricKey = ref<string | null>(null);

// from/to in ISO; keep local inputs separately
const fromLocal = ref("");
const toLocal = ref("");

const experimentMetrics = computed(
  () => experiment.value?.metrics ?? []
);

const primaryMetricKey = computed(() => {
  const primary = experimentMetrics.value.find((m) => m.isPrimary);
  return primary?.metricKey || experiment.value?.primaryMetricKey || null;
});

const primaryMetricLabel = computed(() =>
  primaryMetricKey.value ? metricLabel(primaryMetricKey.value) : null
);

function metricLabel(key: string) {
  const m = metricsCatalog.value.find((mm) => mm.key === key);
  return m?.name || key;
}

function parseQueryWindow() {
  const from = route.query.from as string | undefined;
  const to = route.query.to as string | undefined;
  const now = new Date();
  const defaultTo = now;
  const defaultFrom = new Date(now.getTime() - 24 * 60 * 60 * 1000);
  const f = from ? new Date(from) : defaultFrom;
  const t = to ? new Date(to) : defaultTo;
  fromLocal.value = toLocalInputValue(f);
  toLocal.value = toLocalInputValue(t);
}

function toLocalInputValue(d: Date) {
  const iso = d.toISOString();
  return iso.slice(0, 16);
}

function localToISO(local: string) {
  // Treat local as in local timezone and convert to ISO
  const d = new Date(local);
  return d.toISOString();
}

const currentMetricValues = computed(() => {
  if (!experiment.value || !report.value || !selectedMetricKey.value) {
    return null;
  }
  const variants = experiment.value.variants;
  return report.value.variants.map((v) => {
    const def = variants.find((vv) => vv.id === v.id);
    return {
      id: v.id,
      name: v.name,
      value: v.metrics[selectedMetricKey.value!] ?? 0,
      weight: def?.weight ?? 0,
      isControl: def?.isControl ?? false
    };
  });
});

const chartData = computed(() => {
  if (!currentMetricValues.value) {
    return {
      labels: [],
      datasets: []
    };
  }
  return {
    labels: currentMetricValues.value.map((v) => v.name),
    datasets: [
      {
        label: metricLabel(selectedMetricKey.value || ""),
        data: currentMetricValues.value.map((v) => v.value),
        backgroundColor: "#38bdf8"
      }
    ]
  };
});

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      labels: {
        color: "#e5e7eb"
      }
    }
  },
  scales: {
    x: {
      ticks: { color: "#9ca3af" },
      grid: { color: "rgba(148, 163, 184, 0.2)" }
    },
    y: {
      ticks: { color: "#9ca3af" },
      grid: { color: "rgba(148, 163, 184, 0.2)" }
    }
  }
};

async function loadAll() {
  loading.value = true;
  error.value = null;
  try {
    const [exp, metrics] = await Promise.all([
      getExperiment(id),
      listMetrics()
    ]);
    experiment.value = exp;
    metricsCatalog.value = metrics;

    if (!fromLocal.value || !toLocal.value) {
      parseQueryWindow();
    }
    await reloadReport();

    // default selected metric
    if (experimentMetrics.value.length > 0) {
      selectedMetricKey.value =
        primaryMetricKey.value || experimentMetrics.value[0].metricKey;
    }
  } catch (e: any) {
    error.value = e?.message || "Failed to load experiment";
  } finally {
    loading.value = false;
  }
}

async function reloadReport() {
  if (!fromLocal.value || !toLocal.value) return;
  const fromIso = localToISO(fromLocal.value);
  const toIso = localToISO(toLocal.value);

  // keep permalinkable
  router.replace({
    query: {
      ...route.query,
      from: fromIso,
      to: toIso,
      metric: selectedMetricKey.value || undefined
    }
  });

  const r = await getExperimentReport({
    id,
    from: fromIso,
    to: toIso
  });
  report.value = r;
}

function formatMetricValue(value: number, key: string | null) {
  if (!key) return value.toFixed(3);
  const m = metricsCatalog.value.find((mm) => mm.key === key);
  if (!m) return value.toFixed(3);
  if (m.type === "binomial") {
    return (value * 100).toFixed(2) + "%";
  }
  return value.toFixed(3);
}

function downloadCsv() {
  if (!experiment.value || !report.value || !experimentMetrics.value.length) {
    return;
  }

  const rows: string[] = [];
  const header = [
    "metricKey",
    "name",
    "id",
    "value",
    "weight",
    "isControl"
  ];
  rows.push(header.join(","));

  for (const m of experimentMetrics.value) {
    for (const v of report.value.variants) {
      const variantDef = experiment.value.variants.find(
        (vv) => vv.id === v.id
      );
      const value = v.metrics[m.metricKey] ?? 0;
      rows.push(
        [
          m.metricKey,
          `"${v.name}"`,
          v.id,
          value,
          variantDef?.weight ?? "",
          variantDef?.isControl ? "true" : "false"
        ].join(",")
      );
    }
  }

  const blob = new Blob([rows.join("\n")], {
    type: "text/csv;charset=utf-8;"
  });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `experiment-${id}-report.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

function downloadChart() {
  const canvas = document.querySelector(
    ".metric-chart canvas"
  ) as HTMLCanvasElement | null;
  if (!canvas) return;
  const link = document.createElement("a");
  link.href = canvas.toDataURL("image/png");
  link.download = `experiment-${id}-chart.png`;
  link.click();
}

watch(
  () => route.query.metric as string | undefined,
  (metricFromQuery) => {
    if (metricFromQuery) {
      selectedMetricKey.value = metricFromQuery;
    }
  }
);

onMounted(() => {
  parseQueryWindow();
  loadAll();
});
</script>

<style scoped>
.layout {
  display: grid;
  grid-template-columns: minmax(0, 3fr) minmax(260px, 1fr);
  gap: 1.2rem;
}

.layout-main {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.layout-aside {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
}

.hero-card .card-title {
  font-size: 1.05rem;
}

.time-controls {
  display: flex;
  gap: 0.5rem;
  flex-wrap: wrap;
  align-items: flex-end;
}

.time-controls label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.8rem;
}

input[type="datetime-local"] {
  background: rgba(15, 23, 42, 0.9);
  border-radius: 0.5rem;
  border: 1px solid rgba(148, 163, 184, 0.6);
  color: #e5e7eb;
  padding: 0.25rem 0.5rem;
}

.metric-selector {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem 1rem;
  margin-bottom: 0.75rem;
}

.metric-selector label {
  display: flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
}

.metric-layout {
  display: grid;
  grid-template-columns: minmax(0, 1.4fr) minmax(0, 2fr);
  gap: 1rem;
  align-items: stretch;
}

.metric-table table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8rem;
}

.metric-table th,
.metric-table td {
  padding: 0.35rem 0.4rem;
  border-bottom: 1px solid rgba(30, 41, 59, 0.9);
}

.metric-chart {
  min-height: 220px;
}

.guardrails {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  font-size: 0.8rem;
}
</style>

