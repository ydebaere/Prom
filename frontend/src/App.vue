<template>
  <router-view />
</template>

<script setup>
import { onMounted } from "vue";
import { azureLogin } from "src/services/api.js";

onMounted(async () => {
  azureLogin();

});
</script>

<style>
body {
  font-family: "JetBrains Mono", monospace;
}
</style>
