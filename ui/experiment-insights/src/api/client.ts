import axios from "axios";

const baseURL =
  import.meta.env.VITE_API_BASE_URL?.replace(/\/$/, "") || "/api/v1";

export const api = axios.create({
  baseURL,
  withCredentials: false
});

// Simple JWT helper; adjust if your auth flow differs.
export function setAuthToken(token: string | null) {
  if (token) {
    api.defaults.headers.common.Authorization = `Bearer ${token}`;
  } else {
    delete api.defaults.headers.common.Authorization;
  }
}

