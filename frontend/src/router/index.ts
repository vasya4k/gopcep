import { createRouter, createWebHashHistory, RouteRecordRaw } from "vue-router";
import Home from "../views/Home.vue";

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    name: "Home",
    component: Home
  },
  {
    path: "/about",
    name: "About",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "about" */ "../views/About.vue")
  },
  {
    path: "/bgp",
    name: "BGP",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "about" */ "../views/BGPList.vue")
  },
  {
    path: "/psessions",
    name: "PCEPSessions",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "about" */ "../views/PCEPSessions.vue")
  },
  {
    path: "/lsps",
    name: "LSPList",
    component: () =>
      import(/* webpackChunkName: "about" */ "../views/LSPList.vue")
  },
  {
    path: "/routers",
    name: "Routers",
    component: () =>
      import(/* webpackChunkName: "about" */ "../views/RoutersList.vue")
  }
];

const router = createRouter({
  history: createWebHashHistory(),
  routes
});

export default router;
