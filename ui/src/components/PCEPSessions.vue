<template>
  <div class="grid">
    <div class="col-12">
      <div>
        <Panel header="PCEP Sessions" class="p-mt-5">
          <div class="card">
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
                ></div>
              </template>

              <Column
                selectionMode="multiple"
                style="width: 3rem"
                :exportable="false"
              ></Column>
              <Column
                field="Name"
                header="Name"
                :sortable="true"
                style="min-width:12rem"
              ></Column>
              <Column
                field="State"
                header="State"
                :sortable="true"
                style="min-width:16rem"
              ></Column>
              <Column
                field="SRCap.MSD"
                header="Max Segment Depth"
                :sortable="true"
                style="min-width:8rem"
              >
              </Column>
              <Column
                field="SRCap.Length"
                header="SR Capapability"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
              <Column
                field="StatefulCap.Length"
                header="Stateful Capapability"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
              <Column
                field="MsgCount"
                header="MsgCount"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
              <Column
                field="SRPID"
                header="SRPID"
                :sortable="true"
                style="min-width:10rem"
              >
              </Column>
            </DataTable>
          </div>

          <Dialog
            v-model:visible="productDialog"
            :style="{ width: '450px' }"
            header="Product Details"
            :modal="true"
            class="p-fluid"
          >
            <img
              src="https://www.primefaces.org/wp-content/uploads/2020/05/placeholder.png"
              :alt="product.image"
              class="product-image"
              v-if="product.image"
            />
            <div class="p-field">
              <label for="name">Name</label>
              <InputText
                id="name"
                v-model.trim="product.name"
                required="true"
                autofocus
                :class="{ 'p-invalid': submitted && !product.name }"
              />
              <small class="p-error" v-if="submitted && !product.name"
                >Name is required.</small
              >
            </div>
            <div class="p-field">
              <label for="description">Description</label>
              <Textarea
                id="description"
                v-model="product.description"
                required="true"
                rows="3"
                cols="20"
              />
            </div>

            <div class="p-field">
              <label for="inventoryStatus" class="p-mb-3"
                >Inventory Status</label
              >
              <Dropdown
                id="inventoryStatus"
                v-model="product.inventoryStatus"
                :options="statuses"
                optionLabel="label"
                placeholder="Select a Status"
              >
                <template #value="slotProps">
                  <div v-if="slotProps.value && slotProps.value.value">
                    <span
                      :class="'product-badge status-' + slotProps.value.value"
                      >{{ slotProps.value.label }}</span
                    >
                  </div>
                  <div v-else-if="slotProps.value && !slotProps.value.value">
                    <span
                      :class="
                        'product-badge status-' + slotProps.value.toLowerCase()
                      "
                      >{{ slotProps.value }}</span
                    >
                  </div>
                  <span v-else>
                    {{ slotProps.placeholder }}
                  </span>
                </template>
              </Dropdown>
            </div>

            <div class="p-field">
              <label class="p-mb-3">Category</label>
              <div class="p-formgrid p-grid">
                <div class="p-field-radiobutton p-col-6">
                  <RadioButton
                    id="category1"
                    name="category"
                    value="Accessories"
                    v-model="product.category"
                  />
                  <label for="category1">Accessories</label>
                </div>
                <div class="p-field-radiobutton p-col-6">
                  <RadioButton
                    id="category2"
                    name="category"
                    value="Clothing"
                    v-model="product.category"
                  />
                  <label for="category2">Clothing</label>
                </div>
                <div class="p-field-radiobutton p-col-6">
                  <RadioButton
                    id="category3"
                    name="category"
                    value="Electronics"
                    v-model="product.category"
                  />
                  <label for="category3">Electronics</label>
                </div>
                <div class="p-field-radiobutton p-col-6">
                  <RadioButton
                    id="category4"
                    name="category"
                    value="Fitness"
                    v-model="product.category"
                  />
                  <label for="category4">Fitness</label>
                </div>
              </div>
            </div>

            <div class="p-formgrid p-grid">
              <div class="p-field p-col">
                <label for="price">Price</label>
                <InputNumber
                  id="price"
                  v-model="product.price"
                  mode="currency"
                  currency="USD"
                  locale="en-US"
                />
              </div>
              <div class="p-field p-col">
                <label for="quantity">Quantity</label>
                <InputNumber
                  id="quantity"
                  v-model="product.quantity"
                  integeronly
                />
              </div>
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
                @click="saveProduct"
              />
            </template>
          </Dialog>

          <Dialog
            v-model:visible="deleteProductDialog"
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
                >Are you sure you want to delete <b>{{ product.name }}</b
                >?</span
              >
            </div>
            <template #footer>
              <Button
                label="No"
                icon="pi pi-times"
                class="p-button-text"
                @click="deleteProductDialog = false"
              />
              <Button
                label="Yes"
                icon="pi pi-check"
                class="p-button-text"
                @click="deleteProduct"
              />
            </template>
          </Dialog>

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
    HTTP.get("pcepsessions")
      .then(response => {
        for (const [key, value] of Object.entries(response.data)) {
          let newSession = {};
          newSession = value;
          newSession.Name = key;
          this.products.push(newSession);
        }
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
