import Vue from "vue";
import VueRouter from "vue-router";
import HomeView from "@/views/HomeView.vue";
import HomeDashboard from "@/views/HomeDashboard.vue";

Vue.use(VueRouter);

const routes = [
  { path: "*", redirect: "/" },
  {
    path: "/",
    component: HomeView,
    children: [
      { name: "home", path: "", component: HomeDashboard },
      {
        path: "book/:bid(\\d+)",
        component: () => import("../views/BookView.vue"),
      },
    ],
  },
  {
    // TODO: page params
    path: "/book/:bid(\\d+)/vol/:vid",
    name: "image",
    component: () => import("../views/BookReader.vue"),
  },
];

const router = new VueRouter({
  routes,
});

export default router;
