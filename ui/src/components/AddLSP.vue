<script>
import { HTTP } from '../service/http';
import { reactive } from 'vue';

export default {
    data() {
        return {
            messages: [],
            emptyLSP: {
                Delegate: true,
                Sync: false,
                Remove: false,
                Admin: true,
                Name: '',
                Src: '',
                Dst: '',
                SetupPrio: 7,
                HoldPrio: 7,
                LocalProtect: false,
                BW: 0,
                EROList: [
                    {
                        LooseHop: false,
                        MBit: true,
                        NT: 3,
                        IPv4NodeID: '',
                        SID: 0,
                        NoSID: false,
                        IPv4Adjacency: ['0', '0']
                    }
                ]
            }
        };
    },
    created() {},
    computed: {
        lsp() {
            console.log(this.$route.params);
            if (this.$route.params.new == 'false') {
                return reactive(this.$store.getters['lsp/lspToAdd']);
            }
            return reactive(this.emptyLSP);
        }
    },
    mounted() {},
    methods: {
        add() {
            this.lsp.EROList.push({
                LooseHop: false,
                MBit: true,
                NT: 1,
                IPv4NodeID: '',
                SID: 0,
                NoSID: false
            });
        },
        remove(index) {
            console.log(index);
            this.lsp.EROList.splice(index + 1, 1);
        },
        save() {
            this.messages = [];
            HTTP.post('lsp', this.lsp)
                .then((response) => {
                    console.log(response.data);
                    this.$router.push('ctrlsps');
                })
                .catch((error) => {
                    if (error.response) {
                        this.messages = [{ severity: 'error', content: error.response.data }];
                        console.log(error.response.data);
                        console.log(error.response.status);
                        console.log('Error', error.message);
                        return;
                    }
                    this.messages = [{ severity: 'error', content: error.message }];
                });
        }
    }
};
</script>
<template>
    <div class="grid p-fluid">
        <div class="card">
            <h4>SR LSP</h4>
            <div class="p-fluid grid">
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <InputText type="text" placeholder="Name" v-model="lsp.Name"></InputText>
                </div>
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <InputText type="text" placeholder="Source" v-model="lsp.Src"></InputText>
                </div>
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <InputText type="text" placeholder="Destination" v-model="lsp.Dst" />
                </div>
            </div>

            <div class="p-fluid grid">
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <span class="p-input-icon-left">
                        <h6>Setup Priority</h6>
                        <InputNumber v-model="lsp.SetupPrio" showButtons mode="decimal"></InputNumber>
                    </span>
                </div>
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <span class="p-input-icon-right">
                        <h6>Hold Priority</h6>
                        <InputNumber v-model="lsp.HoldPrio" showButtons mode="decimal"></InputNumber>
                    </span>
                </div>
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <span class="p-input-icon-left p-input-icon-right">
                        <h6>BW</h6>
                        <InputNumber v-model="lsp.BW" showButtons mode="decimal"></InputNumber>
                    </span>
                </div>
            </div>
            <h5>First Hop ERO</h5>
            <hr style="width: 100%" />
            <div class="p-fluid grid">
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <h6>Src Interface Addr</h6>
                    <InputText type="text" placeholder="Src Interface Addr" v-model="lsp.EROList[0].IPv4Adjacency[0]" />
                </div>
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <h6>Dst Interface Addr</h6>
                    <InputText type="text" placeholder="Dst Interface Addr" v-model="lsp.EROList[0].IPv4Adjacency[1]" />
                </div>
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <h6>SID</h6>
                    <InputNumber v-model="lsp.EROList[0].SID" showButtons mode="decimal"></InputNumber>
                </div>
            </div>
            <h5>Additional Hop ERO</h5>
            <hr style="width: 100%" />
            <div class="grid p-fluid" v-for="(ero, k) in lsp.EROList.slice(1)" :key="k">
                <div class="field col-12 md:col-4">
                    <InputText type="text" placeholder="IPv4NodeID" v-model="ero.IPv4NodeID" />
                </div>
                <div class="field col-12 md:col-4">
                    <InputNumber type="text" showButtons placeholder="SID" v-model="ero.SID" />
                </div>
                <div class="field col-12 md:col-4">
                    <Button label="Remove" class="p-button-raised p-button-warning mr-2 mb-2" @click="remove(k)" v-show="k || (!k && lsp.EROList.length >= 1)" />
                </div>
            </div>

            <div class="grid p-fluid mt-2">
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <Button label="Add Hop" class="p-button-raised p-button-secondary mr-2 mb-2" @click="add" />
                </div>
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <Button label="Cancel" class="p-button-raised p-button-warning mr-2 mb-2" @click="$router.push({ name: 'CtrLSPs' })" />
                </div>
                <div class="col-12 mb-2 lg:col-4 lg:mb-0">
                    <Button label="Save" @click="save" class="p-button-raised mr-2 mb-2" />
                </div>
                <transition-group name="p-messages" tag="div">
                    <Message v-for="msg of messages" :severity="msg.severity" :key="msg.content">{{ msg.content }}</Message>
                </transition-group>
            </div>
        </div>
    </div>
</template>
