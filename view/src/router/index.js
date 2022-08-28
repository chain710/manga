import Vue from "vue";
import VueRouter from "vue-router";
import HomeView from "@/views/HomeView.vue";
import HomeDashboard from "@/views/HomeDashboard.vue";
Vue.use(VueRouter);

const routes = [
  { path: "*", redirect: { name: "dashboard" } },
  {
    name: "home",
    path: "/",
    component: HomeView,
    redirect: { name: "dashboard" },
    children: [
      { name: "dashboard", path: "dashboard", component: HomeDashboard },
      {
        name: "libraries",
        path: "libraries/:libraryID",
        component: () => import("../views/HomeLibraries.vue"),
        props: true,
      },
      {
        name: "book",
        path: "book/:bookID(\\d+)",
        component: () => import("../views/HomeBook.vue"),
        props: true,
      },
    ],
  },
  {
    path: "/read/:volumeID(\\d+)",
    name: "read",
    component: () => import("../views/ReaderView.vue"),
    props: true,
  },
  {
    path: "/play",
    name: "play",
    component: () => import("../views/MyPlay.vue"),
    props: true,
  },
];

const router = new VueRouter({
  routes,
});

export default router;
