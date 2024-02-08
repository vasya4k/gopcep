<script setup>
import { ref } from 'vue';

import AppMenuItem from './AppMenuItem.vue';

const model = ref([
    {
        label: 'Inventory',
        icon: 'pi pi-fw pi-clone',
        items: [{ label: 'Routers', icon: 'pi pi-fw pi-user-edit', to: '/' }]
    },
    {
        label: 'PCEP',
        icon: 'pi pi-fw pi-clone',
        items: [
            {
                label: 'Sessions',
                icon: 'pi pi-fw pi-user-edit',
                to: '/sessions'
            }
        ]
    },
    {
        label: 'LSP',
        icon: 'pi pi-fw pi-clone',
        items: [
            { label: 'Network', icon: 'pi pi-fw pi-user-edit', to: '/netlsps' },
            {
                label: 'Controller',
                icon: 'pi pi-fw pi-calendar',
                to: '/ctrlsps'
            }
        ]
    },
    {
        label: 'Topology',
        icon: 'pi pi-fw pi-clone',
        items: [{ label: 'BGP-LS', icon: 'pi pi-fw pi-user-edit', to: '/bgpls' }]
    }
]);
</script>

<template>
    <ul class="layout-menu">
        <template v-for="(item, i) in model" :key="item">
            <app-menu-item v-if="!item.separator" :item="item" :index="i"></app-menu-item>
            <li v-if="item.separator" class="menu-separator"></li>
        </template>
    </ul>
</template>

<style lang="scss" scoped></style>
