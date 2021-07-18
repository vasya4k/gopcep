<template>
  <div class="layout-topbar">
    <Menubar :model="items">
      <template #end>
        <InputText placeholder="Search" type="text" />
      </template>
    </Menubar>
  </div>
</template>

<script lang="ts">
import { Vue } from "vue-class-component";

export default class Topbar extends Vue {
  data() {
    return {
      items: [
        {
          label: "Inventory",
          icon: "pi pi-fw pi-file",
          items: [
            {
              label: "Routers",
              icon: "pi pi-fw pi-plus",
              to: "/routers"
            },
            {
              label: "Services",
              icon: "pi pi-fw pi-angle-double-right",
              to: "/bgp"
            },
            {
              separator: true
            }
          ]
        },
        {
          label: "PCEP",
          icon: "pi pi-fw pi-pencil",
          items: [
            {
              label: "Sessions",
              icon: "pi pi-fw pi-trash",
              to: "/psessions"
            },
            {
              label: "Parameters",
              icon: "pi pi-fw pi-align-right"
            }
          ]
        },
        {
          label: "LSPs",
          icon: "pi pi-fw pi-user",
          items: [
            {
              label: "List",
              icon: "pi pi-fw pi-user-plus",
              to: "/lsps"
            }
          ]
        },
        {
          label: "Topology",
          icon: "pi pi-fw pi-calendar",
          items: [
            {
              label: "BGP-LS",
              icon: "pi pi-fw pi-angle-double-right",
              to: "/bgp"
            },
            {
              label: "IS-IS",
              icon: "pi pi-fw pi-calendar-times"
            }
          ]
        },
        {
          label: "Debug",
          icon: "pi pi-fw pi-calendar",
          items: [
            {
              label: "Tail",
              icon: "pi pi-fw pi-pencil"
            },
            {
              label: "Archieve",
              icon: "pi pi-fw pi-calendar-times",
              items: [
                {
                  label: "Remove",
                  icon: "pi pi-fw pi-calendar-minus"
                }
              ]
            }
          ]
        }
      ]
    };
  }
}
</script>
