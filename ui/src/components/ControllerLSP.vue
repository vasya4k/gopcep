<script>
import { HTTP } from '../service/http';
import { mapActions } from 'vuex';

export default {
    data() {
        return {
            items: [],
            item: {},
            deleteItemDialog: false,
            submitted: false,
            statuses: [
                { label: 'INSTOCK', value: 'instock' },
                { label: 'LOWSTOCK', value: 'lowstock' },
                { label: 'OUTOFSTOCK', value: 'outofstock' }
            ]
        };
    },
    created() {},
    mounted() {
        HTTP.get('ctrlsps')
            .then((response) => {
                this.items = response.data;
            })
            .catch(function (error) {
                console.log(error);
            });
    },
    methods: {
        ...mapActions('lsp', {
            editItem(dispatch, lsp) {
                dispatch('saveLSP', lsp);
                this.$router.push({ name: 'AddLSP', params: { new: false } });
            }
        }),
        confirmDeleteItem(item) {
            console.log(JSON.stringify(item, null, 2));
            this.item = item;
            this.deleteItemDialog = true;
        },
        newLSP() {
            this.$router.push({ name: 'AddLSP', params: { new: true } });
        },

        deleteItem() {
            HTTP.delete('lsp/' + this.item.Name)
                .then((response) => {
                    this.items = this.items.filter((val) => val.Name !== this.item.Name);

                    this.deleteItemDialog = false;
                    console.log(JSON.stringify(response, null, 2));
                    this.item = {};
                    this.$toast.add({
                        severity: 'success',
                        summary: 'Successful',
                        detail: 'Product Deleted',
                        life: 3000
                    });
                })
                .catch((error) => {
                    this.$toast.add({
                        severity: 'danger',
                        summary: 'Failure',
                        detail: error,
                        life: 3000
                    });
                    console.log(error);
                });
        },
        exportCSV() {
            this.$refs.dt.exportCSV();
        }
    }
};
</script>

<template>
    <div class="p-grid">
        <div class="p-col-12">
            <div>
                <Panel header="Controller LSPs" class="p-mt-4">
                    <div class="card">
                        <Toolbar class="p-mb-4">
                            <template v-slot:start>
                                <Button label="New" icon="pi pi-plus" class="p-button-success p-mr-2" @click="newLSP()" />
                            </template>
                        </Toolbar>

                        <DataTable
                            ref="dt"
                            :value="items"
                            :paginator="true"
                            :rows="10"
                            paginatorTemplate="FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink CurrentPageReport RowsPerPageDropdown"
                            :rowsPerPageOptions="[5, 10, 25]"
                            currentPageReportTemplate="Showing {first} to {last} of {totalRecords} items"
                            responsiveLayout="scroll"
                        >
                            <Column field="Name" header="Name" :sortable="true" style="min-width: 4rem"></Column>
                            <Column field="Src" header="SRC" :sortable="true" style="min-width: 8rem"></Column>
                            <Column field="Dst" header="DST" :sortable="true" style="min-width: 8rem"></Column>
                            <Column field="Admin" header="Admin State" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column field="Delegate" header="Delegate" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column field="BW" header="BW" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column field="Sync" header="Sync" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column field="HoldPrio" header="HoldPrio" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column field="LocalProtect" header="Local Protect" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column field="SetupPrio" header="Setup Prio" :sortable="true" style="min-width: 4rem"> </Column>
                            <Column>
                                <template #body="slotProps">
                                    <Button icon="pi pi-pencil" class="p-button-rounded p-button-success mr-2" @click="editItem(slotProps.data)" />
                                    <Button icon="pi pi-trash" class="p-button-rounded p-button-warning" @click="confirmDeleteItem(slotProps.data)" />
                                </template>
                            </Column>
                        </DataTable>
                    </div>

                    <Dialog v-model:visible="deleteItemDialog" :style="{ width: '450px' }" header="Confirm" :modal="true">
                        <div class="flex align-items-center justify-content-center">
                            <i class="pi pi-exclamation-triangle mr-3" style="font-size: 2rem" />
                            <span v-if="item"
                                >Are you sure you want to delete <b>{{ item.Name }}</b
                                >?</span
                            >
                        </div>
                        <template #footer>
                            <Button label="No" icon="pi pi-times" class="p-button-text" @click="deleteItemDialog = false" />
                            <Button label="Yes" icon="pi pi-check" class="p-button-text" @click="deleteItem" />
                        </template>
                    </Dialog>
                </Panel>
            </div>
        </div>
    </div>
</template>
