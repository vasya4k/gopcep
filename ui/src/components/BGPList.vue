<template>
  <div class="grid">
    <div class="col-12">
      <div>
        <Panel header="BGP-LS" class="mt-4">
          <div class="card">
            <Toolbar class="mb-4">
              <template #left>
                <Button
                  label="New"
                  icon="pi pi-plus"
                  class="p-button-success mr-2"
                />
                <Button
                  label="Delete"
                  icon="pi pi-trash"
                  class="p-button-danger"
                />
              </template>

              <template #right>
                <FileUpload
                  mode="basic"
                  accept="image/*"
                  :maxFileSize="1000000"
                  label="Import"
                  chooseLabel="Import"
                  class="mr-2 p-d-inline-block"
                />
                <Button
                  label="Export"
                  icon="pi pi-upload"
                  class="p-button-help"
                />
              </template>
            </Toolbar>
            <DataTable
              ref="dt"
              :value="products"
              dataKey="id"
              :paginator="true"
              :rows="10"
              paginatorTemplate="FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink CurrentPageReport RowsPerPageDropdown"
              :rowsPerPageOptions="[5, 10, 25]"
              currentPageReportTemplate="Showing {first} to {last} of {totalRecords} products"
              responsiveLayout="scroll"
            >
              <template #header>
                <div
                  class="table-header p-d-flex p-flex-column p-flex-md-row p-jc-md-between"
                >
                  <span class="p-input-icon-left">
                    <i class="pi pi-search" />
                    <InputText placeholder="Search..." />
                  </span>
                </div>
              </template>

              <Column
                selectionMode="multiple"
                style="width: 3rem"
                :exportable="false"
              ></Column>
              <Column
                field="state.neighbor_address"
                header="Neighbor"
                :sortable="true"
                style="min-width:12rem"
              ></Column>
              <Column
                field="state.peer_as"
                header="AS"
                :sortable="true"
                style="min-width:16rem"
              ></Column>
              <Column
                field="state.session_state"
                header="State"
                :sortable="true"
                style="min-width:8rem"
              >
              </Column>
              <Column
                field="state.peer_type"
                header="Type"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
              <Column
                field="state.router_id"
                header="Router ID"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
              <Column
                field="state.messages.received.total"
                header="Rcv Msg"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
              <Column
                field="state.messages.sent.total"
                header="Sent Msg"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
              <Column :exportable="false">
                <template>
                  <Button
                    icon="pi pi-pencil"
                    class="p-button-rounded p-button-success p-mr-2"
                  />
                  <Button
                    icon="pi pi-trash"
                    class="p-button-rounded p-button-warning"
                  />
                </template>
              </Column>
            </DataTable>
          </div>
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
  mounted() {
    HTTP.get("bgpneighbors")
      .then(response => {
        this.products = response.data;
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
  methods: {}
};
</script>
