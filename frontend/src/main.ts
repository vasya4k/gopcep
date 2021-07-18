import { createApp } from "vue";

import "primeflex/primeflex.css";

import App from "./App.vue";
import router from "./router";
import PrimeVue from "primevue/config";
import Button from "primevue/button";
import Menubar from "primevue/menubar";
import InputText from "primevue/inputtext";
import FileUpload from "primevue/fileupload";
import Toolbar from "primevue/toolbar";
import Toast from "primevue/toast";
import DataTable from "primevue/datatable";
import Column from "primevue/column";
import ColumnGroup from "primevue/columngroup";
import Textarea from "primevue/textarea";
import Dropdown from "primevue/dropdown";
import RadioButton from "primevue/radiobutton";
import InputNumber from "primevue/inputnumber";
import Dialog from "primevue/dialog";
import Panel from "primevue/panel";
import Checkbox from "primevue/checkbox";
import Divider from "primevue/divider";
import Fieldset from "primevue/fieldset";
import ToastService from "primevue/toastservice";

import "primevue/resources/themes/mdc-light-indigo/theme.css";
import "primevue/resources/primevue.min.css";
import "primeicons/primeicons.css";

const app = createApp(App);

app.use(router);
app.use(PrimeVue);
app.use(ToastService);
app.component("Button", Button);
app.component("Menubar", Menubar);
app.component("InputText", InputText);
app.component("FileUpload", FileUpload);
app.component("Toolbar", Toolbar);
app.component("DataTable", DataTable);
app.component("Column", Column);
app.component("ColumnGroup", ColumnGroup);
app.component("Textarea", Textarea);
app.component("Dropdown", Dropdown);
app.component("RadioButton", RadioButton);
app.component("InputNumber", InputNumber);
app.component("Dialog", Dialog);
app.component("Panel", Panel);
app.component("Checkbox", Checkbox);
app.component("Divider", Divider);
app.component("Fieldset", Fieldset);
app.component("Toast", Toast);

app.mount("#app");
