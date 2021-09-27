<template>
  <div class="grid p-fluid">
    <div class="card">
      <h4>SR LSP</h4>
      <div class="p-fluid grid">
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <InputText
            type="text"
            placeholder="Name"
            v-model="lsp.Name"
          ></InputText>
        </div>
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <InputText
            type="text"
            placeholder="Source"
            v-model="lsp.Src"
          ></InputText>
        </div>
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <InputText type="text" placeholder="Destination" v-model="lsp.Dst" />
        </div>
      </div>

      <div class="p-fluid grid">
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <span class="p-input-icon-left">
            <h6>Setup Priority</h6>
            <InputNumber
              v-model="lsp.SetupPrio"
              showButtons
              mode="decimal"
            ></InputNumber>
          </span>
        </div>
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <span class="p-input-icon-right">
            <h6>Hold Priority</h6>
            <InputNumber
              v-model="lsp.HoldPrio"
              showButtons
              mode="decimal"
            ></InputNumber>
          </span>
        </div>
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <span class="p-input-icon-left p-input-icon-right">
            <h6>BW</h6>
            <InputNumber
              v-model="lsp.BW"
              showButtons
              mode="decimal"
            ></InputNumber>
          </span>
        </div>
      </div>
      <h5>First Hop ERO</h5>
      <hr style="width:100%" />
      <div class="p-fluid grid">
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <InputText
            type="text"
            placeholder="Src Interface Addr"
            v-model="lsp.EROList[0].IPv4Adjacency[0]"
          />
        </div>
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <InputText
            type="text"
            placeholder="Dst Interface Addr"
            v-model="lsp.EROList[0].IPv4Adjacency[1]"
          />
        </div>
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <InputNumber
            v-model="lsp.EROList[0].SID"
            showButtons
            mode="decimal"
          ></InputNumber>
        </div>
      </div>
      <hr style="width:100%" />
      <div class="grid p-fluid" v-for="(ero, k) in EROList" :key="k">
        <div class="field col-12 md:col-4">
          <InputText
            type="text"
            placeholder="IPv4NodeID"
            v-model="ero.IPv4NodeID"
          />
        </div>
        <div class="field col-12 md:col-4">
          <InputNumber
            type="text"
            showButtons
            placeholder="SID"
            v-model="ero.SID"
          />
        </div>
        <div class="field col-12 md:col-4">
          <Button
            label="Remove"
            class="p-button-raised p-button-warning mr-2 mb-2"
            @click="remove(k)"
            v-show="k || (!k && EROList.length >= 1)"
          />
        </div>
      </div>

      <div class="grid p-fluid mt-2">
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <Button
            label="Add Hop"
            class="p-button-raised p-button-secondary mr-2 mb-2"
            @click="add"
          />
        </div>
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <Button
            label="Cancel"
            class="p-button-raised p-button-warning mr-2 mb-2"
            @click="$router.push('ctrlsps')"
          />
        </div>
        <div class="col-12 mb-2 lg:col-4 lg:mb-0">
          <Button
            label="Save"
            @click="save"
            class="p-button-raised mr-2 mb-2"
          />
        </div>
        <transition-group name="p-messages" tag="div">
          <Message
            v-for="msg of messages"
            :severity="msg.severity"
            :key="msg.content"
            >{{ msg.content }}</Message
          >
        </transition-group>
      </div>
    </div>
  </div>
</template>
<script>
// data() {
// 			return {
// 				message: [],
// 				username:null,
// 				email:null
// 			}
// 		},
// 		methods: {
// 			addSuccessMessage() {
// 				this.message = [{severity: 'success', content: 'Message Detail'}]
// 			},
// 			addInfoMessage() {
// 				this.message = [{severity: 'info', content: 'Message Detail'}]
// 			},
// 			addWarnMessage() {
// 				this.message = [{severity: 'warn', content: 'Message Detail'}]
// 			},
// 			addErrorMessage() {
// 				this.message = [{severity: 'error', content: 'Message Detail'}]
// 			},
// 			showSuccess() {
// 				this.$toast.add({severity:'success', summary: 'Success Message', detail:'Message Detail', life: 3000});

import { HTTP } from "../service/http";
export default {
  data() {
    return {
      messages: [],
      EROList: [],
      lsp: {
        Delegate: true,
        Sync: false,
        Remove: false,
        Admin: true,
        Name: "",
        Src: "",
        Dst: "",
        SetupPrio: 7,
        HoldPrio: 7,
        LocalProtect: false,
        BW: 0,
        EROList: [
          {
            LooseHop: false,
            MBit: true,
            NT: 3,
            IPv4NodeID: "",
            SID: 0,
            NoSID: false,
            IPv4Adjacency: ["", ""]
          }
        ]
      }
    };
  },
  created() {},
  mounted() {},
  methods: {
    add() {
      this.EROList.push({
        LooseHop: false,
        MBit: true,
        NT: 1,
        IPv4NodeID: "",
        SID: 0,
        NoSID: false
      });
      console.log(this.EROList);
    },
    remove(index) {
      this.EROList.splice(index, 1);
    },
    save() {
      this.messages = [];
      this.lsp.EROList = this.lsp.EROList.concat(this.EROList);

      HTTP.post("lsp", this.lsp)
        .then(response => {
          console.log(response.data);
          this.$router.push("ctrlsps");
        })
        .catch(error => {
          if (error.response) {
            this.messages = [
              { severity: "error", content: error.response.data }
            ];
            console.log(error.response.data);
            console.log(error.response.status);
            console.log("Error", error.message);
            return;
          }
          this.messages = [{ severity: "error", content: error.message }];
        });
    }
  }
};
</script>
