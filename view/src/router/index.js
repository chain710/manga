import VueRouter from "vue-router";
import HomeView from "@/views/HomeView.vue";
import HomeDashboard from "@/views/HomeDashboard.vue";
import { useHub } from "@/plugins/hub";

async function noLibraryGuard(to, from, next) {
  console.debug(`before each`, to, from);
  const hub = useHub();
  if (!hub.isReady) {
    try {
      await hub.init();
    } catch (error) {
      console.error("init hub error", error);
      next(); // TODO consider next error page?
      return;
    }
  }

  if (to.name != "welcome" && hub.libraries.length == 0) {
    next({ name: "welcome" });
  } else {
    next();
  }
}

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
      {
        name: "welcome",
        path: "welcome",
        component: () => import("../views/HomeWelcome.vue"),
      },
    ],
  },
  {
    path: "/read/:volumeID(\\d+)",
    name: "read",
    component: () => import("../views/ReaderView.vue"),
    props: true,
  },
  // {
  //   path: "/play",
  //   name: "play",
  //   component: () => import("../views/MyPlay.vue"),
  //   props: true,
  // },
];

const router = new VueRouter({
  routes,
});

router.beforeEach(noLibraryGuard);

export default router;
