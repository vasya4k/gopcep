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
          label: "Routers",
          icon: "pi pi-fw pi-file",
          items: [
            {
              label: "New",
              icon: "pi pi-fw pi-plus",
              items: [
                {
                  label: "Bookmark",
                  icon: "pi pi-fw pi-bookmark"
                }
              ]
            },
            {
              label: "List",
              icon: "pi pi-fw pi-trash",
              to: "/about"
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
              label: "Sesstions",
              icon: "pi pi-fw pi-trash",
              to: "/about"
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
              icon: "pi pi-fw pi-user-plus"
            }
          ]
        },
        {
          label: "Logs",
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
        },
        {
          label: "Map",
          icon: "pi pi-fw pi-power-off"
        }
      ]
    };
  }
}
</script>
