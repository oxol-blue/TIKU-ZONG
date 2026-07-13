<template>
  <div class="admin-page">
    <el-card shadow="never" class="admin-card">
      <template #header><div class="card-title"><span>题库系统管理</span><el-tag type="warning">管理员接口</el-tag></div></template>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="OCS 题库源" name="ocs">
          <el-form :model="ocsForm" label-position="top" class="form-grid">
            <el-form-item label="名称"><el-input v-model="ocsForm.name" /></el-form-item>
            <el-form-item label="请求 URL"><el-input v-model="ocsForm.url" /></el-form-item>
            <el-form-item label="方法"><el-select v-model="ocsForm.method"><el-option label="GET" value="GET" /><el-option label="POST" value="POST" /></el-select></el-form-item>
            <el-form-item label="请求参数 JSON"><el-input v-model="ocsForm.dataText" type="textarea" :rows="2" placeholder='{"q":"${title}"}' /></el-form-item>
            <el-form-item label="成功字段路径"><el-input v-model="ocsForm.successPath" placeholder="code" /></el-form-item>
            <el-form-item label="成功值"><el-input v-model="ocsForm.successValue" placeholder="1" /></el-form-item>
            <el-form-item label="题目字段路径"><el-input v-model="ocsForm.questionPath" placeholder="q" /></el-form-item>
            <el-form-item label="答案字段路径"><el-input v-model="ocsForm.answerPath" placeholder="data" /></el-form-item>
          </el-form>
          <el-button type="primary" :loading="saving" @click="saveOcs">保存题库源</el-button>
          <el-table :data="ocsSources" stripe class="table"><el-table-column prop="name" label="名称" /><el-table-column prop="url" label="URL" /><el-table-column prop="priority" label="优先级" /><el-table-column prop="enabled" label="启用" /></el-table>
        </el-tab-pane>

        <el-tab-pane label="AI 模型" name="ai">
          <el-form :model="providerForm" label-position="top" class="form-grid">
            <el-form-item label="服务商名称"><el-input v-model="providerForm.name" /></el-form-item>
            <el-form-item label="Base URL"><el-input v-model="providerForm.baseUrl" placeholder="https://api.deepseek.com" /></el-form-item>
            <el-form-item label="API Key"><el-input v-model="providerForm.apiKey" type="password" show-password /></el-form-item>
          </el-form>
          <el-button type="primary" @click="saveProvider">创建服务商</el-button>
          <el-divider />
          <el-form :model="modelForm" label-position="top" class="form-grid">
            <el-form-item label="Provider ID"><el-input-number v-model="modelForm.providerId" :min="1" /></el-form-item>
            <el-form-item label="模型名称"><el-input v-model="modelForm.name" placeholder="deepseek-chat" /></el-form-item>
            <el-form-item label="优先级"><el-input-number v-model="modelForm.priority" :min="1" /></el-form-item>
            <el-form-item label="AI 扣费次数"><el-input-number v-model="modelForm.aiChargeCount" :min="1" /></el-form-item>
          </el-form>
          <el-button type="primary" @click="saveModel">创建模型</el-button>
          <el-table :data="models" stripe class="table"><el-table-column prop="id" label="ID" width="80" /><el-table-column prop="providerName" label="服务商" /><el-table-column prop="name" label="模型" /><el-table-column prop="priority" label="优先级" /><el-table-column prop="keyConfigured" label="密钥已配置" /></el-table>
        </el-tab-pane>

        <el-tab-pane label="支付网关" name="payment">
          <el-form :model="gatewayForm" label-position="top" class="form-grid">
            <el-form-item label="名称"><el-input v-model="gatewayForm.name" /></el-form-item>
            <el-form-item label="易支付 Base URL"><el-input v-model="gatewayForm.baseUrl" /></el-form-item>
            <el-form-item label="商户 ID"><el-input v-model="gatewayForm.merchantId" /></el-form-item>
            <el-form-item label="密钥"><el-input v-model="gatewayForm.secretKey" type="password" show-password /></el-form-item>
          </el-form>
          <el-button type="primary" @click="saveGateway">保存易支付配置</el-button>
        </el-tab-pane>

        <el-tab-pane label="套餐与优惠券" name="billing">
          <el-form :model="packageForm" label-position="top" class="form-grid">
            <el-form-item label="套餐名称"><el-input v-model="packageForm.name" /></el-form-item>
            <el-form-item label="类型"><el-select v-model="packageForm.type"><el-option label="时间" value="time" /><el-option label="次数" value="count" /><el-option label="时间次数" value="time_count" /></el-select></el-form-item>
            <el-form-item label="普通次数"><el-input-number v-model="packageForm.totalCount" :min="0" /></el-form-item>
            <el-form-item label="AI 次数"><el-input-number v-model="packageForm.aiCount" :min="0" /></el-form-item>
            <el-form-item label="价格（分）"><el-input-number v-model="packageForm.priceCents" :min="0" /></el-form-item>
            <el-form-item label="限购次数"><el-input-number v-model="packageForm.limitCount" :min="0" /></el-form-item>
          </el-form>
          <el-button type="primary" @click="savePackage">创建套餐</el-button>
          <el-divider />
          <el-form :model="couponForm" label-position="top" class="form-grid">
            <el-form-item label="优惠券码"><el-input v-model="couponForm.code" /></el-form-item>
            <el-form-item label="折扣类型"><el-select v-model="couponForm.discountType"><el-option label="固定金额（分）" value="fixed" /><el-option label="百分比" value="percent" /></el-select></el-form-item>
            <el-form-item label="折扣值"><el-input-number v-model="couponForm.discountValue" :min="1" /></el-form-item>
            <el-form-item label="总数量"><el-input-number v-model="couponForm.totalLimit" :min="0" /></el-form-item>
          </el-form>
          <el-button type="primary" @click="saveCoupon">创建优惠券</el-button>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { configurePaymentGateway, createAdminPackage, createAiModel, createAiProvider, createCoupon, createOcsSource, listAiModels, listOcsSources } from "@/api/tiku";

const activeTab = ref("ocs");
const saving = ref(false);
const ocsSources = ref<any[]>([]);
const models = ref<any[]>([]);
const ocsForm = reactive({ name: "", url: "", method: "GET", dataText: '{"q":"${title}"}', successPath: "code", successValue: "1", questionPath: "q", answerPath: "data" });
const providerForm = reactive({ name: "", baseUrl: "", apiKey: "" });
const modelForm = reactive({ providerId: 1, name: "", priority: 100, aiChargeCount: 1 });
const gatewayForm = reactive({ name: "易支付", baseUrl: "", merchantId: "", secretKey: "" });
const packageForm = reactive({ name: "", type: "count", totalCount: 100, aiCount: 0, priceCents: 0, limitCount: 0 });
const couponForm = reactive({ code: "", discountType: "percent", discountValue: 10, totalLimit: 0 });
const refresh = async () => { ocsSources.value = (await listOcsSources()).data ?? []; models.value = (await listAiModels()).data ?? []; };
const saveOcs = async () => { saving.value = true; try { await createOcsSource({ name: ocsForm.name, url: ocsForm.url, method: ocsForm.method, data: JSON.parse(ocsForm.dataText), successPath: ocsForm.successPath, successValue: ocsForm.successValue, questionPath: ocsForm.questionPath, answerPath: ocsForm.answerPath, enabled: true }); ElMessage.success("OCS 源已保存"); await refresh(); } finally { saving.value = false; } };
const saveProvider = async () => { const response = await createAiProvider(providerForm); modelForm.providerId = response.data.id; ElMessage.success(`服务商已创建，Provider ID：${response.data.id}`); await refresh(); };
const saveModel = async () => { await createAiModel(modelForm); ElMessage.success("AI 模型已创建"); await refresh(); };
const saveGateway = async () => { await configurePaymentGateway({ provider: "epay", enabled: true, ...gatewayForm }); ElMessage.success("支付网关已保存"); };
const savePackage = async () => { await createAdminPackage(packageForm); ElMessage.success("套餐已创建"); };
const saveCoupon = async () => { await createCoupon(couponForm); ElMessage.success("优惠券已创建"); };
onMounted(refresh);
</script>

<style scoped lang="scss">
.admin-page { padding: 16px; }.admin-card { border-radius: 10px; }.card-title { display: flex; align-items: center; justify-content: space-between; font-weight: 700; }.form-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 0 16px; }.table { margin-top: 18px; } @media (max-width: 900px) { .form-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } } @media (max-width: 600px) { .form-grid { grid-template-columns: 1fr; } }
</style>
