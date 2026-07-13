<template>
  <div class="admin-page">
    <el-dialog v-model="grantDialog" title="发放套餐" width="460px" destroy-on-close>
      <el-alert title="发放后立即生效，不生成支付订单；请确认用户与套餐。" type="warning" :closable="false" show-icon />
      <el-form label-position="top" class="grant-form">
        <el-form-item label="用户"><el-input :model-value="grantUser?.email ?? ''" disabled /></el-form-item>
        <el-form-item label="套餐">
          <el-select v-model="grantPackageId" placeholder="请选择已上架套餐" class="full-width">
            <el-option v-for="item in availableGrantPackages" :key="item.id" :label="`${item.name}（${packageTypeLabel(item.type)}）`" :value="item.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer><el-button @click="grantDialog = false">取消</el-button><el-button type="primary" :loading="granting" :disabled="!grantPackageId" @click="grantPackage">确认发放</el-button></template>
    </el-dialog>
    <el-dialog v-model="packageDialog" title="编辑套餐" width="680px">
      <el-form :model="packageEditForm" label-position="top" class="form-grid">
        <el-form-item label="套餐名称"><el-input v-model="packageEditForm.name" /></el-form-item>
        <el-form-item label="类型"><el-select v-model="packageEditForm.type"><el-option label="时间" value="time" /><el-option label="次数" value="count" /><el-option label="时间次数" value="time_count" /></el-select></el-form-item>
        <el-form-item label="有效期（秒）"><el-input-number v-model="packageEditForm.durationSeconds" :min="0" /></el-form-item>
        <el-form-item label="普通次数"><el-input-number v-model="packageEditForm.totalCount" :min="0" /></el-form-item>
        <el-form-item label="AI 次数"><el-input-number v-model="packageEditForm.aiCount" :min="0" /></el-form-item>
        <el-form-item label="价格（分）"><el-input-number v-model="packageEditForm.priceCents" :min="0" /></el-form-item>
        <el-form-item label="限购次数"><el-input-number v-model="packageEditForm.limitCount" :min="0" /></el-form-item>
        <el-form-item label="试用"><el-switch v-model="packageEditForm.isTrial" :active-value="1" :inactive-value="0" /></el-form-item>
        <el-form-item label="免费"><el-switch v-model="packageEditForm.isFree" :active-value="1" :inactive-value="0" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="packageDialog = false">取消</el-button><el-button type="primary" :loading="savingPackage" @click="savePackageEdit">保存</el-button></template>
    </el-dialog>
    <el-dialog v-model="modelDialog" title="编辑 AI 模型" width="680px">
      <el-form :model="modelEditForm" label-position="top" class="form-grid">
        <el-form-item label="Provider ID"><el-input-number v-model="modelEditForm.providerId" :min="1" /></el-form-item>
        <el-form-item label="模型名称"><el-input v-model="modelEditForm.name" /></el-form-item>
        <el-form-item label="优先级（数值越小越优先）"><el-input-number v-model="modelEditForm.priority" :min="1" /></el-form-item>
        <el-form-item label="超时（秒）"><el-input-number v-model="modelEditForm.timeoutSeconds" :min="1" /></el-form-item>
        <el-form-item label="计费方式"><el-select v-model="modelEditForm.billingMode"><el-option label="固定次数" value="fixed" /><el-option label="按 Token" value="token" /><el-option label="按成本" value="cost" /></el-select></el-form-item>
        <el-form-item label="固定次数 / 倍数"><el-input-number v-model="modelEditForm.aiChargeCount" :min="1" /></el-form-item>
        <el-form-item v-if="modelEditForm.billingMode === 'token'" label="每单位 Token"><el-input-number v-model="modelEditForm.tokenUnit" :min="1" /></el-form-item>
        <el-form-item v-if="modelEditForm.billingMode === 'cost'" label="每百万 Token 成本（分）"><el-input-number v-model="modelEditForm.costPerMillionTokensCents" :min="1" /></el-form-item>
        <el-form-item v-if="modelEditForm.billingMode === 'cost'" label="成本加价（%）"><el-input-number v-model="modelEditForm.costMarkupPercent" :min="0" /></el-form-item>
        <el-form-item v-if="modelEditForm.billingMode === 'cost'" label="每 AI 配额单位（分）"><el-input-number v-model="modelEditForm.costUnitCents" :min="1" /></el-form-item>
      </el-form>
      <template #footer><el-button @click="modelDialog = false">取消</el-button><el-button type="primary" :loading="savingModel" @click="saveModelEdit">保存</el-button></template>
    </el-dialog>
    <el-card shadow="never" class="admin-card">
      <template #header><div class="card-title"><span>题库系统管理</span><div class="header-actions"><el-input v-model="adminTotp" type="password" maxlength="6" placeholder="TOTP（启用时填写）" @change="saveTotp" /><el-tag type="warning">管理员接口</el-tag></div></div></template>
      <div class="dashboard-grid">
        <el-card shadow="hover"><div class="metric-value">{{ dashboard.userCount }}</div><div class="metric-label">用户数</div></el-card>
        <el-card shadow="hover"><div class="metric-value">{{ dashboard.paidUserCount }}</div><div class="metric-label">付费用户</div></el-card>
        <el-card shadow="hover"><div class="metric-value">¥{{ (dashboard.paidAmountCents / 100).toFixed(2) }}</div><div class="metric-label">已支付金额</div></el-card>
        <el-card shadow="hover"><div class="metric-value">{{ dashboard.callCount }}</div><div class="metric-label">API 调用</div></el-card>
        <el-card shadow="hover"><div class="metric-value">{{ successRate }}%</div><div class="metric-label">调用成功率</div></el-card>
        <el-card shadow="hover"><div class="metric-value">{{ dashboard.aiCallCount }} / {{ dashboard.ocsCallCount }}</div><div class="metric-label">AI / OCS 调用</div></el-card>
        <el-card shadow="hover"><div class="metric-value">{{ dashboard.averageLatencyMs.toFixed(2) }} ms</div><div class="metric-label">平均响应耗时</div></el-card>
      </div>
      <el-tabs v-model="activeTab">
        <el-tab-pane label="公告管理" name="announcements">
          <el-form :model="announcementForm" label-position="top" class="form-grid">
            <el-form-item label="标题"><el-input v-model="announcementForm.title" maxlength="200" /></el-form-item>
            <el-form-item label="置顶"><el-switch v-model="announcementForm.isPinned" :active-value="1" :inactive-value="0" /></el-form-item>
            <el-form-item label="内容" class="wide-form-item"><el-input v-model="announcementForm.content" type="textarea" :rows="4" /></el-form-item>
          </el-form>
          <div class="import-actions"><el-button type="primary" :loading="savingAnnouncement" @click="saveAnnouncement">发布公告</el-button><el-button @click="resetAnnouncement">清空</el-button><el-button @click="refreshAnnouncements">刷新</el-button></div>
          <el-table :data="announcements" stripe class="table">
            <el-table-column prop="title" label="标题" min-width="220" show-overflow-tooltip />
            <el-table-column label="置顶" width="80"><template #default="scope"><el-tag v-if="scope.row.isPinned" type="warning">置顶</el-tag><span v-else>-</span></template></el-table-column>
            <el-table-column label="状态" width="90"><template #default="scope"><el-tag :type="scope.row.status === 1 ? 'success' : 'info'">{{ scope.row.status === 1 ? '已发布' : '已下架' }}</el-tag></template></el-table-column>
            <el-table-column prop="publishedAt" label="发布时间" min-width="180" />
            <el-table-column label="操作" width="150"><template #default="scope"><el-button link type="primary" @click="editAnnouncement(scope.row)">编辑</el-button><el-button link :type="scope.row.status === 1 ? 'danger' : 'success'" @click="toggleAnnouncement(scope.row)">{{ scope.row.status === 1 ? '下架' : '发布' }}</el-button></template></el-table-column>
          </el-table>
        </el-tab-pane>
        <el-tab-pane label="系统配置" name="settings">
          <el-alert title="注册开关会立即影响新用户注册；站点名称、支持链接和维护公告可由公开配置接口读取。不要在此处保存密钥、数据库密码或支付凭据。" type="warning" :closable="false" show-icon />
          <el-form :model="settingsForm" label-position="top" class="form-grid">
            <el-form-item label="站点名称"><el-input v-model="settingsForm.siteName" maxlength="64" /></el-form-item>
            <el-form-item label="支持链接"><el-input v-model="settingsForm.supportUrl" placeholder="https://example.com/help" /></el-form-item>
            <el-form-item label="允许新用户注册"><el-switch v-model="settingsForm.registrationEnabled" /></el-form-item>
            <el-form-item label="维护公告" class="wide-form-item"><el-input v-model="settingsForm.maintenanceNotice" type="textarea" :rows="3" maxlength="500" show-word-limit /></el-form-item>
          </el-form>
          <el-button type="primary" :loading="savingSettings" @click="saveSettings">保存系统配置</el-button>
        </el-tab-pane>
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
            <el-form-item label="服务商预设"><el-select v-model="providerPreset" clearable placeholder="自定义" @change="applyProviderPreset"><el-option label="DeepSeek" value="deepseek" /><el-option label="豆包（火山方舟）" value="doubao" /><el-option label="通义千问（DashScope）" value="qwen" /><el-option label="自定义 OpenAI 兼容服务" value="custom" /></el-select></el-form-item>
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
            <el-form-item label="计费方式"><el-select v-model="modelForm.billingMode"><el-option label="固定次数" value="fixed" /><el-option label="按 Token" value="token" /><el-option label="按成本" value="cost" /></el-select></el-form-item>
            <el-form-item label="固定次数 / 倍数"><el-input-number v-model="modelForm.aiChargeCount" :min="1" /></el-form-item>
            <el-form-item v-if="modelForm.billingMode === 'token'" label="每单位 Token"><el-input-number v-model="modelForm.tokenUnit" :min="1" /></el-form-item>
            <el-form-item v-if="modelForm.billingMode === 'cost'" label="每百万 Token 成本（分）"><el-input-number v-model="modelForm.costPerMillionTokensCents" :min="1" /></el-form-item>
            <el-form-item v-if="modelForm.billingMode === 'cost'" label="成本加价（%）"><el-input-number v-model="modelForm.costMarkupPercent" :min="0" /></el-form-item>
            <el-form-item v-if="modelForm.billingMode === 'cost'" label="每 AI 配额单位（分）"><el-input-number v-model="modelForm.costUnitCents" :min="1" /></el-form-item>
          </el-form>
          <el-button type="primary" @click="saveModel">创建模型</el-button>
          <el-alert title="按 Token：向上取整(总 Token ÷ 每单位 Token) × 倍数；按成本：成本加价后按“每 AI 配额单位”向上取整。服务商未返回 Token 时，回退扣除固定次数。" type="info" :closable="false" show-icon />
          <el-table :data="models" stripe class="table"><el-table-column prop="id" label="ID" width="80" /><el-table-column prop="providerName" label="服务商" /><el-table-column prop="name" label="模型" /><el-table-column prop="priority" label="优先级" /><el-table-column prop="billingMode" label="计费方式" width="110" /><el-table-column prop="aiChargeCount" label="固定/倍数" width="105" /><el-table-column label="状态" width="90"><template #default="scope"><el-tag :type="scope.row.enabled === 1 ? 'success' : 'info'">{{ scope.row.enabled === 1 ? '启用' : '停用' }}</el-tag></template></el-table-column><el-table-column prop="keyConfigured" label="密钥已配置" /><el-table-column label="操作" width="150" fixed="right"><template #default="scope"><el-button link type="primary" @click="editModel(scope.row)">编辑</el-button><el-button link :type="scope.row.enabled === 1 ? 'danger' : 'success'" @click="toggleModel(scope.row)">{{ scope.row.enabled === 1 ? '停用' : '启用' }}</el-button></template></el-table-column></el-table>
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
          <el-alert title="双击套餐行可编辑套餐配置" type="info" :closable="false" class="import-tip" />
          <el-table :data="adminPackages" stripe class="table" @row-dblclick="editPackage">
            <el-table-column label="操作" width="90"><template #default="scope"><el-button link type="primary" @click="editPackage(scope.row)">编辑</el-button></template></el-table-column>
            <el-table-column prop="name" label="套餐" min-width="160" />
            <el-table-column prop="type" label="类型" width="120" />
            <el-table-column label="价格" width="100"><template #default="scope">¥{{ (scope.row.priceCents / 100).toFixed(2) }}</template></el-table-column>
            <el-table-column prop="totalCount" label="次数" width="90" />
            <el-table-column prop="limitCount" label="限购" width="90" />
            <el-table-column label="状态" width="90"><template #default="scope"><el-tag :type="scope.row.status === 1 ? 'success' : 'info'">{{ scope.row.status === 1 ? '上架' : '下架' }}</el-tag></template></el-table-column>
            <el-table-column label="操作" width="100"><template #default="scope"><el-button link :type="scope.row.status === 1 ? 'danger' : 'success'" @click="togglePackage(scope.row)">{{ scope.row.status === 1 ? '下架' : '上架' }}</el-button></template></el-table-column>
          </el-table>
          <el-divider />
          <el-form :model="couponForm" label-position="top" class="form-grid">
            <el-form-item label="优惠券码"><el-input v-model="couponForm.code" /></el-form-item>
            <el-form-item label="折扣类型"><el-select v-model="couponForm.discountType"><el-option label="固定金额（分）" value="fixed" /><el-option label="百分比" value="percent" /></el-select></el-form-item>
            <el-form-item label="折扣值"><el-input-number v-model="couponForm.discountValue" :min="1" /></el-form-item>
            <el-form-item label="总数量"><el-input-number v-model="couponForm.totalLimit" :min="0" /></el-form-item>
          </el-form>
          <el-button type="primary" @click="saveCoupon">创建优惠券</el-button>
          <el-table :data="coupons" stripe class="table" @row-dblclick="toggleCoupon">
            <el-table-column label="操作" width="90"><template #default="scope"><el-button link :type="scope.row.status === 1 ? 'danger' : 'success'" @click="toggleCoupon(scope.row)">{{ scope.row.status === 1 ? '停用' : '启用' }}</el-button></template></el-table-column>
            <el-table-column prop="code" label="优惠码" min-width="150" />
            <el-table-column label="优惠" width="100"><template #default="scope">{{ scope.row.discountType === 'percent' ? `${scope.row.discountValue}%` : `¥${(scope.row.discountValue / 100).toFixed(2)}` }}</template></el-table-column>
            <el-table-column label="使用情况" width="120"><template #default="scope">{{ scope.row.usedCount }} / {{ scope.row.totalLimit || '不限' }}</template></el-table-column>
            <el-table-column prop="reservedCount" label="锁定" width="80" />
            <el-table-column prop="expiresAt" label="到期时间" min-width="180" />
            <el-table-column label="状态" width="90"><template #default="scope"><el-tag :type="scope.row.status === 1 ? 'success' : 'info'">{{ scope.row.status === 1 ? '启用' : '停用' }}</el-tag></template></el-table-column>
          </el-table>
        </el-tab-pane>

        <el-tab-pane label="邀请码" name="invites">
          <el-form :model="inviteForm" label-position="top" class="form-grid">
            <el-form-item label="邀请码"><el-input v-model="inviteForm.code" placeholder="例如 NEWUSER2026" /></el-form-item>
            <el-form-item label="最大使用次数"><el-input-number v-model="inviteForm.maxUses" :min="0" /></el-form-item>
            <el-form-item label="到期时间"><el-date-picker v-model="inviteForm.expiresAt" type="datetime" value-format="YYYY-MM-DDTHH:mm:ssZ" placeholder="不填表示长期有效" /></el-form-item>
          </el-form>
          <el-button type="primary" @click="saveInvite">创建邀请码</el-button>
          <el-table :data="invites" stripe class="table"><el-table-column prop="code" label="邀请码" /><el-table-column prop="usedCount" label="已使用" width="100" /><el-table-column label="次数上限" width="110"><template #default="scope">{{ scope.row.maxUses || '不限' }}</template></el-table-column><el-table-column prop="expiresAt" label="到期时间" /><el-table-column label="状态" width="90"><template #default="scope"><el-tag :type="scope.row.status === 1 ? 'success' : 'info'">{{ scope.row.status === 1 ? '启用' : '停用' }}</el-tag></template></el-table-column></el-table>
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
            <el-table-column label="操作" width="170" fixed="right">
              <template #default="scope"><el-button link type="primary" @click="openGrantDialog(scope.row)">发放套餐</el-button><el-button link :type="scope.row.status === 1 ? 'danger' : 'success'" @click="toggleStatus(scope.row)">{{ scope.row.status === 1 ? '禁用' : '启用' }}</el-button></template>
            </el-table-column>
          </el-table>
          <div class="pagination"><el-pagination v-model:current-page="userPage" v-model:page-size="userPageSize" layout="total, sizes, prev, pager, next" :total="userTotal" @current-change="refreshUsers" @size-change="refreshUsers" /></div>
        </el-tab-pane>

        <el-tab-pane label="题库管理" name="questions">
          <el-alert
            title="支持 JSON、CSV、XLSX 导入；答案请填写选项文字，多选答案使用 ### 分隔。"
            type="info"
            :closable="false"
            show-icon
            class="import-tip"
          />
          <div class="import-panel">
            <el-input
              v-model="importText"
              type="textarea"
              :rows="5"
              placeholder='[{"question":"题目","type":"single","options":[{"key":"A","text":"选项一"}],"answer":"选项一"}]'
            />
            <div class="import-actions">
              <el-button type="primary" :loading="importing" @click="importQuestionData">批量导入</el-button>
              <el-upload accept=".csv,.xlsx" :auto-upload="false" :show-file-list="false" :on-change="importQuestionFile">
                <el-button :loading="importing" plain>导入 CSV / Excel</el-button>
              </el-upload>
              <el-button @click="importText = importExample">填入示例</el-button>
              <el-button @click="importText = ''">清空</el-button>
            </div>
            <el-alert v-if="importResult" :title="importResult" type="success" :closable="false" />
            <el-alert v-if="importErrors.length" :title="`已跳过 ${importErrors.length} 条错误数据，请修正后重新上传。`" type="warning" :closable="false" />
            <el-table v-if="importErrors.length" :data="importErrors" size="small" max-height="180">
              <el-table-column prop="row" label="行号" width="80" />
              <el-table-column prop="question" label="题目" min-width="220" show-overflow-tooltip />
              <el-table-column prop="message" label="错误原因" min-width="220" />
            </el-table>
          </div>
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
            <el-button type="primary" @click="refreshOrders">查询</el-button><el-button @click="closeOrders">关闭过期订单</el-button><el-button type="warning" @click="runReconciliation">订单对账</el-button>
          </div>
          <el-table :data="orders" stripe class="table">
            <el-table-column prop="orderNo" label="订单号" min-width="210" />
            <el-table-column prop="userEmail" label="用户" min-width="190" />
            <el-table-column prop="packageName" label="套餐" min-width="150" />
            <el-table-column label="应付" width="100"><template #default="scope">¥{{ (scope.row.payableCents / 100).toFixed(2) }}</template></el-table-column>
            <el-table-column prop="status" label="状态" width="130" />
            <el-table-column prop="createdAt" label="创建时间" min-width="180" />
            <el-table-column label="操作" width="180" fixed="right"><template #default="scope"><el-button link type="info" @click="showRefunds(scope.row)">退款明细</el-button><el-button v-if="scope.row.status === 'paid' || scope.row.status === 'partial_refunded'" link type="warning" @click="refund(scope.row)">退款</el-button></template></el-table-column>
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
        <el-tab-pane label="操作日志" name="auditLogs">
          <div class="toolbar question-toolbar">
            <el-input v-model="auditSearch" clearable placeholder="搜索管理员邮箱或资源" @keyup.enter="refreshAuditLogs" />
            <el-button type="primary" @click="refreshAuditLogs">查询</el-button>
          </div>
          <el-alert title="仅记录管理员成功执行的写操作；为防止敏感信息泄露，不保存请求正文、密码、密钥或支付凭据。" type="info" :closable="false" show-icon />
          <el-table :data="auditLogs" stripe class="table">
            <el-table-column prop="adminEmail" label="管理员" min-width="190" />
            <el-table-column prop="action" label="动作" width="90" />
            <el-table-column prop="resource" label="资源" min-width="220" show-overflow-tooltip />
            <el-table-column prop="ipAddress" label="IP" min-width="130" />
            <el-table-column prop="httpStatus" label="HTTP" width="90" />
            <el-table-column prop="createdAt" label="时间" min-width="190" />
          </el-table>
          <div class="pagination"><el-pagination v-model:current-page="auditPage" v-model:page-size="auditPageSize" layout="total, sizes, prev, pager, next" :total="auditTotal" @current-change="refreshAuditLogs" @size-change="refreshAuditLogs" /></div>
        </el-tab-pane>
        <el-tab-pane label="答案反馈" name="feedback">
          <div class="toolbar question-toolbar">
            <el-input v-model="feedbackSearch" clearable placeholder="搜索邮箱、请求 ID 或题目哈希" @keyup.enter="refreshFeedback" />
            <el-select v-model="feedbackType" clearable placeholder="反馈类型" @change="refreshFeedback">
              <el-option label="正确" value="correct" /><el-option label="错误" value="incorrect" /><el-option label="题目不匹配" value="mismatch" /><el-option label="解析错误" value="parse_error" /><el-option label="其他" value="other" />
            </el-select>
            <el-button type="primary" @click="refreshFeedback">查询</el-button>
          </div>
          <el-table :data="feedbackItems" stripe class="table">
            <el-table-column prop="userEmail" label="用户" min-width="190" />
            <el-table-column prop="requestId" label="请求 ID" min-width="210" show-overflow-tooltip />
            <el-table-column prop="questionHash" label="题目哈希" min-width="260" show-overflow-tooltip />
            <el-table-column prop="feedbackType" label="反馈类型" width="120" />
            <el-table-column prop="comment" label="备注" min-width="220" show-overflow-tooltip />
            <el-table-column prop="createdAt" label="提交时间" min-width="190" />
          </el-table>
          <div class="pagination"><el-pagination v-model:current-page="feedbackPage" v-model:page-size="feedbackPageSize" layout="total, sizes, prev, pager, next" :total="feedbackTotal" @current-change="refreshFeedback" @size-change="refreshFeedback" /></div>
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
    <el-dialog v-model="refundDialog" title="退款明细" width="760px">
      <el-table :data="refunds" stripe>
        <el-table-column prop="refundNo" label="退款号" min-width="230" />
        <el-table-column label="金额" width="110"><template #default="scope">¥{{ (scope.row.amountCents / 100).toFixed(2) }}</template></el-table-column>
        <el-table-column prop="status" label="状态" width="110" />
        <el-table-column prop="reason" label="原因" min-width="180" show-overflow-tooltip />
        <el-table-column prop="createdAt" label="创建时间" min-width="180" />
      </el-table>
      <el-empty v-if="!refunds.length" description="暂无退款记录" />
    </el-dialog>
    <el-dialog v-model="reconciliationDialog" title="订单对账结果" width="820px">
      <el-alert v-if="!reconciliationIssues.length" title="未发现订单异常" type="success" :closable="false" show-icon />
      <el-table v-else :data="reconciliationIssues" stripe>
        <el-table-column prop="orderNo" label="订单号" min-width="220" />
        <el-table-column prop="issueType" label="异常类型" width="210" />
        <el-table-column prop="detail" label="详情" min-width="300" show-overflow-tooltip />
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, type UploadFile } from "element-plus";
import { closeExpiredOrders, configurePaymentGateway, createAdminPackage, createAiModel, createAiProvider, createCoupon, createInvite, createOcsSource, createAnnouncement, getAdminAiAnswer, getAdminQuestion, getAdminSettings, getDashboardStats, getPaymentGateway, grantAdminPackage, importQuestionFile as uploadQuestionFile, importQuestions, listAdminAiAnswers, listAdminAuditLogs, listAdminCalls, listAdminFeedback, listAdminOrders, listAdminUsers, listAdminQuestions, listAdminPackages, listAdminCoupons, listAdminAnnouncements, listAiModels, listInvites, listOrderRefunds, listOcsSources, reconcileOrders, refundOrder, updateAdminAiAnswerStatus, updateAdminQuestionStatus, updateAdminUserRole, updateAdminUserStatus, updateAdminPackageStatus, updateAdminPackage, updateAdminCouponStatus, updateAdminSettings, updateAiModel, updateAiModelStatus, updateAnnouncement, updateAnnouncementStatus, type AdminAiAnswer, type AdminAuditLog, type AdminCallLog, type AdminFeedbackItem, type AdminOrderItem, type AdminQuestionDetail, type AdminQuestionItem, type AdminUserItem, type DashboardStats, type InviteItem, type ReconciliationIssue, type RefundItem, type PackageItem, type CouponItem, type AnnouncementItem, type SystemSettings } from "@/api/tiku";
import { QUESTION_TYPES } from "@/constants/question";

const activeTab = ref("ocs");
const dashboard = reactive<DashboardStats>({ userCount: 0, paidUserCount: 0, paidOrderCount: 0, paidAmountCents: 0, callCount: 0, successfulCalls: 0, aiCallCount: 0, ocsCallCount: 0, averageLatencyMs: 0 });
const successRate = computed(() => dashboard.callCount ? ((dashboard.successfulCalls / dashboard.callCount) * 100).toFixed(1) : "0.0");
const adminTotp = ref(sessionStorage.getItem("koi-admin-totp") || "");
const saving = ref(false);
const ocsSources = ref<any[]>([]);
const models = ref<any[]>([]);
const ocsForm = reactive({ name: "", url: "", method: "GET", dataText: '{"q":"${title}"}', successPath: "code", successValue: "1", questionPath: "q", answerPath: "data" });
const providerForm = reactive({ name: "", baseUrl: "", apiKey: "" });
const providerPreset = ref("");
const providerPresets: Record<string, { name: string; baseUrl: string }> = {
  deepseek: { name: "DeepSeek", baseUrl: "https://api.deepseek.com" },
  doubao: { name: "豆包", baseUrl: "https://ark.cn-beijing.volces.com/api/v3" },
  qwen: { name: "通义千问", baseUrl: "https://dashscope.aliyuncs.com/compatible-mode/v1" }
};
const modelForm = reactive({ providerId: 1, name: "", priority: 100, aiChargeCount: 1, billingMode: "fixed", tokenUnit: 1000, costPerMillionTokensCents: 0, costMarkupPercent: 0, costUnitCents: 1 });
const modelDialog = ref(false);
const savingModel = ref(false);
const modelEditForm = reactive({ id: 0, providerId: 1, name: "", priority: 100, timeoutSeconds: 30, aiChargeCount: 1, billingMode: "fixed", tokenUnit: 1000, costPerMillionTokensCents: 0, costMarkupPercent: 0, costUnitCents: 1 });
const gatewayForm = reactive({ name: "易支付", baseUrl: "", merchantId: "", secretKey: "" });
const packageForm = reactive({ name: "", type: "count", totalCount: 100, aiCount: 0, priceCents: 0, limitCount: 0 });
const couponForm = reactive({ code: "", discountType: "percent", discountValue: 10, totalLimit: 0 });
const adminPackages = ref<PackageItem[]>([]);
const coupons = ref<CouponItem[]>([]);
const packageDialog = ref(false);
const savingPackage = ref(false);
const packageEditForm = reactive({ id: 0, name: "", type: "count", durationSeconds: 0, totalCount: 100, aiCount: 0, priceCents: 0, limitCount: 0, isTrial: 0, isFree: 0 });
const announcements = ref<AnnouncementItem[]>([]);
const announcementForm = reactive({ id: 0, title: "", content: "", isPinned: 0, status: 1 });
const savingAnnouncement = ref(false);
const settingsForm = reactive<SystemSettings>({ siteName: "题库调用系统", supportUrl: "", maintenanceNotice: "", registrationEnabled: true });
const savingSettings = ref(false);
const users = ref<AdminUserItem[]>([]);
const grantDialog = ref(false);
const granting = ref(false);
const grantUser = ref<AdminUserItem>();
const grantPackageId = ref<number>();
const availableGrantPackages = computed(() => adminPackages.value.filter((item) => item.status === 1));
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
const importing = ref(false);
const importText = ref("");
const importResult = ref("");
const importErrors = ref<{ row: number; question: string; message: string }[]>([]);
const importExample = '[{"question":"下列哪项属于哺乳动物？","type":"single","options":[{"key":"A","text":"鲸鱼"},{"key":"B","text":"鲨鱼"}],"answer":"鲸鱼","subject":"生物","source":"manual"}]';
const questionTypes = QUESTION_TYPES;
const orders = ref<AdminOrderItem[]>([]);
const orderSearch = ref("");
const orderStatus = ref("");
const orderPage = ref(1);
const orderPageSize = ref(20);
const orderTotal = ref(0);
const refunds = ref<RefundItem[]>([]);
const refundDialog = ref(false);
const reconciliationIssues = ref<ReconciliationIssue[]>([]);
const reconciliationDialog = ref(false);
const callLogs = ref<AdminCallLog[]>([]);
const auditLogs = ref<AdminAuditLog[]>([]);
const auditSearch = ref("");
const auditPage = ref(1);
const auditPageSize = ref(20);
const auditTotal = ref(0);
const feedbackItems = ref<AdminFeedbackItem[]>([]);
const feedbackSearch = ref("");
const feedbackType = ref("");
const feedbackPage = ref(1);
const feedbackPageSize = ref(20);
const feedbackTotal = ref(0);
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
const invites = ref<InviteItem[]>([]);
const inviteForm = reactive({ code: "", maxUses: 0, expiresAt: "", status: 1 });
const refresh = async () => { ocsSources.value = (await listOcsSources()).data ?? []; models.value = (await listAiModels()).data ?? []; adminPackages.value = (await listAdminPackages()).data ?? []; coupons.value = (await listAdminCoupons()).data ?? []; await refreshAnnouncements(); const stats = (await getDashboardStats()).data; if (stats) Object.assign(dashboard, stats); };
const refreshUsers = async () => { const result = await listAdminUsers({ page: userPage.value, pageSize: userPageSize.value, search: userSearch.value || undefined, status: userStatus.value }); users.value = result.data?.items ?? []; userTotal.value = result.data?.total ?? 0; };
const refreshQuestions = async () => { const result = await listAdminQuestions({ page: questionPage.value, pageSize: questionPageSize.value, search: questionSearch.value || undefined, subject: questionSubject.value || undefined, type: questionType.value || undefined, status: questionStatus.value }); questions.value = result.data?.items ?? []; questionTotal.value = result.data?.total ?? 0; };
const importQuestionData = async () => {
  if (!importText.value.trim()) {
    ElMessage.warning("请先粘贴题目 JSON");
    return;
  }
  let parsed: unknown;
  try {
    parsed = JSON.parse(importText.value);
  } catch {
    ElMessage.error("JSON 格式不正确");
    return;
  }
  const items = Array.isArray(parsed) ? parsed : (parsed as { items?: unknown })?.items;
  if (!Array.isArray(items) || !items.length) {
    ElMessage.error("导入内容必须是题目数组或包含 items 数组");
    return;
  }
  importing.value = true;
  try {
    const response = await importQuestions(items as any);
    importResult.value = `导入完成：新增 ${response.data.created} 条，重复 ${response.data.duplicates} 条`;
    await refreshQuestions();
  } finally {
    importing.value = false;
  }
};
const importQuestionFile = async (file: UploadFile) => {
  if (!file.raw) return;
  importing.value = true;
  importErrors.value = [];
  importResult.value = "";
  try {
    const result = await uploadQuestionFile(file.raw);
    importErrors.value = result.data.errors ?? [];
    importResult.value = `文件共 ${result.data.total} 条：新增 ${result.data.created} 条，重复跳过 ${result.data.duplicates} 条，无效 ${result.data.invalid} 条`;
    await refreshQuestions();
  } catch {
    ElMessage.error("文件导入失败，请检查表头和内容格式");
  } finally {
    importing.value = false;
  }
};
const refreshOrders = async () => { const result = await listAdminOrders({ page: orderPage.value, pageSize: orderPageSize.value, search: orderSearch.value || undefined, status: orderStatus.value || undefined }); orders.value = result.data?.items ?? []; orderTotal.value = result.data?.total ?? 0; };
const refreshCalls = async () => { callLogs.value = (await listAdminCalls()).data ?? []; };
const refreshAuditLogs = async () => { const result = await listAdminAuditLogs({ page: auditPage.value, pageSize: auditPageSize.value, search: auditSearch.value || undefined }); auditLogs.value = result.data?.items ?? []; auditTotal.value = result.data?.total ?? 0; };
const refreshFeedback = async () => { const result = await listAdminFeedback({ page: feedbackPage.value, pageSize: feedbackPageSize.value, search: feedbackSearch.value || undefined, type: feedbackType.value || undefined }); feedbackItems.value = result.data?.items ?? []; feedbackTotal.value = result.data?.total ?? 0; };
const refreshAiAnswers = async () => { const result = await listAdminAiAnswers({ page: aiPage.value, pageSize: aiPageSize.value, search: aiSearch.value || undefined, provider: aiProvider.value || undefined, model: aiModel.value || undefined, status: aiStatus.value }); aiAnswers.value = result.data?.items ?? []; aiTotal.value = result.data?.total ?? 0; };
const refreshInvites = async () => { invites.value = (await listInvites()).data ?? []; };
const refreshAnnouncements = async () => { announcements.value = (await listAdminAnnouncements()).data ?? []; };
const loadSettings = async () => { Object.assign(settingsForm, (await getAdminSettings()).data); };
const saveTotp = () => sessionStorage.setItem("koi-admin-totp", adminTotp.value.trim());
const applyProviderPreset = (value: string) => { const preset = providerPresets[value]; if (preset) Object.assign(providerForm, preset); };
const saveOcs = async () => { saving.value = true; try { await createOcsSource({ name: ocsForm.name, url: ocsForm.url, method: ocsForm.method, data: JSON.parse(ocsForm.dataText), successPath: ocsForm.successPath, successValue: ocsForm.successValue, questionPath: ocsForm.questionPath, answerPath: ocsForm.answerPath, enabled: true }); ElMessage.success("OCS 源已保存"); await refresh(); } finally { saving.value = false; } };
const saveProvider = async () => { const response = await createAiProvider(providerForm); modelForm.providerId = response.data.id; ElMessage.success(`服务商已创建，Provider ID：${response.data.id}`); await refresh(); };
const saveModel = async () => { await createAiModel(modelForm); ElMessage.success("AI 模型已创建"); await refresh(); };
const editModel = (model: typeof models.value[number]) => { Object.assign(modelEditForm, model); modelDialog.value = true; };
const saveModelEdit = async () => { savingModel.value = true; try { await updateAiModel(modelEditForm.id, modelEditForm); ElMessage.success("AI 模型已更新"); modelDialog.value = false; await refresh(); } finally { savingModel.value = false; } };
const toggleModel = async (model: typeof models.value[number]) => { await updateAiModelStatus(model.id, model.enabled === 1 ? 0 : 1); ElMessage.success("AI 模型状态已更新"); await refresh(); };
const saveGateway = async () => { await configurePaymentGateway({ provider: "epay", enabled: true, ...gatewayForm }); ElMessage.success("支付网关已保存"); await loadGateway(); };
const loadGateway = async () => { try { const response = await getPaymentGateway(); gatewayForm.name = response.data.name; gatewayForm.baseUrl = response.data.baseUrl; gatewayForm.merchantId = response.data.merchantId; gatewayForm.secretKey = ""; } catch { gatewayForm.secretKey = ""; } };
const savePackage = async () => { await createAdminPackage(packageForm); ElMessage.success("套餐已创建"); await refresh(); };
const togglePackage = async (item: PackageItem) => { await updateAdminPackageStatus(item.id, item.status === 1 ? 0 : 1); ElMessage.success("套餐状态已更新"); await refresh(); };
const editPackage = (item: PackageItem) => { Object.assign(packageEditForm, { ...item, durationSeconds: item.durationSeconds ?? 0 }); packageEditForm.id = item.id; packageDialog.value = true; };
const savePackageEdit = async () => { savingPackage.value = true; try { await updateAdminPackage(packageEditForm.id, packageEditForm); ElMessage.success("套餐已更新"); packageDialog.value = false; await refresh(); } finally { savingPackage.value = false; } };
const saveCoupon = async () => { await createCoupon(couponForm); ElMessage.success("优惠券已创建"); await refresh(); };
const toggleCoupon = async (item: CouponItem) => { await updateAdminCouponStatus(item.id, item.status === 1 ? 0 : 1); ElMessage.success("优惠券状态已更新"); await refresh(); };
const resetAnnouncement = () => { announcementForm.id = 0; announcementForm.title = ""; announcementForm.content = ""; announcementForm.isPinned = 0; announcementForm.status = 1; };
const saveAnnouncement = async () => { savingAnnouncement.value = true; try { if (announcementForm.id) { await updateAnnouncement(announcementForm.id, announcementForm); ElMessage.success("公告已更新"); } else { await createAnnouncement(announcementForm); ElMessage.success("公告已发布"); } resetAnnouncement(); await refreshAnnouncements(); } finally { savingAnnouncement.value = false; } };
const editAnnouncement = (item: AnnouncementItem) => { announcementForm.id = item.id; announcementForm.title = item.title; announcementForm.content = item.content; announcementForm.isPinned = item.isPinned; announcementForm.status = item.status; activeTab.value = "announcements"; };
const toggleAnnouncement = async (item: AnnouncementItem) => { await updateAnnouncementStatus(item.id, item.status === 1 ? 0 : 1); ElMessage.success("公告状态已更新"); await refreshAnnouncements(); };
const saveSettings = async () => { savingSettings.value = true; try { Object.assign(settingsForm, (await updateAdminSettings(settingsForm)).data); ElMessage.success("系统配置已保存"); } finally { savingSettings.value = false; } };
const toggleStatus = async (user: AdminUserItem) => { await updateAdminUserStatus(user.id, user.status === 1 ? 0 : 1); ElMessage.success("用户状态已更新"); await refreshUsers(); };
const changeRole = async (user: AdminUserItem, role: "admin" | "user") => { try { await updateAdminUserRole(user.id, role); ElMessage.success("用户角色已更新"); } catch { await refreshUsers(); } };
const packageTypeLabel = (type: PackageItem["type"]) => ({ time: "时间套餐", count: "次数套餐", time_count: "时间次数套餐" }[type]);
const openGrantDialog = (user: AdminUserItem) => {
  grantUser.value = user;
  grantPackageId.value = availableGrantPackages.value[0]?.id;
  grantDialog.value = true;
};
const grantPackage = async () => {
  if (!grantUser.value || !grantPackageId.value) return;
  granting.value = true;
  try {
    await grantAdminPackage(grantPackageId.value, grantUser.value.id);
    ElMessage.success(`已向 ${grantUser.value.email} 发放套餐`);
    grantDialog.value = false;
    await refresh();
  } finally {
    granting.value = false;
  }
};
const toggleQuestion = async (question: AdminQuestionItem) => { await updateAdminQuestionStatus(question.id, question.status === 1 ? 0 : 1); ElMessage.success("题目状态已更新"); await refreshQuestions(); };
const showQuestion = async (id: number) => { questionDetail.value = (await getAdminQuestion(id)).data; questionDialog.value = true; };
const showRefunds = async (order: AdminOrderItem) => { refunds.value = (await listOrderRefunds(order.orderNo)).data ?? []; refundDialog.value = true; };
const runReconciliation = async () => { reconciliationIssues.value = (await reconcileOrders()).data?.issues ?? []; reconciliationDialog.value = true; };
const closeOrders = async () => { const result = await closeExpiredOrders(); ElMessage.success(`已关闭 ${result.data?.count ?? 0} 个过期订单`); await refreshOrders(); };
const refund = async (order: AdminOrderItem) => { const value = window.prompt(`请输入退款金额（分），最多 ${order.payableCents - order.refundedCents}`); const amount = Number(value); if (!Number.isInteger(amount) || amount <= 0) return; const refundNo = `admin-${order.orderNo}-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`; await refundOrder(order.orderNo, { amountCents: amount, reason: "管理员后台退款", refundNo }); ElMessage.success("退款已记录"); await refreshOrders(); await showRefunds(order); };
const toggleAiAnswer = async (answer: AdminAiAnswer) => { await updateAdminAiAnswerStatus(answer.id, answer.status === 1 ? 0 : 1); ElMessage.success("AI 答案状态已更新"); await refreshAiAnswers(); };
const showAiAnswer = async (id: number) => { aiDetail.value = (await getAdminAiAnswer(id)).data; aiDialog.value = true; };
const saveInvite = async () => { await createInvite({ ...inviteForm, expiresAt: inviteForm.expiresAt || undefined }); ElMessage.success("邀请码已创建"); inviteForm.code = ""; await refreshInvites(); };
onMounted(async () => { await refresh(); await loadGateway(); await loadSettings(); await refreshUsers(); await refreshQuestions(); await refreshOrders(); await refreshCalls(); await refreshAuditLogs(); await refreshFeedback(); await refreshAiAnswers(); await refreshInvites(); });
</script>

<style scoped lang="scss">
.admin-page { padding: 16px; }.admin-card { border-radius: 10px; }.dashboard-grid { display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); gap: 12px; margin-bottom: 16px; }.dashboard-grid :deep(.el-card__body) { padding: 14px; }.metric-value { font-size: 22px; font-weight: 700; color: var(--el-color-primary); white-space: nowrap; }.metric-label { color: var(--el-text-color-secondary); margin-top: 6px; font-size: 13px; }.card-title, .header-actions { display: flex; align-items: center; justify-content: space-between; gap: 8px; font-weight: 700; }.header-actions :deep(.el-input) { width: 190px; }.form-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 0 16px; }.grant-form { margin-top: 16px; }.full-width { width: 100%; }.import-tip { margin-bottom: 12px; }.import-panel { display: flex; flex-direction: column; gap: 10px; margin-bottom: 16px; }.import-actions { display: flex; gap: 10px; }.toolbar { display: flex; gap: 12px; max-width: 900px; }.toolbar .el-input { width: 240px; }.toolbar .el-select { width: 150px; }.table { margin-top: 18px; }.pagination { display: flex; justify-content: flex-end; margin-top: 18px; }.detail-text { white-space: pre-wrap; line-height: 1.7; }.detail-list { white-space: pre-wrap; line-height: 1.8; }.raw-text { max-height: 260px; overflow: auto; white-space: pre-wrap; word-break: break-word; } @media (max-width: 1200px) { .dashboard-grid { grid-template-columns: repeat(4, minmax(0, 1fr)); } } @media (max-width: 900px) { .form-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }.question-toolbar { flex-wrap: wrap; } } @media (max-width: 600px) { .dashboard-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }.form-grid { grid-template-columns: 1fr; }.toolbar { flex-wrap: wrap; }.toolbar .el-input { width: 100%; } }
</style>
