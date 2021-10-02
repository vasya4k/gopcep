import { createRouter, createWebHashHistory } from 'vue-router';
import Dashboard from './components/Dashboard.vue';
import RoutersList from './components/RoutersList.vue';
import PCEPSessions from './components/PCEPSessions.vue';
import NetworkLSP from './components/NetworkLSP.vue';
import ControllerLSP from './components/ControllerLSP.vue';
import BGPList from './components/BGPList.vue';
import AddLSP from './components/AddLSP.vue';

const routes = [
    {
        path: '/',
        name: 'dashboard',
        component: Dashboard,
    },
    {
        path: "/routers",
        name: "Routers",
        component: RoutersList,
    },
    {
        path: "/sessions",
        name: "Sessions",
        component: PCEPSessions,
    },
    {
        path: "/netlsps",
        name: "NetLSPs",
        component: NetworkLSP,
    },
    {
        path: "/ctrlsps",
        name: "CtrLSPs",
        component: ControllerLSP,
    },
    {
        path: "/bgpls",
        name: "BGPLS",
        component: BGPList,
    },
    {
        path: "/addlsp",
        name: "AddLSP",
        props: true,
        component: AddLSP,
    }
];

const router = createRouter({
    history: createWebHashHistory(),
    routes,
});

export default router;
