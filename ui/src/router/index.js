import { createRouter, createWebHashHistory } from 'vue-router';
import AppLayout from '@/layout/AppLayout.vue';
import RoutersList from '@/components/RoutersList.vue';
import PCEPSessions from '@/components/PCEPSessions.vue';
import NetworkLSP from '@/components/NetworkLSP.vue';
import ControllerLSP from '@/components/ControllerLSP.vue';
import BGPList from '@/components/BGPList.vue';
import AddLSP from '@/components/AddLSP.vue';
import LSPOverview from '@/components/LSPOverview.vue';

const router = createRouter({
    history: createWebHashHistory(),
    routes: [
        {
            path: '/',
            component: AppLayout,
            children: [
                {
                    path: '/',
                    name: 'Routers',
                    component: RoutersList
                },
                {
                    path: '/sessions',
                    name: 'Sessions',
                    component: PCEPSessions
                },
                {
                    path: '/netlsps',
                    name: 'NetLSPs',
                    component: NetworkLSP
                },
                {
                    path: '/ctrlsps',
                    name: 'CtrLSPs',
                    component: ControllerLSP
                },
                {
                    path: '/bgpls',
                    name: 'BGPLS',
                    component: BGPList
                },
                {
                    path: '/addlsp/:new',
                    name: 'AddLSP',
                    query: {
                        new: true
                      },
                    component: AddLSP
                },
                {
                    path: '/lspoverview',
                    name: 'LSPOverview',
                    component: LSPOverview
                }
            ]
        },
        {
            path: '/landing',
            name: 'landing',
            component: () => import('@/views/pages/Landing.vue')
        },
        {
            path: '/pages/notfound',
            name: 'notfound',
            component: () => import('@/views/pages/NotFound.vue')
        },

        {
            path: '/auth/login',
            name: 'login',
            component: () => import('@/views/pages/auth/Login.vue')
        },
        {
            path: '/auth/access',
            name: 'accessDenied',
            component: () => import('@/views/pages/auth/Access.vue')
        },
        {
            path: '/auth/error',
            name: 'error',
            component: () => import('@/views/pages/auth/Error.vue')
        }
    ]
});

export default router;
