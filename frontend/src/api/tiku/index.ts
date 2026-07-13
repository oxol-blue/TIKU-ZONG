import koi from "@/utils/axios.ts";

export interface QuestionSearchResult {
  request_id: string;
  question: string;
  answer: string;
  type: string;
  is_ai: boolean;
  similarity?: number;
  search_time: number;
  sources: string[];
}

export interface PackageItem {
  id: number;
  name: string;
  type: "time" | "count" | "time_count";
  durationSeconds?: number;
  totalCount: number;
  aiCount: number;
  priceCents: number;
  status: number;
  limitCount: number;
  isTrial: number;
  isFree: number;
  createdAt: string;
}

export interface PackageInstance {
  id: number;
  packageId: number;
  packageName: string;
  packageType: string;
  startsAt: string;
  expiresAt?: string;
  remainingCount: number;
  remainingAiCount: number;
  status: number;
}

export interface OrderItem {
  id: number;
  orderNo: string;
  packageId: number;
  packageName: string;
  provider: string;
  amountCents: number;
  payableCents: number;
  refundedCents: number;
  status: string;
  expiresAt: string;
  paidAt?: string;
  createdAt: string;
}

export interface ApiKeyView {
  prefix: string;
  masked: string;
  createdAt: string;
  lastUsedAt?: string;
}

export const searchQuestion = (params: { q: string; type?: string; options?: string; package_id?: number }) =>
  koi.get<{ code: number; message: string; data: QuestionSearchResult }>("/api/v1/search", params);

export const submitFeedback = (data: { requestId: string; question: string; feedbackType: string; comment?: string }) =>
  koi.post<{ code: number; message: string }>("/api/v1/feedback", data);
export interface FeedbackItem {
  id: number;
  requestId: string;
  questionHash: string;
  feedbackType: string;
  comment: string;
  createdAt: string;
}
export const listMyFeedback = (limit = 50) =>
  koi.get<{ code: number; message: string; data: FeedbackItem[] }>("/api/v1/feedback/my", { limit });
export interface AdminFeedbackItem extends FeedbackItem { userId: number; userEmail: string; }
export const listAdminFeedback = (params: { page?: number; pageSize?: number; search?: string; type?: string }) =>
  koi.get<{ code: number; message: string; data: { items: AdminFeedbackItem[]; page: number; pageSize: number; total: number } }>("/api/v1/admin/feedback", params);
export const listMyCalls = (limit = 100) =>
  koi.get<{ code: number; message: string; data: AdminCallLog[] }>("/api/v1/calls/my", { limit });

export interface SearchHistoryItem {
  id: number;
  requestId: string;
  question: string;
  type: string;
  answer: string;
  source: string;
  isAi: boolean;
  elapsedMicros: number;
  createdAt: string;
}
export interface SearchHistoryPage {
  items: SearchHistoryItem[];
  page: number;
  pageSize: number;
  total: number;
}
export const listMySearchHistory = (params: { page?: number; pageSize?: number; isAi?: boolean }) =>
  koi.get<{ code: number; message: string; data: SearchHistoryPage }>("/api/v1/search-history/my", params);

export const listPackages = () => koi.get<{ code: number; message: string; data: PackageItem[] }>("/api/v1/packages");

export const listMyPackages = () =>
  koi.get<{ code: number; message: string; data: PackageInstance[] }>("/api/v1/packages/my");

export const createOrder = (data: { packageId: number; provider?: string; couponCode?: string }) =>
  koi.post<{ code: number; message: string; data: { order: OrderItem; paymentUrl: string } }>("/api/v1/orders", data);

export const listMyOrders = () => koi.get<{ code: number; message: string; data: OrderItem[] }>("/api/v1/orders/my");

export const getApiKey = () => koi.get<{ code: number; message: string; data: ApiKeyView }>("/api/v1/api-key");

export const createApiKey = () =>
  koi.post<{ code: number; message: string; data: { key: string; info: ApiKeyView } }>("/api/v1/api-key");

export const rotateApiKey = () =>
  koi.post<{ code: number; message: string; data: { key: string; info: ApiKeyView } }>("/api/v1/api-key/rotate");

export const revokeApiKey = () => koi.delete<{ code: number; message: string }>("/api/v1/api-key");

export const getOcsConfig = () => koi.get<any>("/api/ocs/config", { key: "" });

export const listOcsSources = () => koi.get<{ code: number; message: string; data: any[] }>("/api/v1/admin/ocs/sources");
export const createOcsSource = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: any }>("/api/v1/admin/ocs/sources", data);
export const updateOcsSource = (id: number, data: Record<string, unknown>) =>
  koi.put<{ code: number; message: string }>(`/api/v1/admin/ocs/sources/${id}`, data);
export const updateOcsSourceStatus = (id: number, status: number) =>
  koi.patch<{ code: number; message: string }>(`/api/v1/admin/ocs/sources/${id}/status`, { status });
export const createAiProvider = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: { id: number } }>("/api/v1/admin/ai/providers", data);
export const createAiModel = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: { id: number } }>("/api/v1/admin/ai/models", data);
export const listAiModels = () => koi.get<{ code: number; message: string; data: any[] }>("/api/v1/admin/ai/models");
export const updateAiModel = (id: number, data: Record<string, unknown>) =>
  koi.put<{ code: number; message: string }>(`/api/v1/admin/ai/models/${id}`, data);
export const updateAiModelStatus = (id: number, status: number) =>
  koi.patch<{ code: number; message: string }>(`/api/v1/admin/ai/models/${id}/status`, { status });
export const configurePaymentGateway = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: any }>("/api/v1/admin/payment/gateways", data);
export const createAdminPackage = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: PackageItem }>("/api/v1/admin/packages", data);
export const listAdminPackages = () =>
  koi.get<{ code: number; message: string; data: PackageItem[] }>("/api/v1/admin/packages");
export const updateAdminPackageStatus = (id: number, status: number) =>
  koi.patch<{ code: number; message: string }>(`/api/v1/admin/packages/${id}/status`, { status });
export const updateAdminPackage = (id: number, data: Record<string, unknown>) =>
  koi.put<{ code: number; message: string; data: PackageItem }>(`/api/v1/admin/packages/${id}`, data);
export const grantAdminPackage = (packageId: number, userId: number) =>
  koi.post<{ code: number; message: string; data: PackageInstance }>(`/api/v1/admin/packages/${packageId}/grant/${userId}`, {});
export const createCoupon = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: any }>("/api/v1/admin/coupons", data);
export interface CouponItem {
  id: number;
  code: string;
  discountType: string;
  discountValue: number;
  totalLimit: number;
  usedCount: number;
  reservedCount: number;
  expiresAt?: string;
  status: number;
}
export const listAdminCoupons = () =>
  koi.get<{ code: number; message: string; data: CouponItem[] }>("/api/v1/admin/coupons");
export const updateAdminCouponStatus = (id: number, status: number) =>
  koi.patch<{ code: number; message: string }>(`/api/v1/admin/coupons/${id}/status`, { status });
export interface AnnouncementItem {
  id: number;
  title: string;
  content: string;
  status: number;
  isPinned: number;
  publishedAt?: string;
  createdAt: string;
  updatedAt: string;
}
export const listAnnouncements = () =>
  koi.get<{ code: number; message: string; data: AnnouncementItem[] }>("/api/v1/announcements");
export const listAdminAnnouncements = () =>
  koi.get<{ code: number; message: string; data: AnnouncementItem[] }>("/api/v1/admin/announcements");
export const createAnnouncement = (data: { title: string; content: string; isPinned: number; status: number }) =>
  koi.post<{ code: number; message: string; data: AnnouncementItem }>("/api/v1/admin/announcements", data);
export const updateAnnouncement = (id: number, data: { title: string; content: string; isPinned: number }) =>
  koi.put<{ code: number; message: string; data: AnnouncementItem }>(`/api/v1/admin/announcements/${id}`, data);
export const updateAnnouncementStatus = (id: number, status: number) =>
  koi.patch<{ code: number; message: string }>(`/api/v1/admin/announcements/${id}/status`, { status });
export interface SystemSettings {
  siteName: string;
  supportUrl: string;
  maintenanceNotice: string;
  registrationEnabled: boolean;
}
export interface PublicSystemSettings {
  siteName: string;
  supportUrl: string;
  maintenanceNotice: string;
  registrationEnabled: boolean;
}
export const getPublicSettings = () =>
  koi.get<{ code: number; message: string; data: PublicSystemSettings }>("/api/v1/settings/public");
export const getAdminSettings = () =>
  koi.get<{ code: number; message: string; data: SystemSettings }>("/api/v1/admin/settings");
export const updateAdminSettings = (data: SystemSettings) =>
  koi.put<{ code: number; message: string; data: SystemSettings }>("/api/v1/admin/settings", data);
export interface PaymentGatewayView { id: number; provider: string; name: string; baseUrl: string; merchantId: string; keyConfigured: boolean; enabled: number; }
export const getPaymentGateway = (provider = "epay") =>
  koi.get<{ code: number; message: string; data: PaymentGatewayView }>("/api/v1/admin/payment/gateways", { provider });

export interface AdminUserItem {
  id: number;
  email: string;
  role: "admin" | "user";
  status: number;
  failedLoginCount: number;
  lockedUntil?: string;
  lastLoginAt?: string;
  createdAt: string;
  apiKeyPrefix?: string;
}

export const listAdminUsers = (params: { page?: number; pageSize?: number; search?: string; status?: number }) =>
  koi.get<{ code: number; message: string; data: { items: AdminUserItem[]; page: number; pageSize: number; total: number } }>("/api/v1/admin/users", params);
export const updateAdminUserStatus = (id: number, status: number) =>
  koi.patch<{ code: number; message: string }>(`/api/v1/admin/users/${id}/status`, { status });
export const updateAdminUserRole = (id: number, role: "admin" | "user") =>
  koi.patch<{ code: number; message: string }>(`/api/v1/admin/users/${id}/role`, { role });

export interface InviteItem {
  id: number;
  code: string;
  maxUses: number;
  usedCount: number;
  status: number;
  expiresAt?: string;
  createdAt: string;
}
export const listInvites = () => koi.get<{ code: number; message: string; data: InviteItem[] }>("/api/v1/admin/invites");
export const createInvite = (data: { code: string; maxUses: number; expiresAt?: string; status: number }) =>
  koi.post<{ code: number; message: string; data: InviteItem }>("/api/v1/admin/invites", data);

export interface AdminQuestionItem {
  id: number;
  question: string;
  type: string;
  platform: string;
  subject: string;
  source: string;
  status: number;
  collectedAt?: string;
  createdAt: string;
  optionCount: number;
  answerCount: number;
}

export interface AdminQuestionDetail extends AdminQuestionItem {
  options: { key: string; text: string; position: number }[];
  answers: { text: string; raw: string; position: number }[];
}

export const listAdminQuestions = (params: { page?: number; pageSize?: number; search?: string; type?: string; subject?: string; status?: number }) =>
  koi.get<{ code: number; message: string; data: { items: AdminQuestionItem[]; page: number; pageSize: number; total: number } }>("/api/v1/admin/questions", params);
export const getAdminQuestion = (id: number) =>
  koi.get<{ code: number; message: string; data: AdminQuestionDetail }>(`/api/v1/admin/questions/${id}`);
export const updateAdminQuestionStatus = (id: number, status: number) =>
  koi.patch<{ code: number; message: string }>(`/api/v1/admin/questions/${id}/status`, { status });
export interface QuestionImportItem {
  question: string;
  type?: string;
  options?: { key: string; text: string }[];
  answer: string;
  answerRaw?: string;
  platform?: string;
  subject?: string;
  source?: string;
  collectedAt?: string;
}
export const importQuestions = (items: QuestionImportItem[]) =>
  koi.post<{ code: number; message: string; data: { created: number; duplicates: number } }>("/api/v1/admin/questions/import", { items });
export interface QuestionFileImportReport {
  total: number;
  valid: number;
  created: number;
  duplicates: number;
  invalid: number;
  preview: QuestionImportItem[];
  errors: { row: number; question: string; message: string }[];
}
export const importQuestionFile = (file: File) => {
  const formData = new FormData();
  formData.append("file", file);
  return koi.upload<{ code: number; message: string; data: QuestionFileImportReport }>("/api/v1/admin/questions/import/file", formData);
};

export interface AdminOrderItem extends OrderItem {
  userId: number;
  userEmail: string;
  amountCents: number;
  payableCents: number;
  discountCents?: number;
  providerTradeNo?: string;
  packageInstanceId?: number;
  closedAt?: string;
}

export interface AdminCallLog {
  requestId: string;
  userId?: number;
  apiKeyId?: number;
  endpoint: string;
  questionHash: string;
  success: boolean;
  isAi: boolean;
  elapsedMicros: number;
  httpStatus: number;
  errorCode: string;
  createdAt: string;
}

export const listAdminOrders = (params: { page?: number; pageSize?: number; search?: string; status?: string }) =>
  koi.get<{ code: number; message: string; data: { items: AdminOrderItem[]; page: number; pageSize: number; total: number } }>("/api/v1/admin/orders", params);
export const closeExpiredOrders = () =>
  koi.post<{ code: number; message: string; data: { count: number } }>("/api/v1/admin/orders/close-expired");
export const refundOrder = (orderNo: string, data: { amountCents: number; reason?: string; refundNo?: string }) =>
  koi.post<{ code: number; message: string; data: AdminOrderItem }>(`/api/v1/admin/orders/${encodeURIComponent(orderNo)}/refund`, data);
export interface RefundItem {
  id: number;
  refundNo: string;
  orderNo: string;
  amountCents: number;
  reason: string;
  status: string;
  createdAt: string;
}
export const listOrderRefunds = (orderNo: string) =>
  koi.get<{ code: number; message: string; data: RefundItem[] }>(`/api/v1/admin/orders/${encodeURIComponent(orderNo)}/refunds`);
export interface ReconciliationIssue {
  orderNo: string;
  issueType: string;
  detail: string;
}
export const reconcileOrders = () =>
  koi.get<{ code: number; message: string; data: { issues: ReconciliationIssue[]; count: number } }>("/api/v1/admin/orders/reconciliation");
export const listAdminCalls = (limit = 100) =>
  koi.get<{ code: number; message: string; data: AdminCallLog[] }>("/api/v1/admin/calls", { limit });
export interface AdminAuditLog {
  id: number;
  adminId: number;
  adminEmail: string;
  action: string;
  resource: string;
  requestPath: string;
  ipAddress: string;
  httpStatus: number;
  createdAt: string;
}
export const listAdminAuditLogs = (params: { page?: number; pageSize?: number; search?: string }) =>
  koi.get<{ code: number; message: string; data: { items: AdminAuditLog[]; page: number; pageSize: number; total: number } }>("/api/v1/admin/audit-logs", params);
export interface DashboardStats {
  userCount: number;
  paidUserCount: number;
  paidOrderCount: number;
  paidAmountCents: number;
  callCount: number;
  successfulCalls: number;
  aiCallCount: number;
  ocsCallCount: number;
  onlineSearchCount: number;
  localHitCount: number;
  ocsHitCount: number;
  tokenCount: number;
  packageConsumeCount: number;
  errorRate: number;
  averageLatencyMs: number;
}
export const getDashboardStats = () =>
  koi.get<{ code: number; message: string; data: DashboardStats }>("/api/v1/admin/dashboard");

export interface AdminAiAnswer {
  id: number;
  questionHash: string;
  question: string;
  type: string;
  answer: string;
  prompt: string;
  rawResponse: string;
  provider: string;
  model: string;
  tokenCount: number;
  elapsedMicros: number;
  status: number;
  createdAt: string;
  updatedAt: string;
}

export const listAdminAiAnswers = (params: { page?: number; pageSize?: number; search?: string; provider?: string; model?: string; status?: number }) =>
  koi.get<{ code: number; message: string; data: { items: AdminAiAnswer[]; page: number; pageSize: number; total: number } }>("/api/v1/admin/ai/answers", params);
export const getAdminAiAnswer = (id: number) =>
  koi.get<{ code: number; message: string; data: AdminAiAnswer }>(`/api/v1/admin/ai/answers/${id}`);
export const updateAdminAiAnswerStatus = (id: number, status: number) =>
  koi.patch<{ code: number; message: string }>(`/api/v1/admin/ai/answers/${id}/status`, { status });
