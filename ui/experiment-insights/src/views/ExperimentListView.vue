<template>
  <div class="list-layout">
    <div class="card list-header-card">
      <div class="list-header-main">
        <div>
          <div class="card-title">Experiments</div>
          <p class="muted">
            Browse experiments and open a detail view to inspect metrics and
            guardrails.
          </p>
        </div>
        <div class="filters">
          <label>
            <span class="muted label">Status</span>
            <select v-model="status">
              <option value="">All</option>
              <option v-for="s in statuses" :key="s" :value="s">
                {{ s }}
              </option>
            </select>
          </label>
          <label>
            <span class="muted label">Outcome</span>
            <select v-model="outcome">
              <option value="">All</option>
              <option v-for="o in outcomes" :key="o" :value="o">
                {{ o }}
              </option>
            </select>
          </label>
        </div>
      </div>
    </div>

    <div class="card">
      <div v-if="loading" class="muted">Loading experiments…</div>
      <div v-else-if="error" class="muted">Error: {{ error }}</div>
      <table v-else class="table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Status</th>
            <th>Audience</th>
            <th>Primary metric</th>
            <th>Outcome</th>
            <th>Updated</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="exp in experiments"
            :key="exp.id"
            @click="openExperiment(exp.id)"
          >
            <td class="exp-name-cell">
              <div class="exp-name">{{ exp.name }}</div>
              <div class="muted exp-sub">
                Flag: {{ exp.flagId }} • ID: {{ exp.id }}
              </div>
            </td>
            <td>
              <span class="tag">{{ exp.status }}</span>
            </td>
            <td>{{ exp.audiencePercentage }}%</td>
            <td>{{ exp.primaryMetricKey || "—" }}</td>
            <td>{{ exp.outcome || "—" }}</td>
            <td>{{ formatDate(exp.updatedAt) }}</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from "vue";
import { useRouter, useRoute } from "vue-router";
import {
  listExperiments,
  type ExperimentSummary
} from "../api/experiments";

const router = useRouter();
const route = useRoute();

const experiments = ref<ExperimentSummary[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const statuses = [
  "DRAFT",
  "ON_REVIEW",
  "APPROVED",
  "DECLINED",
  "LAUNCHED",
  "PAUSED",
  "FINISHED",
  "ARCHIVED"
];
const outcomes = ["ROLLOUT", "ROLLBACK", "NO_EFFECT"];

const status = ref<string>((route.query.status as string) || "");
const outcome = ref<string>((route.query.outcome as string) || "");

async function fetchExperiments() {
  loading.value = true;
  error.value = null;
  try {
    const data = await listExperiments({
      status: status.value || undefined,
      outcome: outcome.value || undefined,
      page: 0,
      size: 50
    });
    experiments.value = data.items;
  } catch (e: any) {
    error.value = e?.message || "Failed to load experiments";
  } finally {
    loading.value = false;
  }
}

function openExperiment(id: string) {
  router.push({ name: "experiment-detail", params: { id } });
}

function formatDate(value: string) {
  const d = new Date(value);
  return d.toLocaleString();
}

watch(
  [status, outcome],
  () => {
    router.replace({
      query: {
        ...route.query,
        status: status.value || undefined,
        outcome: outcome.value || undefined
      }
    });
    fetchExperiments();
  },
  { immediate: false }
);

onMounted(() => {
  fetchExperiments();
});
</script>

<style scoped>
.list-layout {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.list-header-card {
  margin-bottom: 0.25rem;
}

.list-header-main {
  display: flex;
  justify-content: space-between;
  gap: 1.5rem;
  align-items: flex-start;
}

.filters {
  display: flex;
  gap: 1rem;
  flex-wrap: wrap;
}

label {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
  font-size: 0.85rem;
}

select {
  background: rgba(15, 23, 42, 0.9);
  color: #e5e7eb;
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.6);
  padding: 0.3rem 0.7rem;
}

.table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.85rem;
}

.table th,
.table td {
  padding: 0.5rem 0.5rem;
  border-bottom: 1px solid rgba(30, 41, 59, 0.9);
  text-align: left;
}

.table tbody tr {
  cursor: pointer;
}

.table tbody tr:hover {
  background: rgba(15, 23, 42, 0.6);
}

.exp-name-cell {
  max-width: 340px;
}

.exp-name {
  font-weight: 500;
}

.exp-sub {
  font-size: 0.72rem;
}
</style>

