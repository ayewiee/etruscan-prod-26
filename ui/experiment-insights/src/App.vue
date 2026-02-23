<template>
  <div class="app-shell">
    <header class="app-header">
      <div class="app-header-inner">
        <div class="app-title">
          <h1>Experiment Insights</h1>
        </div>
        <div class="auth-panel">
          <span class="muted small-label">JWT token</span>
          <div class="auth-input-row">
            <input
              v-model="jwt"
              class="jwt-input"
              type="password"
              placeholder="Paste JWT here…"
            />
            <button class="btn btn-sm" @click="applyJwt">
              Apply
            </button>
          </div>
        </div>
      </div>
    </header>
    <main class="app-main">
      <div class="app-main-inner">
        <router-view />
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { setAuthToken } from "./api/client";

const jwt = ref("");

const STORAGE_KEY = "experiment_insights_jwt";

function applyJwt() {
  const token = jwt.value.trim();
  if (token) {
    localStorage.setItem(STORAGE_KEY, token);
    setAuthToken(token);
  } else {
    localStorage.removeItem(STORAGE_KEY);
    setAuthToken(null);
  }
}

onMounted(() => {
  const saved = localStorage.getItem(STORAGE_KEY);
  if (saved) {
    jwt.value = saved;
    setAuthToken(saved);
  }
});
</script>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI",
    sans-serif;
  background: #0f172a;
  color: #e5e7eb;
}

.app-header {
  padding: 0.85rem 1.75rem;
  border-bottom: 1px solid rgba(148, 163, 184, 0.3);
  background: radial-gradient(circle at top left, #1d4ed8, #020617 55%);
  backdrop-filter: blur(12px);
}

.app-header-inner {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1.5rem;
}

.app-title {
  display: flex;
  align-items: center;
  gap: 0.6rem;
}

.app-header h1 {
  margin: 0;
  font-size: 1.1rem;
  font-weight: 600;
}

.app-main {
  flex: 1;
  padding: 1.5rem 1.75rem 1.75rem;
}

.app-main-inner {
  max-width: 1200px;
  margin: 0 auto;
}

.auth-panel {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 0.25rem;
}

.auth-input-row {
  display: flex;
  align-items: center;
  gap: 0.35rem;
}

.jwt-input {
  min-width: 260px;
  max-width: 360px;
  background: rgba(15, 23, 42, 0.8);
  border-radius: 999px;
  border: 1px solid rgba(148, 163, 184, 0.8);
  color: #e5e7eb;
  padding: 0.35rem 0.75rem;
  font-size: 0.78rem;
}

.jwt-input::placeholder {
  color: rgba(148, 163, 184, 0.9);
}

.small-label {
  font-size: 0.7rem;
  text-transform: uppercase;
  letter-spacing: 0.08em;
}

.btn-sm {
  padding: 0.25rem 0.6rem;
  font-size: 0.7rem;
}
</style>

