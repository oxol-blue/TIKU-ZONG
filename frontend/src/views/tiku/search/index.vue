<template>
  <div class="tiku-page">
    <el-card shadow="never" class="tiku-card">
      <template #header>
        <div class="card-title">
          <div>
            <div class="title">在线搜题</div>
            <div class="subtitle">支持纯文本题目；题库无结果时由 AI 进行兜底。</div>
          </div>
          <el-tag type="success" effect="light">单题搜索</el-tag>
        </div>
      </template>
      <el-form label-position="top" @submit.prevent="search">
        <el-form-item label="题目">
          <el-input v-model="form.q" type="textarea" :rows="5" maxlength="2000" show-word-limit placeholder="请输入完整题目内容" />
        </el-form-item>
        <div class="form-grid">
          <el-form-item label="题型">
            <el-select v-model="form.type" clearable placeholder="请选择题型">
              <el-option v-for="item in questionTypes" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
          </el-form-item>
          <el-form-item label="选项（每行一个）">
            <el-input v-model="form.options" type="textarea" :rows="2" placeholder="A. 选项一&#10;B. 选项二" />
          </el-form-item>
          <el-form-item label="指定普通套餐">
            <el-select v-model="form.package_id" clearable placeholder="自动使用即将过期套餐">
              <el-option v-for="item in packageInstances" :key="item.id" :label="`${item.packageName}（剩余 ${item.remainingCount < 0 ? '不限' : item.remainingCount}）`" :value="item.packageId" />
            </el-select>
          </el-form-item>
        </div>
        <div class="actions">
          <el-button type="primary" :loading="loading" @click="search">搜索答案</el-button>
          <el-button @click="reset">清空</el-button>
        </div>
      </el-form>
    </el-card>

    <el-card v-if="result" shadow="never" class="tiku-card result-card">
      <template #header>
        <div class="card-title">
          <span class="title">搜索结果</span>
          <div class="result-tags">
            <el-tag :type="result.is_ai ? 'warning' : 'success'">{{ result.is_ai ? 'AI 兜底' : '题库命中' }}</el-tag>
            <el-tag type="info">{{ result.type || '其它' }}</el-tag>
            <span v-if="result.similarity && !result.is_ai">相似度 {{ (result.similarity * 100).toFixed(1) }}%</span>
            <span class="elapsed">耗时 {{ result.search_time }} μs</span>
          </div>
        </div>
      </template>
      <div class="question-text">{{ result.question }}</div>
      <div class="answer-label">答案</div>
      <div class="answer-text">{{ result.answer }}</div>
      <div v-if="result.sources?.length" class="sources">来源：{{ result.sources.join('、') }}</div>
      <div class="feedback-row">
        <span>答案是否有帮助？</span>
        <el-button size="small" type="success" plain @click="feedback('correct')">正确</el-button>
        <el-button size="small" type="danger" plain @click="feedback('incorrect')">错误</el-button>
        <el-button size="small" plain @click="feedback('mismatch')">题目不匹配</el-button>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { listMyPackages, searchQuestion, submitFeedback, type PackageInstance, type QuestionSearchResult } from "@/api/tiku";
import { QUESTION_TYPES } from "@/constants/question";

const loading = ref(false);
const result = ref<QuestionSearchResult>();
const packageInstances = ref<PackageInstance[]>([]);
const form = reactive({ q: "", type: "", options: "", package_id: undefined as number | undefined });
const questionTypes = QUESTION_TYPES;

const loadPackages = async () => {
  try {
    const response = await listMyPackages();
    packageInstances.value = response.data ?? [];
  } catch {
    packageInstances.value = [];
  }
};

const search = async () => {
  if (!form.q.trim()) {
    ElMessage.warning("请输入题目");
    return;
  }
  loading.value = true;
  try {
    const response = await searchQuestion({ ...form, q: form.q.trim() });
    result.value = response.data;
    await loadPackages();
  } finally {
    loading.value = false;
  }
};

const reset = () => {
  form.q = "";
  form.type = "";
  form.options = "";
  form.package_id = undefined;
  result.value = undefined;
};

const feedback = async (feedbackType: string) => {
  if (!result.value?.request_id) return;
  await submitFeedback({ requestId: result.value.request_id, question: result.value.question, feedbackType });
  ElMessage.success("感谢反馈");
};

onMounted(loadPackages);
</script>

<style scoped lang="scss">
.tiku-page { display: flex; flex-direction: column; gap: 16px; padding: 16px; }
.tiku-card { border-radius: 10px; }
.card-title { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
.title { color: var(--el-text-color-primary); font-size: 18px; font-weight: 700; }
.subtitle, .elapsed, .sources { color: var(--el-text-color-secondary); font-size: 13px; }
.subtitle { margin-top: 6px; }
.form-grid { display: grid; grid-template-columns: 1fr 1.5fr 1fr; gap: 16px; }
.actions { display: flex; gap: 10px; }
.result-tags { display: flex; align-items: center; gap: 8px; }
.question-text { white-space: pre-wrap; color: var(--el-text-color-primary); line-height: 1.8; }
.answer-label { margin-top: 20px; color: var(--el-text-color-secondary); font-size: 13px; }
.answer-text { padding: 14px; margin-top: 8px; color: var(--el-color-success); background: var(--el-color-success-light-9); border-radius: 8px; font-size: 18px; font-weight: 700; white-space: pre-wrap; }
.sources { margin-top: 14px; }
.feedback-row { display: flex; align-items: center; gap: 8px; margin-top: 18px; color: var(--el-text-color-secondary); font-size: 13px; }
@media (max-width: 900px) { .form-grid { grid-template-columns: 1fr; gap: 0; } .result-tags { flex-wrap: wrap; } }
</style>
