<script setup lang="ts">
import type { IncidentStatus } from '@/types';

const props = defineProps<{
  status: IncidentStatus;
}>()

</script>

<template>
  <span class="badge" :class="'status-' + props.status">{{ props.status }}</span>
</template>
