<template>
  <div class="admin-page">
    <el-card shadow="never" class="admin-card">
      <template #header><div class="card-title"><span>题库系统管理</span><div class="header-actions"><el-input v-model="adminTotp" type="password" maxlength="6" placeholder="TOTP（启用时填写）" @change="saveTotp" /><el-tag type="warning">管理员接口</el-tag></div></div></template>
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

        <el-tab-pane label="用户管理" name="users">
          <div class="toolbar">
            <el-input v-model="userSearch" clearable placeholder="按邮箱搜索" @keyup.enter="refreshUsers" />
            <el-select v-model="userStatus" clearable placeholder="全部状态" @change="refreshUsers">
              <el-option label="正常" :value="1" /><el-option label="禁用" :value="0" />
            </el-select>
            <el-button type="primary" @click="refreshUsers">查询</el-button>
          </div>
          <el-table :data="users" stripe class="table">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="email" label="邮箱" min-width="220" />
            <el-table-column label="角色" width="130">
              <template #default="scope"><el-select :model-value="scope.row.role" size="small" @change="changeRole(scope.row, $event)"><el-option label="用户" value="user" /><el-option label="管理员" value="admin" /></el-select></template>
            </el-table-column>
            <el-table-column label="状态" width="110">
              <template #default="scope"><el-tag :type="scope.row.status === 1 ? 'success' : 'danger'">{{ scope.row.status === 1 ? '正常' : '禁用' }}</el-tag></template>
            </el-table-column>
            <el-table-column prop="apiKeyPrefix" label="API Key" width="150" />
            <el-table-column prop="lastLoginAt" label="最近登录" min-width="180" />
            <el-table-column prop="createdAt" label="注册时间" min-width="180" />
            <el-table-column label="操作" width="110" fixed="right">
              <template #default="scope"><el-button link :type="scope.row.status === 1 ? 'danger' : 'success'" @click="toggleStatus(scope.row)">{{ scope.row.status === 1 ? '禁用' : '启用' }}</el-button></template>
            </el-table-column>
          </el-table>
          <div class="pagination"><el-pagination v-model:current-page="userPage" v-model:page-size="userPageSize" layout="total, sizes, prev, pager, next" :total="userTotal" @current-change="refreshUsers" @size-change="refreshUsers" /></div>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { configurePaymentGateway, createAdminPackage, createAiModel, createAiProvider, createCoupon, createOcsSource, listAdminUsers, listAiModels, listOcsSources, updateAdminUserRole, updateAdminUserStatus, type AdminUserItem } from "@/api/tiku";

const activeTab = ref("ocs");
const adminTotp = ref(sessionStorage.getItem("koi-admin-totp") || "");
const saving = ref(false);
const ocsSources = ref<any[]>([]);
const models = ref<any[]>([]);
const ocsForm = reactive({ name: "", url: "", method: "GET", dataText: '{"q":"${title}"}', successPath: "code", successValue: "1", questionPath: "q", answerPath: "data" });
const providerForm = reactive({ name: "", baseUrl: "", apiKey: "" });
const modelForm = reactive({ providerId: 1, name: "", priority: 100, aiChargeCount: 1 });
const gatewayForm = reactive({ name: "易支付", baseUrl: "", merchantId: "", secretKey: "" });
const packageForm = reactive({ name: "", type: "count", totalCount: 100, aiCount: 0, priceCents: 0, limitCount: 0 });
const couponForm = reactive({ code: "", discountType: "percent", discountValue: 10, totalLimit: 0 });
const users = ref<AdminUserItem[]>([]);
const userSearch = ref("");
const userStatus = ref<number | undefined>();
const userPage = ref(1);
const userPageSize = ref(20);
const userTotal = ref(0);
const refresh = async () => { ocsSources.value = (await listOcsSources()).data ?? []; models.value = (await listAiModels()).data ?? []; };
const refreshUsers = async () => { const result = await listAdminUsers({ page: userPage.value, pageSize: userPageSize.value, search: userSearch.value || undefined, status: userStatus.value }); users.value = result.data?.items ?? []; userTotal.value = result.data?.total ?? 0; };
const saveTotp = () => sessionStorage.setItem("koi-admin-totp", adminTotp.value.trim());
const saveOcs = async () => { saving.value = true; try { await createOcsSource({ name: ocsForm.name, url: ocsForm.url, method: ocsForm.method, data: JSON.parse(ocsForm.dataText), successPath: ocsForm.successPath, successValue: ocsForm.successValue, questionPath: ocsForm.questionPath, answerPath: ocsForm.answerPath, enabled: true }); ElMessage.success("OCS 源已保存"); await refresh(); } finally { saving.value = false; } };
const saveProvider = async () => { const response = await createAiProvider(providerForm); modelForm.providerId = response.data.id; ElMessage.success(`服务商已创建，Provider ID：${response.data.id}`); await refresh(); };
const saveModel = async () => { await createAiModel(modelForm); ElMessage.success("AI 模型已创建"); await refresh(); };
const saveGateway = async () => { await configurePaymentGateway({ provider: "epay", enabled: true, ...gatewayForm }); ElMessage.success("支付网关已保存"); };
const savePackage = async () => { await createAdminPackage(packageForm); ElMessage.success("套餐已创建"); };
const saveCoupon = async () => { await createCoupon(couponForm); ElMessage.success("优惠券已创建"); };
const toggleStatus = async (user: AdminUserItem) => { await updateAdminUserStatus(user.id, user.status === 1 ? 0 : 1); ElMessage.success("用户状态已更新"); await refreshUsers(); };
const changeRole = async (user: AdminUserItem, role: "admin" | "user") => { try { await updateAdminUserRole(user.id, role); ElMessage.success("用户角色已更新"); } catch { await refreshUsers(); } };
onMounted(async () => { await refresh(); await refreshUsers(); });
</script>

<style scoped lang="scss">
.admin-page { padding: 16px; }.admin-card { border-radius: 10px; }.card-title, .header-actions { display: flex; align-items: center; justify-content: space-between; gap: 8px; font-weight: 700; }.header-actions :deep(.el-input) { width: 190px; }.form-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 0 16px; }.toolbar { display: flex; gap: 12px; max-width: 620px; }.toolbar .el-select { width: 150px; }.table { margin-top: 18px; }.pagination { display: flex; justify-content: flex-end; margin-top: 18px; } @media (max-width: 900px) { .form-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } } @media (max-width: 600px) { .form-grid { grid-template-columns: 1fr; }.toolbar { flex-wrap: wrap; } }
</style>
