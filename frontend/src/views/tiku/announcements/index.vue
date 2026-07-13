<template>
  <div class="tiku-page">
    <el-card shadow="never" class="tiku-card">
      <template #header><div class="card-title"><span class="title">系统公告</span><el-button link type="primary" :loading="loading" @click="load">刷新</el-button></div></template>
      <div v-if="items.length" class="announcement-list">
        <article v-for="item in items" :key="item.id" class="announcement-item">
          <div class="announcement-head"><h3>{{ item.title }}</h3><div><el-tag v-if="item.isPinned" type="warning" size="small">置顶</el-tag><span>{{ item.publishedAt || item.createdAt }}</span></div></div>
          <div class="content">{{ item.content }}</div>
        </article>
      </div>
      <el-empty v-else description="暂无公告" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { listAnnouncements, type AnnouncementItem } from "@/api/tiku";
const items = ref<AnnouncementItem[]>([]);
const loading = ref(false);
const load = async () => { loading.value = true; try { items.value = (await listAnnouncements()).data ?? []; } finally { loading.value = false; } };
onMounted(load);
</script>

<style scoped lang="scss">
.tiku-page { padding: 16px; }
.tiku-card { border-radius: 10px; }
.card-title, .announcement-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.title, h3 { color: var(--el-text-color-primary); font-weight: 700; }
.announcement-item { padding: 16px 0; border-bottom: 1px solid var(--el-border-color-lighter); }
.announcement-item:first-child { padding-top: 0; }
.announcement-item:last-child { border-bottom: 0; }
h3 { margin: 0; font-size: 17px; }
.announcement-head div { display: flex; align-items: center; gap: 8px; color: var(--el-text-color-secondary); font-size: 12px; }
.content { margin-top: 10px; color: var(--el-text-color-regular); white-space: pre-wrap; line-height: 1.8; }
</style>
