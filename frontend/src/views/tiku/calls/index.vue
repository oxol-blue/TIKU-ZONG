<template>
  <div class="tiku-page">
    <el-card shadow="never" class="tiku-card">
      <template #header>
        <div class="card-title"><div><div class="title">调用记录</div><div class="subtitle">查看当前账号最近的搜题 API 调用情况。</div></div><el-button :loading="loading" @click="load">刷新</el-button></div>
      </template>
      <el-table :data="items" stripe v-loading="loading">
        <el-table-column prop="requestId" label="请求 ID" min-width="210" show-overflow-tooltip />
        <el-table-column prop="endpoint" label="接口" min-width="150" />
        <el-table-column label="来源" width="90"><template #default="scope"><el-tag :type="scope.row.isAi ? 'warning' : 'success'">{{ scope.row.isAi ? 'AI' : '题库' }}</el-tag></template></el-table-column>
        <el-table-column label="结果" width="90"><template #default="scope"><el-tag :type="scope.row.success ? 'success' : 'danger'">{{ scope.row.success ? '成功' : '失败' }}</el-tag></template></el-table-column>
        <el-table-column label="耗时" width="110"><template #default="scope">{{ (scope.row.elapsedMicros / 1000).toFixed(2) }} ms</template></el-table-column>
        <el-table-column prop="httpStatus" label="HTTP" width="80" />
        <el-table-column prop="errorCode" label="错误码" width="150" />
        <el-table-column prop="createdAt" label="时间" min-width="190" />
      </el-table>
      <el-empty v-if="!loading && !items.length" description="暂无调用记录" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { listMyCalls, type AdminCallLog } from "@/api/tiku";

const items = ref<AdminCallLog[]>([]);
const loading = ref(false);
const load = async () => {
  loading.value = true;
  try {
    items.value = (await listMyCalls()).data ?? [];
  } finally {
    loading.value = false;
  }
};
onMounted(load);
</script>

<style scoped lang="scss">
.tiku-page { padding: 16px; }.tiku-card { border-radius: 10px; }.card-title { display: flex; align-items: center; justify-content: space-between; gap: 12px; }.title { font-size: 18px; font-weight: 700; }.subtitle { margin-top: 6px; color: var(--el-text-color-secondary); font-size: 13px; }
</style>
