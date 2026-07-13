import koi from "@/utils/axios.ts";

export interface QuestionSearchResult {
  request_id: string;
  question: string;
  answer: string;
  type: string;
  is_ai: boolean;
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

export const listPackages = () => koi.get<{ code: number; message: string; data: PackageItem[] }>("/api/v1/packages");

export const listMyPackages = () =>
  koi.get<{ code: number; message: string; data: PackageInstance[] }>("/api/v1/packages/my");

export const createOrder = (data: { packageId: number; provider?: string; couponCode?: string }) =>
  koi.post<{ code: number; message: string; data: { order: OrderItem; paymentUrl: string } }>("/api/v1/orders", data);

export const listMyOrders = () => koi.get<{ code: number; message: string; data: OrderItem[] }>("/api/v1/orders/my");

export const getApiKey = () => koi.get<{ code: number; message: string; data: ApiKeyView }>("/api/v1/api-key");

export const createApiKey = () =>
  koi.post<{ code: number; message: string; data: { key: string; info: ApiKeyView } }>("/api/v1/api-key");

export const getOcsConfig = () => koi.get<any>("/api/ocs/config", { key: "" });

export const listOcsSources = () => koi.get<{ code: number; message: string; data: any[] }>("/api/v1/admin/ocs/sources");
export const createOcsSource = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: any }>("/api/v1/admin/ocs/sources", data);
export const createAiProvider = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: { id: number } }>("/api/v1/admin/ai/providers", data);
export const createAiModel = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: { id: number } }>("/api/v1/admin/ai/models", data);
export const listAiModels = () => koi.get<{ code: number; message: string; data: any[] }>("/api/v1/admin/ai/models");
export const configurePaymentGateway = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: any }>("/api/v1/admin/payment/gateways", data);
export const createAdminPackage = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: PackageItem }>("/api/v1/admin/packages", data);
export const createCoupon = (data: Record<string, unknown>) =>
  koi.post<{ code: number; message: string; data: any }>("/api/v1/admin/coupons", data);

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
