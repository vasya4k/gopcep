<template>
  <div class="p-grid">
    <div class="p-col-12">
      <div>
        <Panel header="Controller LSPs" class="p-mt-4">
          <div class="card">
            <Toolbar class="p-mb-4">
              <template #left>
                <Button
                  label="New"
                  icon="pi pi-plus"
                  class="p-button-success p-mr-2"
                  @click="$router.push('addlsp')"
                />
              </template>
              <template #right>
                <Button
                  label="Export"
                  icon="pi pi-upload"
                  class="p-button-help"
                  @click="exportCSV($event)"
                />
              </template>
            </Toolbar>

            <DataTable
              ref="dt"
              :value="products"
              v-model:selection="selectedProducts"
              dataKey="id"
              :paginator="true"
              :rows="10"
              :filters="filters"
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
                    <InputText
                      v-model="filters['global'].value"
                      placeholder="Search..."
                    />
                  </span>
                </div>
              </template>

              <Column
                field="Name"
                header="Name"
                :sortable="true"
                style="min-width:4rem"
              ></Column>
              <Column
                field="Src"
                header="SRC"
                :sortable="true"
                style="min-width:8rem"
              ></Column>
              <Column
                field="Dst"
                header="DST"
                :sortable="true"
                style="min-width:8rem"
              ></Column>
              <Column
                field="Admin"
                header="Admin State"
                :sortable="true"
                style="min-width:4rem"
              >
              </Column>
              <Column
                field="Delegate"
                header="Delegate"
                :sortable="true"
                style="min-width:4rem"
              >
              </Column>
              <Column
                field="BW"
                header="BW"
                :sortable="true"
                style="min-width:4rem"
              >
              </Column>
              <Column
                field="Sync"
                header="Sync"
                :sortable="true"
                style="min-width:4rem"
              >
              </Column>
              <Column
                field="HoldPrio"
                header="HoldPrio"
                :sortable="true"
                style="min-width:4rem"
              >
              </Column>
              <Column
                field="LocalProtect"
                header="Local Protect"
                :sortable="true"
                style="min-width:4rem"
              >
              </Column>
              <Column
                field="SetupPrio"
                header="Setup Prio"
                :sortable="true"
                style="min-width:4rem"
              >
              </Column>
              <Column :exportable="false">
                <template>
                  <Button
                    icon="pi pi-pencil"
                    class="p-button-rounded p-button-success p-mr-2"
                    @click="$router.push('addlsp')"
                  />
                  <Button
                    icon="pi pi-trash"
                    class="p-button-rounded p-button-warning"
                    @click="$router.push('addlsp')"
                  />
                </template>
              </Column>
            </DataTable>
          </div>

          <Dialog
            v-model:visible="deleteProductsDialog"
            :style="{ width: '450px' }"
            header="Confirm"
            :modal="true"
          >
            <div class="confirmation-content">
              <i
                class="pi pi-exclamation-triangle p-mr-3"
                style="font-size: 2rem"
              />
              <span v-if="product"
                >Are you sure you want to delete the selected products?</span
              >
            </div>
            <template #footer>
              <Button
                label="No"
                icon="pi pi-times"
                class="p-button-text"
                @click="deleteProductsDialog = false"
              />
              <Button
                label="Yes"
                icon="pi pi-check"
                class="p-button-text"
                @click="deleteSelectedProducts"
              />
            </template>
          </Dialog>
        </Panel>
      </div>
    </div>
  </div>
</template>

<script>
import { FilterMatchMode } from "primevue/api";
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
      submitted: false,
      statuses: [
        { label: "INSTOCK", value: "instock" },
        { label: "LOWSTOCK", value: "lowstock" },
        { label: "OUTOFSTOCK", value: "outofstock" }
      ]
    };
  },
  productService: null,
  created() {
    this.initFilters();
  },
  mounted() {
    HTTP.get("ctrlsps")
      .then(response => {
        console.log(response);
        this.products = response.data;
      })
      .catch(function(error) {
        console.log(error);
      });
  },
  methods: {
    formatCurrency(value) {
      if (value)
        return value.toLocaleString("en-US", {
          style: "currency",
          currency: "USD"
        });
      return;
    },
    openNew() {
      this.product = {};
      this.submitted = false;
      this.productDialog = true;
    },
    hideDialog() {
      this.productDialog = false;
      this.submitted = false;
    },
    saveProduct() {
      this.submitted = true;

      if (this.product.name.trim()) {
        if (this.product.id) {
          this.product.inventoryStatus = this.product.inventoryStatus.value
            ? this.product.inventoryStatus.value
            : this.product.inventoryStatus;
          this.products[this.findIndexById(this.product.id)] = this.product;
          this.$toast.add({
            severity: "success",
            summary: "Successful",
            detail: "Product Updated",
            life: 3000
          });
        } else {
          this.product.id = this.createId();
          this.product.code = this.createId();
          this.product.image = "product-placeholder.svg";
          this.product.inventoryStatus = this.product.inventoryStatus
            ? this.product.inventoryStatus.value
            : "INSTOCK";
          this.products.push(this.product);
          this.$toast.add({
            severity: "success",
            summary: "Successful",
            detail: "Product Created",
            life: 3000
          });
        }

        this.productDialog = false;
        this.product = {};
      }
    },
    editProduct(product) {
      this.product = { ...product };
      this.productDialog = true;
    },
    confirmDeleteProduct(product) {
      this.product = product;
      this.deleteProductDialog = true;
    },
    deleteProduct() {
      this.products = this.products.filter(val => val.id !== this.product.id);
      this.deleteProductDialog = false;
      this.product = {};
      this.$toast.add({
        severity: "success",
        summary: "Successful",
        detail: "Product Deleted",
        life: 3000
      });
    },
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
    createId() {
      let id = "";
      const chars =
        "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
      for (let i = 0; i < 5; i++) {
        id += chars.charAt(Math.floor(Math.random() * chars.length));
      }
      return id;
    },
    exportCSV() {
      this.$refs.dt.exportCSV();
    },
    confirmDeleteSelected() {
      this.deleteProductsDialog = true;
    },
    deleteSelectedProducts() {
      this.products = this.products.filter(
        val => !this.selectedProducts.includes(val)
      );
      this.deleteProductsDialog = false;
      this.selectedProducts = null;
      this.$toast.add({
        severity: "success",
        summary: "Successful",
        detail: "Products Deleted",
        life: 3000
      });
    },
    initFilters() {
      this.filters = {
        global: { value: null, matchMode: FilterMatchMode.CONTAINS }
      };
    }
  }
};
</script>
