<template>
  <div class="tiku-page">
    <el-card shadow="never" class="tiku-card">
      <template #header><div class="card-title"><span class="title">API Key</span><el-tag type="warning">每个用户一个</el-tag></div></template>
      <el-alert title="API Key 只在创建时显示完整值，请妥善保存。系统数据库只保存哈希。" type="warning" :closable="false" show-icon />
      <div v-if="keyView" class="key-box"><div><span class="label">当前 Key</span><code>{{ keyView.masked }}</code></div><span class="created">创建于 {{ keyView.createdAt }}</span></div>
      <el-empty v-else description="尚未创建 API Key" />
      <div class="actions"><el-button type="primary" :loading="creating" :disabled="!!keyView" @click="create">创建 API Key</el-button><el-button type="warning" :loading="creating" :disabled="!keyView" @click="rotate">重新生成 Key</el-button><el-button type="danger" plain :loading="revoking" :disabled="!keyView" @click="revoke">停用 Key</el-button><el-button :disabled="!keyView" @click="load">刷新状态</el-button></div>
      <el-alert v-if="plainKey" class="plain-key" title="请立即复制完整 Key，此值不会再次返回" type="success" :closable="false"><template #default><code>{{ plainKey }}</code><el-button link type="primary" @click="copyKey">复制</el-button></template></el-alert>
    </el-card>

    <el-card shadow="never" class="tiku-card">
      <template #header><span class="title">OCS 题库配置</span></template>
      <p class="hint">生成后可复制到 OCS AnswererWrapper 题库配置中使用。</p>
      <el-alert title="请输入本次要写入 OCS 配置的完整 API Key。该值只保留在当前浏览器内存中，不会保存到服务器。" type="info" :closable="false" show-icon />
      <el-input v-model="ocsKey" class="ocs-key-input" type="password" show-password autocomplete="off" placeholder="粘贴完整 API Key" />
      <el-button type="primary" :loading="loadingConfig" :disabled="!ocsKey.trim()" @click="generateConfig">生成 OCS 配置</el-button>
      <el-input v-if="ocsConfig" v-model="ocsConfig" class="config-input" type="textarea" :rows="12" readonly />
      <div class="actions"><el-button v-if="ocsConfig" @click="copyConfig">复制配置</el-button><el-button v-if="ocsConfig" type="success" @click="downloadConfig">下载配置</el-button></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from "vue";
import { ElMessage } from "element-plus";
import { createApiKey, getApiKey, revokeApiKey, rotateApiKey, type ApiKeyView } from "@/api/tiku";
import { getToken } from "@/utils/storage";

const keyView = ref<ApiKeyView>();
const plainKey = ref("");
const ocsKey = ref("");
const ocsConfig = ref("");
const creating = ref(false);
const revoking = ref(false);
const loadingConfig = ref(false);
const load = async () => { try { keyView.value = (await getApiKey()).data; } catch { keyView.value = undefined; } };
const create = async () => { creating.value = true; try { const response = await createApiKey(); plainKey.value = response.data.key; ocsKey.value = response.data.key; keyView.value = response.data.info; } finally { creating.value = false; } };
const rotate = async () => { if (!window.confirm("重新生成后旧 API Key 将立即失效，是否继续？")) return; creating.value = true; try { const response = await rotateApiKey(); plainKey.value = response.data.key; ocsKey.value = response.data.key; keyView.value = response.data.info; ocsConfig.value = ""; } finally { creating.value = false; } };
const revoke = async () => { if (!window.confirm("停用后 API 和 OCS 调用将立即失效，是否继续？")) return; revoking.value = true; try { await revokeApiKey(); keyView.value = undefined; plainKey.value = ""; ocsKey.value = ""; ocsConfig.value = ""; ElMessage.success("API Key 已停用"); } finally { revoking.value = false; } };
const copy = async (value: string) => { await navigator.clipboard.writeText(value); ElMessage.success("已复制"); };
const copyKey = () => copy(plainKey.value);
const copyConfig = () => copy(ocsConfig.value);
const downloadConfig = () => { const blob = new Blob([ocsConfig.value], { type: "application/json;charset=utf-8" }); const url = URL.createObjectURL(blob); const anchor = document.createElement("a"); anchor.href = url; anchor.download = "tiku-zong-ocs-config.json"; anchor.click(); URL.revokeObjectURL(url); };
const generateConfig = async () => {
  loadingConfig.value = true;
  try {
    const response = await fetch(`${import.meta.env.VITE_WEB_BASE_API}/api/ocs/config?key=${encodeURIComponent(ocsKey.value.trim())}`, { headers: { Authorization: `Bearer ${getToken()}` } });
    if (!response.ok) throw new Error("配置生成失败");
    ocsConfig.value = JSON.stringify(await response.json(), null, 2);
  } catch { ElMessage.error("OCS 配置生成失败，请确认 API Key 和登录状态"); } finally { loadingConfig.value = false; }
};
onMounted(load);
</script>

<style scoped lang="scss">
.tiku-page { display: flex; flex-direction: column; gap: 16px; padding: 16px; }.tiku-card { border-radius: 10px; }.card-title { display: flex; justify-content: space-between; align-items: center; }.title { font-weight: 700; color: var(--el-text-color-primary); }.key-box { display: flex; justify-content: space-between; align-items: center; padding: 20px 0; }.label { margin-right: 12px; color: var(--el-text-color-secondary); }.created, .hint { color: var(--el-text-color-secondary); font-size: 13px; }.actions { display: flex; gap: 10px; }.plain-key { margin-top: 16px; }.ocs-key-input, .config-input { margin: 16px 0; }.ocs-key-input { max-width: 560px; }
</style>
