<script>
import { FilterMatchMode } from 'primevue/api';
import { HTTP } from '../service/http';
import { mapActions } from 'vuex';

export default {
    data() {
        return {
            statuses: ['unqualified', 'qualified', 'new', 'negotiation', 'renewal', 'proposal'],
            products: [],
            productDialog: false,
            deleteProductDialog: false,
            deleteProductsDialog: false,
            product: {},
            selectedProducts: null,
            filters: {},
            submitted: false
        };
    },
    productService: null,
    created() {
        this.initFilters();
    },
    mounted() {
        HTTP.get('pceplsps')
            .then((response) => {
                this.products = response.data;
            })
            .catch(function (error) {
                if (error.response) {
                    // The request was made and the server responded with a status code
                    // that falls out of the range of 2xx
                    console.log(error.response.data);
                    console.log(error.response.status);
                    console.log(error.response.headers);
                } else if (error.request) {
                    // The request was made but no response was received
                    // `error.request` is an instance of XMLHttpRequest in the browser and an instance of
                    // http.ClientRequest in node.js
                    console.log(error.request);
                } else {
                    // Something happened in setting up the request that triggered an Error
                    console.log('Error', error.message);
                }
                console.log(error.config);
            });
    },
    methods: {
        ...mapActions('netlsp', {
            itemDetails(dispatch, lsp) {
                dispatch('saveLSP', lsp);
                this.$router.push('lspoverview');
            }
        }),
        findIndexById(id) {
            let index = -1;
            for (let i = 0; i < this.products.length; i++) {
                if (this.products[i].id === id) {
                    index = i;
                    break;
                }
            }

            return index;
        },
        exportCSV() {
            this.$refs.dt.exportCSV();
        },
        initFilters() {
            this.filters = {
                global: { value: null, matchMode: FilterMatchMode.CONTAINS }
            };
        }
    }
};
</script>

<template>
    <div class="p-grid">
        <div class="p-col-12">
            <div>
                <Panel header="Network LSPs" class="p-mt-4">
                    <div class="card">
                        <Toolbar class="p-mb-4">
                            <template v-slot:start>
                                <Button label="Export" icon="pi pi-upload" class="p-button-help" @click="exportCSV($event)" />
                            </template>
                        </Toolbar>

                        <DataTable
                            :value="products"
                            dataKey="Name"
                            :paginator="true"
                            :rows="10"
                            :filters="filters"
                            paginatorTemplate="FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink CurrentPageReport RowsPerPageDropdown"
                            :rowsPerPageOptions="[5, 10, 25]"
                            currentPageReportTemplate="Showing {first} to {last} of {totalRecords} LSPs"
                            responsiveLayout="scroll"
                        >
                            <template #header>
                                <div class="table-header p-d-flex p-flex-column p-flex-md-row p-jc-md-between">
                                    <span class="p-input-icon-left">
                                        <i class="pi pi-search" />
                                        <InputText v-model="filters['global'].value" placeholder="Search..." />
                                    </span>
                                </div>
                            </template>
                            <Column field="Name" header="Name" :sortable="true" style="min-width: 4rem"></Column>
                            <Column field="Src" header="SRC" :sortable="true" style="min-width: 4rem"></Column>
                            <Column field="Dst" header="DST" :sortable="true" style="min-width: 4rem"></Column>
                            <Column field="Admin" header="Admin State" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column field="Oper" header="Oper State" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column field="BW" header="BW" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column field="PLSPID" header="LSP ID" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column field="SRPID" header="SRP ID" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column header="Status">
                                <template #body="slotProps">
                                    <span v-if="slotProps.data.Oper == 0" class="customer-badge status-qualified"> Down </span>
                                    <span v-if="slotProps.data.Oper == 1" class="customer-badge status-qualified"> Up </span>
                                    <span v-if="slotProps.data.Oper == 2" class="customer-badge status-qualified"> ACTIVE </span>
                                    <span v-if="slotProps.data.Oper == 3" class="customer-badge status-qualified"> GOING-DOWN </span>
                                    <span v-if="slotProps.data.Oper == 4" class="customer-badge status-qualified"> GOING-UP </span>
                                    <span v-if="slotProps.data.Oper > 4" class="customer-badge status-qualified"> WEIRD </span>
                                </template>
                            </Column>
                            <Column>
                                <template #body="slotProps">
                                    <Button label="Details" class="p-button-rounded p-button-info mr-2 mb-2" @click="itemDetails(slotProps.data)" />
                                </template>
                            </Column>
                        </DataTable>
                    </div>
                </Panel>
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
