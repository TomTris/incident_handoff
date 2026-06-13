<script setup lang="ts">
import type { Severity } from '@/types';

const props = defineProps<{
  severity: Severity;
}>()

</script>

<template>
  <span class="badge" :class="'sev-' + props.severity">{{ props.severity }}</span>
</template>
