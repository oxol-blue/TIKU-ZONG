<template>
  <div class="tiku-page">
    <el-card shadow="never" class="tiku-card">
      <template #header>
        <div class="card-title"><div><div class="title">答案反馈</div><div class="subtitle">查看你提交过的答案反馈记录。</div></div><el-button :loading="loading" @click="load">刷新</el-button></div>
      </template>
      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="requestId" label="请求 ID" min-width="220" show-overflow-tooltip />
        <el-table-column prop="questionHash" label="题目哈希" min-width="270" show-overflow-tooltip />
        <el-table-column label="反馈类型" width="130"><template #default="scope"><el-tag :type="tagType(scope.row.feedbackType)">{{ typeLabel(scope.row.feedbackType) }}</el-tag></template></el-table-column>
        <el-table-column prop="comment" label="备注" min-width="220" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="提交时间" min-width="190" />
      </el-table>
      <el-empty v-if="!loading && !items.length" description="暂无反馈记录" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { listMyFeedback, type FeedbackItem } from "@/api/tiku";

const items = ref<FeedbackItem[]>([]);
const loading = ref(false);
const labels: Record<string, string> = { correct: "正确", incorrect: "错误", mismatch: "题目不匹配", parse_error: "解析错误", other: "其他" };
const typeLabel = (value: string) => labels[value] || value;
const tagType = (value: string) => value === "correct" ? "success" : value === "incorrect" || value === "parse_error" ? "danger" : "warning";
const load = async () => {
  loading.value = true;
  try {
    items.value = (await listMyFeedback()).data ?? [];
  } finally {
    loading.value = false;
  }
};
onMounted(load);
</script>

<style scoped lang="scss">
.tiku-page { padding: 16px; }.tiku-card { border-radius: 10px; }.card-title { display: flex; align-items: center; justify-content: space-between; gap: 12px; }.title { font-size: 18px; font-weight: 700; }.subtitle { margin-top: 6px; color: var(--el-text-color-secondary); font-size: 13px; }
</style>
