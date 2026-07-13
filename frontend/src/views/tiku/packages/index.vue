<template>
  <div class="tiku-page">
    <div class="page-heading">
      <div><h2>套餐中心</h2><p>购买后可用于普通题库调用和 AI 兜底。</p></div>
      <div class="heading-actions"><el-input v-model="couponCode" clearable placeholder="优惠券码" /><el-button :loading="loading" @click="load">刷新</el-button></div>
    </div>
    <el-row :gutter="16">
      <el-col v-for="item in packages" :key="item.id" :xs="24" :sm="12" :lg="8" class="package-col">
        <el-card shadow="hover" class="package-card">
          <div class="package-head"><span class="package-name">{{ item.name }}</span><div class="tags"><el-tag v-if="item.isTrial" type="warning">试用</el-tag><el-tag v-if="item.isFree" type="success">免费</el-tag><el-tag>{{ typeText(item.type) }}</el-tag></div></div>
          <div class="price">¥{{ (item.priceCents / 100).toFixed(2) }}</div>
          <ul>
            <li>{{ item.type === 'time' ? '有效期内普通调用不限次数' : `普通调用 ${item.totalCount} 次` }}</li>
            <li v-if="item.durationSeconds">有效期 {{ durationText(item.durationSeconds) }}</li>
            <li>AI 额度 {{ item.aiCount }} 次</li>
          </ul>
          <el-button type="primary" class="buy-button" :loading="buying === item.id" @click="buy(item)">立即购买</el-button>
        </el-card>
      </el-col>
    </el-row>
    <el-card shadow="never" class="tiku-card">
      <template #header><span class="title">我的套餐</span></template>
      <el-table :data="instances" stripe>
        <el-table-column prop="packageName" label="套餐" />
        <el-table-column prop="packageType" label="类型" />
        <el-table-column label="普通剩余"><template #default="scope">{{ scope.row.remainingCount < 0 ? '不限' : scope.row.remainingCount }}</template></el-table-column>
        <el-table-column prop="remainingAiCount" label="AI 剩余" />
        <el-table-column prop="expiresAt" label="到期时间"><template #default="scope">{{ scope.row.expiresAt || '长期' }}</template></el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { createOrder, listMyPackages, listPackages, type PackageInstance, type PackageItem } from "@/api/tiku";

const packages = ref<PackageItem[]>([]);
const instances = ref<PackageInstance[]>([]);
const loading = ref(false);
const buying = ref<number>();
const couponCode = ref("");
const typeText = (type: string) => ({ time: "时间套餐", count: "次数套餐", time_count: "时间次数套餐" }[type] || type);
const durationText = (seconds: number) => seconds >= 86400 ? `${Math.floor(seconds / 86400)} 天` : `${Math.floor(seconds / 3600)} 小时`;
const load = async () => {
  loading.value = true;
  try {
    const [catalog, mine] = await Promise.all([listPackages(), listMyPackages()]);
    packages.value = catalog.data ?? [];
    instances.value = mine.data ?? [];
  } finally { loading.value = false; }
};
const buy = async (item: PackageItem) => {
  buying.value = item.id;
  try {
    const response = await createOrder({ packageId: item.id, provider: "epay", couponCode: couponCode.value.trim() || undefined });
    if (response.data.paymentUrl) window.open(response.data.paymentUrl, "_blank", "noopener,noreferrer");
    ElMessage.success("订单已创建，请在新窗口完成支付");
  } finally { buying.value = undefined; }
};
onMounted(load);
</script>

<style scoped lang="scss">
.tiku-page { display: flex; flex-direction: column; gap: 16px; padding: 16px; }
.page-heading { display: flex; align-items: center; justify-content: space-between; gap: 16px; }.heading-actions { display: flex; align-items: center; gap: 8px; }.tags { display: flex; gap: 4px; }
h2 { margin: 0; color: var(--el-text-color-primary); } p { margin: 6px 0 0; color: var(--el-text-color-secondary); }
.package-col { margin-bottom: 16px; }.package-card { height: 100%; border-radius: 10px; }.package-head { display: flex; justify-content: space-between; gap: 8px; }.package-name, .title { font-weight: 700; color: var(--el-text-color-primary); }.price { margin: 20px 0 8px; color: var(--el-color-primary); font-size: 30px; font-weight: 800; }.package-card ul { min-height: 72px; padding-left: 18px; color: var(--el-text-color-secondary); line-height: 1.9; }.buy-button { width: 100%; }.tiku-card { border-radius: 10px; }
</style>
