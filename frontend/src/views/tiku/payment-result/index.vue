<template>
  <div class="payment-result tiku-page">
    <el-card shadow="never" class="result-card">
      <template #header><div class="card-title"><span>支付结果</span><el-button link type="primary" @click="backToPackages">返回套餐中心</el-button></div></template>
      <el-skeleton v-if="loading" :rows="5" animated />
      <el-result v-else-if="errorText" icon="error" title="订单查询失败" :sub-title="errorText"><template #extra><el-button type="primary" @click="loadOrder">重新查询</el-button></template></el-result>
      <template v-else-if="order">
        <el-result :icon="resultIcon" :title="resultTitle" :sub-title="resultDescription">
          <template #extra><el-button type="primary" :loading="loading" @click="loadOrder">刷新状态</el-button></template>
        </el-result>
        <el-alert v-if="isPending" title="支付平台的异步通知可能需要数秒到账。页面会自动刷新，请勿重复创建订单。" type="info" :closable="false" show-icon />
        <el-descriptions :column="1" border class="order-details">
          <el-descriptions-item label="订单号">{{ order.orderNo }}</el-descriptions-item>
          <el-descriptions-item label="套餐">{{ order.packageName }}</el-descriptions-item>
          <el-descriptions-item label="实付金额">¥{{ (order.payableCents / 100).toFixed(2) }}</el-descriptions-item>
          <el-descriptions-item label="订单状态"><el-tag :type="statusTagType">{{ statusText }}</el-tag></el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ order.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="支付时间">{{ order.paidAt || "-" }}</el-descriptions-item>
        </el-descriptions>
      </template>
      <el-empty v-else description="未提供订单号" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import { getMyOrder, type OrderItem } from "@/api/tiku";

const route = useRoute();
const router = useRouter();
const order = ref<OrderItem>();
const loading = ref(false);
const errorText = ref("");
let refreshTimer: ReturnType<typeof setTimeout> | undefined;
let refreshAttempts = 0;

const orderNo = computed(() => typeof route.query.order_no === "string" ? route.query.order_no.trim() : "");
const isPending = computed(() => order.value?.status === "pending");
const statusText = computed(() => ({ pending: "待支付", paid: "支付成功", closed: "订单已关闭", partial_refunded: "部分退款", refunded: "已退款" }[order.value?.status ?? ""] ?? order.value?.status ?? "未知"));
const statusTagType = computed(() => ({ pending: "warning", paid: "success", closed: "info", partial_refunded: "warning", refunded: "info" }[order.value?.status ?? ""] ?? "info"));
const resultIcon = computed(() => order.value?.status === "paid" ? "success" : order.value?.status === "pending" ? "info" : "warning");
const resultTitle = computed(() => order.value?.status === "paid" ? "支付成功" : order.value?.status === "pending" ? "等待支付确认" : statusText.value);
const resultDescription = computed(() => order.value?.status === "paid" ? "套餐权益已发放至账户，可返回套餐中心查看。" : order.value?.status === "pending" ? "请完成支付，系统会在收到支付通知后自动更新。" : "订单当前不再处于待支付状态。" );

const scheduleRefresh = () => {
  if (!isPending.value || refreshAttempts >= 10) return;
  refreshTimer = setTimeout(() => { refreshAttempts += 1; loadOrder(); }, 3000);
};
const loadOrder = async () => {
  if (!orderNo.value) return;
  if (refreshTimer) clearTimeout(refreshTimer);
  loading.value = true;
  errorText.value = "";
  try {
    order.value = (await getMyOrder(orderNo.value)).data;
    scheduleRefresh();
  } catch {
    errorText.value = "订单不存在，或不属于当前登录账户。";
  } finally {
    loading.value = false;
  }
};
const backToPackages = () => router.push("/tiku/packages");
onMounted(loadOrder);
onUnmounted(() => { if (refreshTimer) clearTimeout(refreshTimer); });
</script>

<style scoped lang="scss">
.payment-result { display: flex; justify-content: center; padding: 32px 16px; }
.result-card { width: min(720px, 100%); border-radius: 10px; }
.card-title { display: flex; align-items: center; justify-content: space-between; font-weight: 700; }
.order-details { margin-top: 20px; }
</style>
