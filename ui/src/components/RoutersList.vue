<template>
  <div class="p-grid">
    <div class="p-col-12">
      <div>
        <Panel header="Routers" class="p-mt-4">
          <div class="card">
            <Toast />
            <Toolbar class="p-mb-4">
              <template #left>
                <Button
                  label="New"
                  icon="pi pi-plus"
                  class="p-button-success p-mr-2"
                  @click="openNew"
                />
              </template>

              <template #right>
                <Button
                  label="Export"
                  icon="pi pi-upload"
                  class="p-button-help"
                />
              </template>
            </Toolbar>

            <DataTable
              ref="dt"
              :value="routers"
              dataKey="id"
              :paginator="true"
              :rows="10"
              paginatorTemplate="FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink CurrentPageReport RowsPerPageDropdown"
              :rowsPerPageOptions="[5, 10, 25]"
              currentPageReportTemplate="Showing {first} to {last} of {totalRecords} products"
              responsiveLayout="scroll"
            >
              <Column
                field="Name"
                header="Name"
                :sortable="true"
                style="min-width:12rem"
              ></Column>
              <Column
                field="MgmtIP"
                header="MgmtIP"
                :sortable="true"
                style="min-width:12rem"
              ></Column>
              <Column
                field="ISOAddr"
                header="ISOAddr"
                :sortable="true"
                style="min-width:16rem"
              ></Column>
              <Column
                field="LoopbackIP"
                header="LoopbackIP"
                :sortable="true"
                style="min-width:8rem"
              >
              </Column>
              <Column
                field="PCEPSessionSrcIP"
                header="PCEPSessionSrcIP"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
              <Column
                field="IncludeInFullMesh"
                header="IncludeInFullMesh"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
              <Column
                field="BGPLSPeer"
                header="BGPLSPeer"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
              <Column :exportable="false">
                <template #body="slotProps">
                  <Button
                    icon="pi pi-pencil"
                    class="p-button-rounded p-button-success mr-2"
                    @click="editProduct(slotProps.data)"
                  />
                  <Button
                    icon="pi pi-trash"
                    class="p-button-rounded p-button-warning"
                    @click="confirmDeleteRouter(slotProps.data)"
                  />
                </template>
              </Column>
            </DataTable>
          </div>

          <Dialog
            v-model:visible="routerDialog"
            :style="{ width: '600px' }"
            header="Router Details"
            :modal="true"
            class="p-fluid"
          >
            <div class="field">
              <label for="name">Name</label>
              <InputText
                id="name"
                v-model.trim="router.Name"
                required="true"
                autofocus
                :class="{ 'p-invalid': submitted && !router.Name }"
              />
              <small class="p-error" v-if="submitted && !router.Name"
                >Name is required.</small
              >
            </div>
            <div class="formgrid grid">
              <div class="field col">
                <label for="price">Management IP</label>
                <InputText
                  id="MgmtIP"
                  v-model.trim="router.MgmtIP"
                  required="true"
                  autofocus
                  :class="{ 'p-invalid': submitted && !router.MgmtIP }"
                />
              </div>
              <div class="field col">
                <label for="ISO Address">ISO Address</label>
                <InputText
                  id="ISOAddr"
                  v-model.trim="router.ISOAddr"
                  required="true"
                  autofocus
                  :class="{ 'p-invalid': submitted && !router.ISOAddr }"
                />
              </div>
            </div>
            <div class="grid">
              <div class="field col">
                <label for="Loopback IP">LoopbackIP</label>
                <InputText
                  id="LoopbackIP"
                  v-model.trim="router.LoopbackIP"
                  required="true"
                  autofocus
                  :class="{ 'p-invalid': submitted && !router.LoopbackIP }"
                />
              </div>
              <div class="field col">
                <label for="PCEP Session Src IP">PCEP Session Src IP</label>
                <InputText
                  id="PCEPSessionSrcIP"
                  v-model.trim="router.PCEPSessionSrcIP"
                  required="true"
                  autofocus
                  :class="{
                    'p-invalid': submitted && !router.PCEPSessionSrcIP
                  }"
                />
              </div>
            </div>
            <div class="grid col">
              <Fieldset legend="Auto Full Mesh">
                <div class="field-checkbox col checkbox-icon">
                  <Checkbox
                    name="LSP Full Mesh"
                    v-model="router.IncludeInFullMesh"
                    :binary="true"
                  />
                  <label for="binary">LSP Full Mesh</label>
                </div>
              </Fieldset>
            </div>
            <div class="grid col">
              <Fieldset legend="BGP LS">
                <div class="field-checkbox col checkbox-icon">
                  <Checkbox
                    name="BGP LS Peer"
                    v-model="router.BGPLSPeer"
                    :binary="true"
                  />
                  <label for="binary">BGP LS Peer</label>
                </div>
                <div class="field-checkbox col checkbox-icon">
                  <Checkbox
                    name="eBGP Multihop"
                    v-model="router.BGPLSPeerCfg.EbgpMultihopEnabled"
                    :binary="true"
                  />
                  <label for="binary">eBGP Multihop</label>
                </div>
                <div class="field col">
                  <label for="NeighborAddress">Neighbour Address</label>
                  <InputText
                    id="NeighborAddress"
                    v-model.trim="router.BGPLSPeerCfg.NeighborAddress"
                  />
                </div>
                <div class="field col">
                  <label for="PeerAs">Peer AS</label>
                  <InputNumber
                    id="PeerAs"
                    v-model="router.BGPLSPeerCfg.PeerAs"
                    :useGrouping="false"
                  />
                </div>
                <div class="field col">
                  <label for="EBGPMultihopTtl">Multihop TTL</label>
                  <InputNumber
                    id="EBGPMultihopTtl"
                    v-model="router.BGPLSPeerCfg.EBGPMultihopTtl"
                  />
                </div>
              </Fieldset>
            </div>
            <template #footer>
              <Button
                label="Cancel"
                icon="pi pi-times"
                class="p-button-text"
                @click="hideDialog"
              />
              <Button
                label="Save"
                icon="pi pi-check"
                class="p-button-text"
                @click="saveRouter"
              />
            </template>
          </Dialog>

          <Dialog
            v-model:visible="deleteRouterDialog"
            :style="{ width: '450px' }"
            header="Confirm"
            :modal="true"
          >
            <div class="flex align-items-center justify-content-center">
              <i
                class="pi pi-exclamation-triangle mr-3"
                style="font-size: 2rem"
              />
              <span v-if="router"
                >Are you sure you want to delete <b>{{ router.Name }}</b
                >?</span
              >
            </div>
            <template #footer>
              <Button
                label="No"
                icon="pi pi-times"
                class="p-button-text"
                @click="deleteRouterDialog = false"
              />
              <Button
                label="Yes"
                icon="pi pi-check"
                class="p-button-text"
                @click="deleteRouter"
              />
            </template>
          </Dialog>
        </Panel>
      </div>
    </div>
  </div>
</template>

<script>
import { HTTP } from "../service/http";

export default {
  data() {
    return {
      routers: null,
      routerDialog: false,
      deleteRouterDialog: false,
      deleteProductsDialog: false,
      router: { BGPLSPeerCfg: {} },
      selectedProducts: null,
      filters: {},
      submitted: false
    };
  },
  mounted() {
    HTTP.get("routers")
      .then(response => {
        this.routers = response.data;
      })
      .catch(function(error) {
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
          console.log("Error", error.message);
        }
        console.log(error.config);
      });
  },
  methods: {
    openNew() {
      this.router = { BGPLSPeerCfg: {} };
      this.submitted = false;
      this.routerDialog = true;
    },
    hideDialog() {
      this.routerDialog = false;
      this.submitted = false;
    },
    saveRouter() {
      this.submitted = true;

      HTTP.post("router", this.router)
        .then(response => {
          console.log(response.data);
          if (this.router.ID) {
            this.routers[this.findIndexById(this.router.ID)] = this.router;
          } else {
            this.routers.push(response.data);
          }
          this.routerDialog = false;
          this.$toast.add({
            severity: "success",
            summary: "Successful",
            detail: "Updated",
            life: 3000
          });
          this.router = { BGPLSPeerCfg: {} };
        })
        .catch(function(error) {
          console.log(error);
        });
    },
    editProduct(router) {
      this.router = { ...router };
      this.routerDialog = true;
    },
    confirmDeleteRouter(router) {
      this.router = router;
      this.deleteRouterDialog = true;
    },
    deleteRouter() {
      HTTP.delete("router/" + this.router.ID)
        .then(response => {
          this.deleteRouterDialog = false;
          this.routers = this.routers.filter(val => val.ID !== this.router.ID);
          console.log(response);

          this.router = { BGPLSPeerCfg: {} };
          this.$toast.add({
            severity: "success",
            summary: "Successful",
            detail: "Product Deleted",
            life: 3000
          });
        })
        .catch(error => {
          console.log(this.router.ID);
          this.$toast.add({
            severity: "danger",
            summary: "Failure",
            detail: error,
            life: 3000
          });
        });
    },
    findIndexById(id) {
      let index = -1;
      for (let i = 0; i < this.routers.length; i++) {
        if (this.routers[i].ID === id) {
          index = i;
          break;
        }
      }
      return index;
    },
    exportCSV() {
      this.$refs.dt.exportCSV();
    }
  }
};
</script>
