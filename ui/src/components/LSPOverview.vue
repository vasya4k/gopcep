<script>
import { reactive } from 'vue';

export default {
    data() {
        return {
            messages: []
        };
    },
    created() {},
    computed: {
        lsp() {
            let lsps = reactive(this.$store.getters['netlsp/lspToGet']);
            console.log(JSON.stringify(lsps, null, 2));
            return lsps;
        }
    },
    mounted() {},
    methods: {}
};
</script>
<template>
    <div class="grid p-fluid">
        <div class="col-12 lg:col-6 xl:col-3">
            <div class="card mb-0">
                <div class="flex justify-content-between mb-3">
                    <div>
                        <span class="block text-500 font-medium mb-3">Name</span>
                        <div class="text-900 font-medium text-xl">{{ lsp.Name }}</div>
                    </div>
                </div>
                <span class="text-500">Bandwidth: </span>
                <span class="text-green-500 font-medium">{{ lsp.BW }} </span>
                <span class="text-500"> PLSPID: </span>
                <span class="text-green-500 font-medium">{{ lsp.PLSPID }} </span>
            </div>
        </div>
        <div class="col-12 lg:col-6 xl:col-3">
            <div class="card mb-0">
                <div class="flex justify-content-between mb-3">
                    <div>
                        <span class="block text-500 font-medium mb-3">Source</span>
                        <div class="text-900 font-medium text-xl">{{ lsp.Src }}</div>
                    </div>
                </div>
                <span class="text-green-500 font-medium"></span>
                <span class="text-500"></span>
            </div>
        </div>
        <div class="col-12 lg:col-6 xl:col-3">
            <div class="card mb-0">
                <div class="flex justify-content-between mb-3">
                    <div>
                        <span class="block text-500 font-medium mb-3">Destination</span>
                        <div class="text-900 font-medium text-xl">{{ lsp.Dst }}</div>
                    </div>
                </div>
                <span class="text-green-500 font-medium"> </span>
                <span class="text-500"></span>
            </div>
        </div>
        <div class="col-12 lg:col-6 xl:col-3">
            <div class="card mb-0">
                <div class="flex justify-content-between mb-3">
                    <div>
                        <span class="block text-500 font-medium mb-3">Status</span>
                        <span v-if="lsp.Oper == 0" class="customer-badge status-qualified"> Down </span>
                        <span v-if="lsp.Oper == 1" class="customer-badge status-qualified"> Up </span>
                        <span v-if="lsp.Oper == 2" class="customer-badge status-qualified"> ACTIVE </span>
                        <span v-if="lsp.Oper == 3" class="customer-badge status-qualified"> GOING-DOWN </span>
                        <span v-if="lsp.Oper == 4" class="customer-badge status-qualified"> GOING-UP </span>
                        <span v-if="lsp.Oper > 4" class="customer-badge status-qualified"> WEIRD </span>
                    </div>
                </div>
                <span class="text-500">Setup Priority: </span>
                <span class="text-green-500 font-medium">{{ lsp.SetupPrio }} </span>
                <span class="text-500"> Hold Priority: </span>
                <span class="text-green-500 font-medium">{{ lsp.HoldPrio }} </span>
                <span class="text-500"> Admin: </span>
                <span class="text-green-500 font-medium" v-if="lsp.Admin == true"> UP </span>
                <span class="text-green-500 font-medium" v-if="lsp.Delegate == true"> Delegated</span>
            </div>
        </div>

        <div class="col-12 xl:col-6">
            <div class="card">
                <h4>ERO</h4>
                <DataTable :value="lsp.SREROList" :rows="5" :paginator="false" responsiveLayout="scroll">
                    <Column field="SID" header="SID" :sortable="true" style="width: 35%"></Column>
                    <Column field="IPv4NodeID" header="IPv4NodeID" :sortable="true"></Column>
                    <Column header="IP Adj 0" :sortable="true">
                        <template #body="slotProps">
                            {{ slotProps.data.IPv4Adjacency[0] }}
                        </template>
                    </Column>
                    <Column header="IP Adj 1" :sortable="true">
                        <template #body="slotProps">
                            {{ slotProps.data.IPv4Adjacency[1] }}
                        </template>
                    </Column>
                </DataTable>
            </div>
        </div>
    </div>
</template>

<style scoped lang="scss">
.customer-badge {
    border-radius: 2px;
    padding: 0.25em 0.5rem;
    text-transform: uppercase;
    font-weight: 700;
    font-size: 12px;
    letter-spacing: 0.3px;

    &.status-qualified {
        background: #c8e6c9;
        color: #256029;
    }

    &.status-unqualified {
        background: #ffcdd2;
        color: #c63737;
    }

    &.status-negotiation {
        background: #feedaf;
        color: #8a5340;
    }

    &.status-new {
        background: #b3e5fc;
        color: #23547b;
    }

    &.status-renewal {
        background: #eccfff;
        color: #694382;
    }

    &.status-proposal {
        background: #ffd8b2;
        color: #805b36;
    }
}

.product-badge {
    border-radius: 2px;
    padding: 0.25em 0.5rem;
    text-transform: uppercase;
    font-weight: 700;
    font-size: 12px;
    letter-spacing: 0.3px;

    &.status-instock {
        background: #c8e6c9;
        color: #256029;
    }

    &.status-outofstock {
        background: #ffcdd2;
        color: #c63737;
    }

    &.status-lowstock {
        background: #feedaf;
        color: #8a5340;
    }
}

.order-badge {
    border-radius: 2px;
    padding: 0.25em 0.5rem;
    text-transform: uppercase;
    font-weight: 700;
    font-size: 12px;
    letter-spacing: 0.3px;

    &.order-delivered {
        background: #c8e6c9;
        color: #256029;
    }

    &.order-cancelled {
        background: #ffcdd2;
        color: #c63737;
    }

    &.order-pending {
        background: #feedaf;
        color: #8a5340;
    }

    &.order-returned {
        background: #eccfff;
        color: #694382;
    }
}
</style>
