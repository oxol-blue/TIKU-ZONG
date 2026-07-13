<template>
  <div class="tiku-page">
    <el-card shadow="never" class="tiku-card">
      <template #header>
        <div class="card-title">
          <div><div class="title">{{ isAiOnly ? "AI 搜索记录" : "搜索记录" }}</div><div class="subtitle">{{ isAiOnly ? "仅展示 AI 兜底成功返回的历史答案。" : "查看当前账号成功返回的本地、第三方和 AI 搜题结果。" }}</div></div>
          <el-button :loading="loading" @click="load">刷新</el-button>
        </div>
      </template>
      <el-alert title="搜索记录仅对当前账号可见，包含题目与答案文本；调用日志仍用于查看接口状态与错误码。" type="info" :closable="false" show-icon />
      <el-table :data="items" stripe v-loading="loading" class="table">
        <el-table-column prop="question" label="题目" min-width="310" show-overflow-tooltip />
        <el-table-column prop="answer" label="答案（文字）" min-width="240" show-overflow-tooltip />
        <el-table-column prop="type" label="题型" width="120" />
        <el-table-column label="来源" width="100"><template #default="scope"><el-tag :type="scope.row.isAi ? 'warning' : 'success'">{{ scope.row.isAi ? 'AI' : '题库' }}</el-tag></template></el-table-column>
        <el-table-column prop="source" label="服务/题库来源" min-width="160" show-overflow-tooltip />
        <el-table-column label="耗时" width="105"><template #default="scope">{{ (scope.row.elapsedMicros / 1000).toFixed(2) }} ms</template></el-table-column>
        <el-table-column prop="createdAt" label="时间" min-width="180" />
        <el-table-column label="操作" width="80" fixed="right"><template #default="scope"><el-button link type="primary" @click="showDetail(scope.row)">详情</el-button></template></el-table-column>
      </el-table>
      <el-empty v-if="!loading && !items.length" description="暂无搜索记录" />
      <div class="pagination"><el-pagination v-model:current-page="page" v-model:page-size="pageSize" layout="total, sizes, prev, pager, next" :total="total" @current-change="load" @size-change="resetAndLoad" /></div>
    </el-card>
    <el-dialog v-model="detailDialog" title="搜索记录详情" width="720px">
      <template v-if="detail">
        <el-descriptions :column="2" border><el-descriptions-item label="题型">{{ detail.type || "-" }}</el-descriptions-item><el-descriptions-item label="来源">{{ detail.isAi ? "AI" : "题库" }}</el-descriptions-item><el-descriptions-item label="服务/题库" :span="2">{{ detail.source || "-" }}</el-descriptions-item><el-descriptions-item label="请求 ID" :span="2">{{ detail.requestId }}</el-descriptions-item><el-descriptions-item label="题目" :span="2"><div class="detail-text">{{ detail.question }}</div></el-descriptions-item><el-descriptions-item label="答案（文字）" :span="2"><div class="detail-text">{{ detail.answer }}</div></el-descriptions-item></el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useRoute } from "vue-router";
import { listMySearchHistory, type SearchHistoryItem } from "@/api/tiku";

const route = useRoute();
const isAiOnly = computed(() => route.path === "/tiku/ai-search-history");
const items = ref<SearchHistoryItem[]>([]);
const detail = ref<SearchHistoryItem>();
const detailDialog = ref(false);
const loading = ref(false);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);

const load = async () => {
  loading.value = true;
  try {
    const response = await listMySearchHistory({ page: page.value, pageSize: pageSize.value, isAi: isAiOnly.value || undefined });
    items.value = response.data?.items ?? [];
    total.value = response.data?.total ?? 0;
  } finally {
    loading.value = false;
  }
};
const resetAndLoad = () => { page.value = 1; void load(); };
const showDetail = (item: SearchHistoryItem) => { detail.value = item; detailDialog.value = true; };
watch(isAiOnly, () => resetAndLoad());
onMounted(load);
</script>

<style scoped lang="scss">
.tiku-page { padding: 16px; }.tiku-card { border-radius: 10px; }.card-title { display: flex; align-items: center; justify-content: space-between; gap: 12px; }.title { font-size: 18px; font-weight: 700; }.subtitle { margin-top: 6px; color: var(--el-text-color-secondary); font-size: 13px; }.table { margin-top: 16px; }.pagination { display: flex; justify-content: flex-end; margin-top: 16px; }.detail-text { white-space: pre-wrap; line-height: 1.7; }
</style>
