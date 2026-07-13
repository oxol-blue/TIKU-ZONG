<template>
  <div class="tiku-page">
    <el-row :gutter="16">
      <el-col :xs="24" :lg="10">
        <el-card shadow="never" class="tiku-card">
          <template #header><span class="title">账号信息</span></template>
          <el-descriptions v-if="user" :column="1" border>
            <el-descriptions-item label="用户 ID">{{ user.id }}</el-descriptions-item>
            <el-descriptions-item label="邮箱">{{ user.email }}</el-descriptions-item>
            <el-descriptions-item label="角色"><el-tag>{{ user.role === "admin" ? "管理员" : "普通用户" }}</el-tag></el-descriptions-item>
            <el-descriptions-item label="注册时间">{{ user.createdAt }}</el-descriptions-item>
          </el-descriptions>
          <el-skeleton v-else :rows="4" animated />
        </el-card>
      </el-col>
      <el-col :xs="24" :lg="14">
        <el-card shadow="never" class="tiku-card">
          <template #header><div class="card-title"><span class="title">使用概览</span><el-button link type="primary" @click="goPackages">购买套餐</el-button></div></template>
          <div class="summary-grid">
            <div><strong>{{ activePackages.length }}</strong><span>有效套餐</span></div>
            <div><strong>{{ normalCount }}</strong><span>普通剩余次数</span></div>
            <div><strong>{{ aiCount }}</strong><span>AI 剩余次数</span></div>
          </div>
          <el-alert title="套餐可叠加使用；搜索失败不会扣除次数，AI 兜底单独消耗 AI 次数。" type="info" :closable="false" show-icon />
        </el-card>
      </el-col>
    </el-row>

    <el-card shadow="never" class="tiku-card">
      <template #header><div class="card-title"><span class="title">账号安全</span><el-button :loading="loading" @click="load">刷新</el-button></div></template>
      <div class="security-row"><div><strong>API Key</strong><p>{{ keyView ? keyView.masked : "尚未创建" }}</p></div><el-button type="primary" @click="goAccount">管理 Key 与 OCS 配置</el-button></div>
      <el-divider />
      <el-alert title="API Key 仅在创建或重新生成时返回一次明文，请不要提交到公开代码仓库。" type="warning" :closable="false" show-icon />
      <el-divider />
      <el-form ref="passwordFormRef" :model="passwordForm" :rules="passwordRules" label-position="top" class="password-form">
        <el-form-item label="当前密码" prop="currentPassword"><el-input v-model="passwordForm.currentPassword" type="password" show-password autocomplete="current-password" /></el-form-item>
        <el-form-item label="新密码" prop="newPassword"><el-input v-model="passwordForm.newPassword" type="password" show-password autocomplete="new-password" /></el-form-item>
        <el-form-item label="确认新密码" prop="confirmPassword"><el-input v-model="passwordForm.confirmPassword" type="password" show-password autocomplete="new-password" /></el-form-item>
        <el-button type="primary" :loading="changingPassword" @click="submitPasswordChange">修改密码</el-button>
      </el-form>
    </el-card>

    <el-card shadow="never" class="tiku-card">
      <template #header><span class="title">有效套餐明细</span></template>
      <el-table v-if="activePackages.length" :data="activePackages" stripe>
        <el-table-column prop="packageName" label="套餐" min-width="180" />
        <el-table-column prop="packageType" label="类型" width="150" />
        <el-table-column label="普通剩余" width="120"><template #default="scope">{{ scope.row.remainingCount < 0 ? "不限" : scope.row.remainingCount }}</template></el-table-column>
        <el-table-column prop="remainingAiCount" label="AI 剩余" width="110" />
        <el-table-column label="到期时间" min-width="180"><template #default="scope">{{ scope.row.expiresAt || "长期有效" }}</template></el-table-column>
      </el-table>
      <el-empty v-else description="暂无有效套餐" />
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, type FormInstance, type FormRules } from "element-plus";
import { getApiKey, listMyPackages, type ApiKeyView, type PackageInstance } from "@/api/tiku";
import { changePassword, getMe, type AuthUser } from "@/api/auth";

const router = useRouter();
const user = ref<AuthUser>();
const keyView = ref<ApiKeyView>();
const packages = ref<PackageInstance[]>([]);
const loading = ref(false);
const changingPassword = ref(false);
const passwordFormRef = ref<FormInstance>();
const passwordForm = reactive({ currentPassword: "", newPassword: "", confirmPassword: "" });
const passwordRules: FormRules = {
  currentPassword: [{ required: true, message: "请输入当前密码", trigger: "blur" }],
  newPassword: [
    { required: true, message: "请输入新密码", trigger: "blur" },
    { min: 8, max: 72, message: "新密码长度需为 8-72 位", trigger: "blur" }
  ],
  confirmPassword: [
    { required: true, message: "请再次输入新密码", trigger: "blur" },
    { validator: (_rule, value, callback) => value === passwordForm.newPassword ? callback() : callback(new Error("两次输入的新密码不一致")), trigger: "blur" }
  ]
};
const activePackages = computed(() => packages.value.filter(item => item.status === 1));
const normalCount = computed(() => activePackages.value.reduce((total, item) => total + (item.remainingCount < 0 ? 0 : item.remainingCount), 0));
const aiCount = computed(() => activePackages.value.reduce((total, item) => total + item.remainingAiCount, 0));
const load = async () => {
  loading.value = true;
  try {
    const [me, key, mine] = await Promise.all([getMe(), getApiKey(), listMyPackages()]);
    user.value = me.data;
    keyView.value = key.data;
    packages.value = mine.data ?? [];
  } finally { loading.value = false; }
};
const goAccount = () => router.push("/tiku/account");
const goPackages = () => router.push("/tiku/packages");
const submitPasswordChange = async () => {
  if (!passwordFormRef.value || !(await passwordFormRef.value.validate().catch(() => false))) return;
  changingPassword.value = true;
  try {
    await changePassword({ currentPassword: passwordForm.currentPassword, newPassword: passwordForm.newPassword });
    passwordForm.currentPassword = "";
    passwordForm.newPassword = "";
    passwordForm.confirmPassword = "";
    passwordFormRef.value.resetFields();
    ElMessage.success("密码修改成功");
  } finally { changingPassword.value = false; }
};
onMounted(load);
</script>

<style scoped lang="scss">
.tiku-page { display: flex; flex-direction: column; gap: 16px; padding: 16px; }
.tiku-card { border-radius: 10px; }
.title { color: var(--el-text-color-primary); font-weight: 700; }
.card-title, .security-row { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.security-row p { margin: 8px 0 0; color: var(--el-text-color-secondary); }
.password-form { max-width: 520px; }
.summary-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12px; margin-bottom: 18px; }
.summary-grid div { padding: 16px; background: var(--el-fill-color-light); border-radius: 8px; }
.summary-grid strong, .summary-grid span { display: block; }
.summary-grid strong { color: var(--el-color-primary); font-size: 24px; }
.summary-grid span { margin-top: 4px; color: var(--el-text-color-secondary); font-size: 13px; }
@media (max-width: 600px) { .summary-grid { grid-template-columns: 1fr; } .security-row { align-items: flex-start; flex-direction: column; } }
</style>
