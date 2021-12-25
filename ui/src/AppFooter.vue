<template>
  <div class="layout-footer">
    <span class="font-medium ml-2">Chetki Design Buro 2021</span>
  </div>
</template>

<script>
export default {
  name: "AppFooter",
  computed: {
    darkTheme() {
      return this.$appState.theme.startsWith("saga");
    }
  }
};
</script>

<style scoped></style>
