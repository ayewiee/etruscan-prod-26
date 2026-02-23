import type { RouteRecordRaw } from "vue-router";
import ExperimentListView from "./views/ExperimentListView.vue";
import ExperimentDetailView from "./views/ExperimentDetailView.vue";

export const routes: RouteRecordRaw[] = [
  {
    path: "/",
    redirect: "/experiments"
  },
  {
    path: "/experiments",
    name: "experiments",
    component: ExperimentListView
  },
  {
    path: "/experiments/:id",
    name: "experiment-detail",
    component: ExperimentDetailView,
    props: true
  }
];

