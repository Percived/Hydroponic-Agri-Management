<template>
  <router-view v-slot="{ Component, route }">
    <AppLayout v-if="route.meta.requiresAuth">
      <Transition name="page-fade" mode="out-in">
        <component :is="Component" :key="route.path" />
      </Transition>
    </AppLayout>
    <Transition v-else name="page-fade" mode="out-in">
      <component :is="Component" :key="route.path" />
    </Transition>
  </router-view>
</template>

<script setup lang="ts">
import AppLayout from '@/components/layout/AppLayout.vue'
</script>
