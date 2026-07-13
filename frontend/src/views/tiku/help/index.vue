<template>
  <div class="tiku-page">
    <el-card shadow="never" class="tiku-card">
      <template #header><div class="hero"><div><h2>{{ settings.siteName }}：使用帮助</h2><p>题库调用系统的常用操作和接入说明</p></div><el-tag type="success">中文帮助</el-tag></div></template>
      <el-alert v-if="settings.maintenanceNotice" :title="settings.maintenanceNotice" type="warning" :closable="false" show-icon class="notice" />
      <el-collapse v-model="activeNames">
        <el-collapse-item title="如何搜索题目？" name="search"><p>进入“在线搜题”，输入完整纯文本题目，可按题型和套餐筛选。题库命中时返回文字答案；题库没有答案时，系统按配置的 AI 优先级进行兜底。</p></el-collapse-item>
        <el-collapse-item title="套餐如何扣费？" name="billing"><p>时间套餐在有效期内不限普通调用次数；次数套餐用完次数失效；时间次数套餐在到期或次数用完时失效。普通命中和 AI 兜底分别计费，失败请求不会扣除次数。</p></el-collapse-item>
        <el-collapse-item title="如何接入 API？" name="api"><p>在“API 与 OCS”创建 API Key，单题接口使用 Query 参数传递 key。请使用 HTTPS，并将 Key 保存在服务端环境变量，不要写入前端源码或公开仓库。</p><pre>GET /api/v1/search?key=你的Key&amp;q=题目内容</pre></el-collapse-item>
        <el-collapse-item title="如何接入 OCS？" name="ocs"><p>在“API 与 OCS”页面生成 AnswererWrapper 配置并下载 JSON。配置已经包含当前 Key 和特殊占位符；重新生成 Key 后，旧配置会立即失效。</p></el-collapse-item>
        <el-collapse-item title="答案不正确怎么办？" name="feedback"><p>在搜索结果下方提交正确、错误或题目不匹配反馈。管理员可在后台查看反馈，并停用不准确的 AI 答案缓存。</p></el-collapse-item>
      </el-collapse>
      <div v-if="settings.supportUrl" class="support"><el-link :href="settings.supportUrl" target="_blank" type="primary">打开支持与帮助页面</el-link></div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { getPublicSettings, type PublicSystemSettings } from "@/api/tiku";
const activeNames = ref(["search"]);
const settings = reactive<PublicSystemSettings>({ siteName: "题库调用系统", supportUrl: "", maintenanceNotice: "", registrationEnabled: true });
onMounted(async () => { try { Object.assign(settings, (await getPublicSettings()).data); } catch { /* The help page remains usable if public settings are unavailable. */ } });
</script>

<style scoped lang="scss">
.tiku-page { padding: 16px; }
.tiku-card { border-radius: 10px; }
.hero { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
h2 { margin: 0; color: var(--el-text-color-primary); }
p { color: var(--el-text-color-regular); line-height: 1.8; }
pre { padding: 12px; overflow: auto; background: var(--el-fill-color-light); border-radius: 6px; }
.notice { margin-bottom: 16px; }.support { margin-top: 18px; }
</style>
