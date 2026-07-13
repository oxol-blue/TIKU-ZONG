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

        <el-tab-pane label="AI 答案管理" name="aiAnswers">
          <div class="toolbar question-toolbar">
            <el-input v-model="aiSearch" clearable placeholder="搜索原始题目" @keyup.enter="refreshAiAnswers" />
            <el-input v-model="aiProvider" clearable placeholder="服务商" @keyup.enter="refreshAiAnswers" />
            <el-input v-model="aiModel" clearable placeholder="模型" @keyup.enter="refreshAiAnswers" />
            <el-select v-model="aiStatus" clearable placeholder="全部状态" @change="refreshAiAnswers"><el-option label="启用" :value="1" /><el-option label="停用" :value="0" /></el-select>
            <el-button type="primary" @click="refreshAiAnswers">查询</el-button>
          </div>
          <el-alert title="AI 答案会直接写入 question_ai；停用仅代表不再命中缓存，不阻止后续重新生成。" type="info" :closable="false" show-icon />
          <el-table :data="aiAnswers" stripe class="table">
            <el-table-column prop="id" label="ID" width="75" /><el-table-column prop="question" label="原始题目" min-width="330" show-overflow-tooltip /><el-table-column prop="answer" label="解析答案" min-width="220" show-overflow-tooltip /><el-table-column prop="provider" label="服务商" width="120" /><el-table-column prop="model" label="模型" width="150" /><el-table-column prop="tokenCount" label="Token" width="85" /><el-table-column label="耗时" width="100"><template #default="scope">{{ (scope.row.elapsedMicros / 1000).toFixed(2) }} ms</template></el-table-column>
            <el-table-column label="状态" width="90"><template #default="scope"><el-tag :type="scope.row.status === 1 ? 'success' : 'info'">{{ scope.row.status === 1 ? '启用' : '停用' }}</el-tag></template></el-table-column>
            <el-table-column label="操作" width="160" fixed="right"><template #default="scope"><el-button link type="primary" @click="showAiAnswer(scope.row.id)">详情</el-button><el-button link :type="scope.row.status === 1 ? 'danger' : 'success'" @click="toggleAiAnswer(scope.row)">{{ scope.row.status === 1 ? '停用' : '启用' }}</el-button></template></el-table-column>
          </el-table>
          <div class="pagination"><el-pagination v-model:current-page="aiPage" v-model:page-size="aiPageSize" layout="total, sizes, prev, pager, next" :total="aiTotal" @current-change="refreshAiAnswers" @size-change="refreshAiAnswers" /></div>
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

        <el-tab-pane label="题库管理" name="questions">
          <div class="toolbar question-toolbar">
            <el-input v-model="questionSearch" clearable placeholder="搜索题目内容" @keyup.enter="refreshQuestions" />
            <el-input v-model="questionSubject" clearable placeholder="科目" @keyup.enter="refreshQuestions" />
            <el-select v-model="questionType" clearable placeholder="题型" @change="refreshQuestions">
              <el-option v-for="item in questionTypes" :key="item.value" :label="item.label" :value="item.value" />
            </el-select>
            <el-select v-model="questionStatus" clearable placeholder="全部状态" @change="refreshQuestions"><el-option label="启用" :value="1" /><el-option label="停用" :value="0" /></el-select>
            <el-button type="primary" @click="refreshQuestions">查询</el-button>
          </div>
          <el-table :data="questions" stripe class="table">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column prop="question" label="题目" min-width="360" show-overflow-tooltip />
            <el-table-column prop="type" label="题型" width="120" />
            <el-table-column prop="subject" label="科目" width="130" />
            <el-table-column prop="platform" label="平台" width="130" />
            <el-table-column prop="source" label="来源" width="150" show-overflow-tooltip />
            <el-table-column label="状态" width="90"><template #default="scope"><el-tag :type="scope.row.status === 1 ? 'success' : 'info'">{{ scope.row.status === 1 ? '启用' : '停用' }}</el-tag></template></el-table-column>
            <el-table-column label="操作" width="160" fixed="right"><template #default="scope"><el-button link type="primary" @click="showQuestion(scope.row.id)">详情</el-button><el-button link :type="scope.row.status === 1 ? 'danger' : 'success'" @click="toggleQuestion(scope.row)">{{ scope.row.status === 1 ? '停用' : '启用' }}</el-button></template></el-table-column>
          </el-table>
          <div class="pagination"><el-pagination v-model:current-page="questionPage" v-model:page-size="questionPageSize" layout="total, sizes, prev, pager, next" :total="questionTotal" @current-change="refreshQuestions" @size-change="refreshQuestions" /></div>
        </el-tab-pane>

        <el-tab-pane label="订单管理" name="orders">
          <div class="toolbar">
            <el-input v-model="orderSearch" clearable placeholder="订单号或用户邮箱" @keyup.enter="refreshOrders" />
            <el-select v-model="orderStatus" clearable placeholder="订单状态" @change="refreshOrders"><el-option label="待支付" value="pending" /><el-option label="已支付" value="paid" /><el-option label="已关闭" value="closed" /><el-option label="部分退款" value="partial_refunded" /><el-option label="已退款" value="refunded" /></el-select>
            <el-button type="primary" @click="refreshOrders">查询</el-button><el-button @click="closeOrders">关闭过期订单</el-button>
          </div>
          <el-table :data="orders" stripe class="table">
            <el-table-column prop="orderNo" label="订单号" min-width="210" />
            <el-table-column prop="userEmail" label="用户" min-width="190" />
            <el-table-column prop="packageName" label="套餐" min-width="150" />
            <el-table-column label="应付" width="100"><template #default="scope">¥{{ (scope.row.payableCents / 100).toFixed(2) }}</template></el-table-column>
            <el-table-column prop="status" label="状态" width="130" />
            <el-table-column prop="createdAt" label="创建时间" min-width="180" />
            <el-table-column label="操作" width="120" fixed="right"><template #default="scope"><el-button v-if="scope.row.status === 'paid' || scope.row.status === 'partial_refunded'" link type="warning" @click="refund(scope.row)">退款</el-button></template></el-table-column>
          </el-table>
          <div class="pagination"><el-pagination v-model:current-page="orderPage" v-model:page-size="orderPageSize" layout="total, sizes, prev, pager, next" :total="orderTotal" @current-change="refreshOrders" @size-change="refreshOrders" /></div>
        </el-tab-pane>

        <el-tab-pane label="调用日志" name="calls">
          <div class="toolbar"><el-button type="primary" @click="refreshCalls">刷新日志</el-button></div>
          <el-table :data="callLogs" stripe class="table">
            <el-table-column prop="requestId" label="请求 ID" min-width="190" />
            <el-table-column prop="userId" label="用户 ID" width="90" />
            <el-table-column prop="endpoint" label="接口" min-width="150" />
            <el-table-column label="结果" width="90"><template #default="scope"><el-tag :type="scope.row.success ? 'success' : 'danger'">{{ scope.row.success ? '成功' : '失败' }}</el-tag></template></el-table-column>
            <el-table-column label="来源" width="80"><template #default="scope">{{ scope.row.isAi ? 'AI' : '题库' }}</template></el-table-column>
            <el-table-column label="耗时" width="100"><template #default="scope">{{ (scope.row.elapsedMicros / 1000).toFixed(2) }} ms</template></el-table-column>
            <el-table-column prop="httpStatus" label="HTTP" width="80" /><el-table-column prop="errorCode" label="错误码" width="140" /><el-table-column prop="createdAt" label="时间" min-width="180" />
          </el-table>
        </el-tab-pane>
      </el-tabs>
    </el-card>
    <el-dialog v-model="questionDialog" title="题目详情" width="720px">
      <template v-if="questionDetail">
        <el-descriptions :column="2" border><el-descriptions-item label="ID">{{ questionDetail.id }}</el-descriptions-item><el-descriptions-item label="题型">{{ questionDetail.type }}</el-descriptions-item><el-descriptions-item label="科目">{{ questionDetail.subject || "-" }}</el-descriptions-item><el-descriptions-item label="平台">{{ questionDetail.platform || "-" }}</el-descriptions-item><el-descriptions-item label="来源" :span="2">{{ questionDetail.source || "-" }}</el-descriptions-item><el-descriptions-item label="题目" :span="2"><div class="detail-text">{{ questionDetail.question }}</div></el-descriptions-item></el-descriptions>
        <el-divider content-position="left">选项</el-divider>
        <div v-if="questionDetail.options.length" class="detail-list"><div v-for="option in questionDetail.options" :key="option.position">{{ option.key }}. {{ option.text }}</div></div><el-empty v-else description="无选项" />
        <el-divider content-position="left">答案（保存为文字）</el-divider>
        <div class="detail-list"><div v-for="answer in questionDetail.answers" :key="answer.position">{{ answer.text }}</div></div>
      </template>
    </el-dialog>
    <el-dialog v-model="aiDialog" title="AI 答案详情" width="820px">
      <template v-if="aiDetail">
        <el-descriptions :column="2" border><el-descriptions-item label="ID">{{ aiDetail.id }}</el-descriptions-item><el-descriptions-item label="题型">{{ aiDetail.type }}</el-descriptions-item><el-descriptions-item label="服务商">{{ aiDetail.provider }}</el-descriptions-item><el-descriptions-item label="模型">{{ aiDetail.model }}</el-descriptions-item><el-descriptions-item label="Token">{{ aiDetail.tokenCount }}</el-descriptions-item><el-descriptions-item label="状态">{{ aiDetail.status === 1 ? '启用' : '停用' }}</el-descriptions-item><el-descriptions-item label="题目" :span="2"><div class="detail-text">{{ aiDetail.question }}</div></el-descriptions-item><el-descriptions-item label="答案" :span="2"><div class="detail-text">{{ aiDetail.answer }}</div></el-descriptions-item><el-descriptions-item label="Prompt" :span="2"><pre class="raw-text">{{ aiDetail.prompt }}</pre></el-descriptions-item><el-descriptions-item label="AI 原始响应" :span="2"><pre class="raw-text">{{ aiDetail.rawResponse }}</pre></el-descriptions-item></el-descriptions>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import { closeExpiredOrders, configurePaymentGateway, createAdminPackage, createAiModel, createAiProvider, createCoupon, createOcsSource, getAdminAiAnswer, getAdminQuestion, listAdminAiAnswers, listAdminCalls, listAdminOrders, listAdminUsers, listAdminQuestions, listAiModels, listOcsSources, refundOrder, updateAdminAiAnswerStatus, updateAdminQuestionStatus, updateAdminUserRole, updateAdminUserStatus, type AdminAiAnswer, type AdminCallLog, type AdminOrderItem, type AdminQuestionDetail, type AdminQuestionItem, type AdminUserItem } from "@/api/tiku";

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
const questions = ref<AdminQuestionItem[]>([]);
const questionDetail = ref<AdminQuestionDetail>();
const questionDialog = ref(false);
const questionSearch = ref("");
const questionSubject = ref("");
const questionType = ref("");
const questionStatus = ref<number | undefined>();
const questionPage = ref(1);
const questionPageSize = ref(20);
const questionTotal = ref(0);
const questionTypes = [{ label: "选择题", value: "single" }, { label: "多选题", value: "multiple" }, { label: "判断题", value: "judge" }, { label: "简答题", value: "short_answer" }, { label: "填空题", value: "fill" }, { label: "其它", value: "other" }];
const orders = ref<AdminOrderItem[]>([]);
const orderSearch = ref("");
const orderStatus = ref("");
const orderPage = ref(1);
const orderPageSize = ref(20);
const orderTotal = ref(0);
const callLogs = ref<AdminCallLog[]>([]);
const aiAnswers = ref<AdminAiAnswer[]>([]);
const aiDetail = ref<AdminAiAnswer>();
const aiDialog = ref(false);
const aiSearch = ref("");
const aiProvider = ref("");
const aiModel = ref("");
const aiStatus = ref<number | undefined>();
const aiPage = ref(1);
const aiPageSize = ref(20);
const aiTotal = ref(0);
const refresh = async () => { ocsSources.value = (await listOcsSources()).data ?? []; models.value = (await listAiModels()).data ?? []; };
const refreshUsers = async () => { const result = await listAdminUsers({ page: userPage.value, pageSize: userPageSize.value, search: userSearch.value || undefined, status: userStatus.value }); users.value = result.data?.items ?? []; userTotal.value = result.data?.total ?? 0; };
const refreshQuestions = async () => { const result = await listAdminQuestions({ page: questionPage.value, pageSize: questionPageSize.value, search: questionSearch.value || undefined, subject: questionSubject.value || undefined, type: questionType.value || undefined, status: questionStatus.value }); questions.value = result.data?.items ?? []; questionTotal.value = result.data?.total ?? 0; };
const refreshOrders = async () => { const result = await listAdminOrders({ page: orderPage.value, pageSize: orderPageSize.value, search: orderSearch.value || undefined, status: orderStatus.value || undefined }); orders.value = result.data?.items ?? []; orderTotal.value = result.data?.total ?? 0; };
const refreshCalls = async () => { callLogs.value = (await listAdminCalls()).data ?? []; };
const refreshAiAnswers = async () => { const result = await listAdminAiAnswers({ page: aiPage.value, pageSize: aiPageSize.value, search: aiSearch.value || undefined, provider: aiProvider.value || undefined, model: aiModel.value || undefined, status: aiStatus.value }); aiAnswers.value = result.data?.items ?? []; aiTotal.value = result.data?.total ?? 0; };
const saveTotp = () => sessionStorage.setItem("koi-admin-totp", adminTotp.value.trim());
const saveOcs = async () => { saving.value = true; try { await createOcsSource({ name: ocsForm.name, url: ocsForm.url, method: ocsForm.method, data: JSON.parse(ocsForm.dataText), successPath: ocsForm.successPath, successValue: ocsForm.successValue, questionPath: ocsForm.questionPath, answerPath: ocsForm.answerPath, enabled: true }); ElMessage.success("OCS 源已保存"); await refresh(); } finally { saving.value = false; } };
const saveProvider = async () => { const response = await createAiProvider(providerForm); modelForm.providerId = response.data.id; ElMessage.success(`服务商已创建，Provider ID：${response.data.id}`); await refresh(); };
const saveModel = async () => { await createAiModel(modelForm); ElMessage.success("AI 模型已创建"); await refresh(); };
const saveGateway = async () => { await configurePaymentGateway({ provider: "epay", enabled: true, ...gatewayForm }); ElMessage.success("支付网关已保存"); };
const savePackage = async () => { await createAdminPackage(packageForm); ElMessage.success("套餐已创建"); };
const saveCoupon = async () => { await createCoupon(couponForm); ElMessage.success("优惠券已创建"); };
const toggleStatus = async (user: AdminUserItem) => { await updateAdminUserStatus(user.id, user.status === 1 ? 0 : 1); ElMessage.success("用户状态已更新"); await refreshUsers(); };
const changeRole = async (user: AdminUserItem, role: "admin" | "user") => { try { await updateAdminUserRole(user.id, role); ElMessage.success("用户角色已更新"); } catch { await refreshUsers(); } };
const toggleQuestion = async (question: AdminQuestionItem) => { await updateAdminQuestionStatus(question.id, question.status === 1 ? 0 : 1); ElMessage.success("题目状态已更新"); await refreshQuestions(); };
const showQuestion = async (id: number) => { questionDetail.value = (await getAdminQuestion(id)).data; questionDialog.value = true; };
const closeOrders = async () => { const result = await closeExpiredOrders(); ElMessage.success(`已关闭 ${result.data?.count ?? 0} 个过期订单`); await refreshOrders(); };
const refund = async (order: AdminOrderItem) => { const value = window.prompt(`请输入退款金额（分），最多 ${order.payableCents - order.refundedCents}`); const amount = Number(value); if (!Number.isInteger(amount) || amount <= 0) return; await refundOrder(order.orderNo, { amountCents: amount, reason: "管理员后台退款" }); ElMessage.success("退款已记录"); await refreshOrders(); };
const toggleAiAnswer = async (answer: AdminAiAnswer) => { await updateAdminAiAnswerStatus(answer.id, answer.status === 1 ? 0 : 1); ElMessage.success("AI 答案状态已更新"); await refreshAiAnswers(); };
const showAiAnswer = async (id: number) => { aiDetail.value = (await getAdminAiAnswer(id)).data; aiDialog.value = true; };
onMounted(async () => { await refresh(); await refreshUsers(); await refreshQuestions(); await refreshOrders(); await refreshCalls(); await refreshAiAnswers(); });
</script>

<style scoped lang="scss">
.admin-page { padding: 16px; }.admin-card { border-radius: 10px; }.card-title, .header-actions { display: flex; align-items: center; justify-content: space-between; gap: 8px; font-weight: 700; }.header-actions :deep(.el-input) { width: 190px; }.form-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 0 16px; }.toolbar { display: flex; gap: 12px; max-width: 900px; }.toolbar .el-input { width: 240px; }.toolbar .el-select { width: 150px; }.table { margin-top: 18px; }.pagination { display: flex; justify-content: flex-end; margin-top: 18px; }.detail-text { white-space: pre-wrap; line-height: 1.7; }.detail-list { white-space: pre-wrap; line-height: 1.8; }.raw-text { max-height: 260px; overflow: auto; white-space: pre-wrap; word-break: break-word; } @media (max-width: 900px) { .form-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }.question-toolbar { flex-wrap: wrap; } } @media (max-width: 600px) { .form-grid { grid-template-columns: 1fr; }.toolbar { flex-wrap: wrap; }.toolbar .el-input { width: 100%; } }
</style>
