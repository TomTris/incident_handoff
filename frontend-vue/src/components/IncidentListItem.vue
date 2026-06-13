<script setup lang="ts">
import type { Incident } from '@/types';
import SeverityBadge from './SeverityBadge.vue';
import StatusBadge from './StatusBadge.vue';
import { computed, ref } from 'vue';

const props = defineProps<{inc : Incident}>()
function formatAge(createdIso: string, nowMs: number): string {
  const secs = Math.floor((nowMs - new Date(createdIso).getTime()) / 1000)
  if (secs < 60) return `${secs}s ago`
  const mins = Math.floor(secs / 60)
  if (mins < 60) return `${mins}m ago`
  const hrs = Math.floor(mins / 60)
  if (hrs < 24) return `${hrs}h ago`
  return `${Math.floor(hrs / 24)}d ago`
}

const now = ref(Date.now())
const created_at = computed(()=> formatAge(props.inc.created_at, now.value))
</script>

<template>
    <div class="incident-card">
        <RouterLink :to="{name: 'incident-detail', params: { id: props.inc.id } }" class="incident-link">
            <div class="incident-left">
                <span class="incident-id mono">{{ inc.id }}</span>
                <SeverityBadge :severity="inc.severity"></SeverityBadge>
            </div>
    
            <div class="incident-main">
                <h2 class="incident-title">{{ inc.title }}</h2>
                <div class="incident-meta">
                    <span class="meta-item">
                        <span class="meta-key">service</span>
                        <span class="mono">{{ inc.service }}</span>
                    </span>
                    <span class="meta-item">
                        <span class="meta-key">on-call</span>
                        <span class="mono">{{ inc.on_call || "-" }}</span>
                    </span>
                    <span class="meta-item">
                        <span class="meta-key">opened by</span>
                        <span class="mono">{{ inc.service }}</span>
                    </span>
                </div>
            </div>
    
            <div class="incident-right">
                <StatusBadge :status="inc.status"></StatusBadge>
                <span class="incident-created mono dim">{{ created_at }}</span>
            </div>
        </RouterLink>
    </div>
</template>

<style scoped>

.incident-card {
    background-color: var(--color-panel);
    border: 1px solid var(--color-border);
    border-left: 3px solid var(--color-border-strong);
    border-radius: 8px;
    transition: border-color 0.25s;
}

.incident-card:hover {
    border-color: var(--color-border-strong);
    border-left-color: var(--color-accent);
}

.incident-link {
  align-items: center;
  color: var(--color-text);
  display: flex;
  gap: 20px;
  padding: 16px 18px;
}

.incident-left {
    align-items: flex-start;
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 80px;
}

.incident-id {
    color: var(--color-text-dim);
    font-size: 13px;
}

.incident-main {
    flex:1;
}

.incident-title {
    color: var(--color-text-bright);
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 8px;
}

.incident-meta {
    display: flex;
    gap: 50px;
}

.meta-item {
    display: flex;
    flex-direction: column;
    font-size: 13px;
}

.meta-key {
    color: var(--color-text-dim);
    font-size: 11px;
    letter-spacing: 1px;
    text-transform: uppercase;
}

.incident-right {
    align-items: flex-end;
    display: flex;
    flex-direction: column;
    gap: 8px;
    width: 110px;
}

.incident-created {
    font-size: 12px
}
</style>